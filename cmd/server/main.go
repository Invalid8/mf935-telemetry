package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"mf935-telemetry/internal/auth"
	"mf935-telemetry/internal/client"
	"mf935-telemetry/internal/poller"
	"mf935-telemetry/internal/sse"
	"mf935-telemetry/internal/ws"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	rc := client.New()
	session := auth.New(rc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wsHub := ws.NewHub()
	sseHub := sse.NewHub()
	wsHub.SetSSE(sseHub)

	go wsHub.Run()

	p := poller.New(rc, session, wsHub)

	http.HandleFunc("/stream", wsHub.ServeWS)
	http.HandleFunc("/events", sseHub.ServeHTTP)
	http.HandleFunc("/api/login", loginHandler(session, p, ctx))
	serveStatic()

	if cfg.Password != "" {
		log.Println("server: config password found — authenticating")
		if err := session.Login(cfg.Password); err != nil {
			log.Fatalf("server: config login failed: %v", err)
		}
		log.Println("server: authenticated via config")
		go p.Run(ctx)
	} else {
		log.Println("server: no config password — waiting for web login")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		ip := lanIP()
		log.Printf("server: dashboard  → http://%s:9000", ip)
		log.Printf("server: ws stream  → ws://%s:9000/stream", ip)
		log.Printf("server: sse events → http://%s:9000/events", ip)
		if err := http.ListenAndServe(":9000", nil); err != nil {
			log.Fatalf("http server: %v", err)
		}
	}()

	<-quit
	log.Println("server: shutting down")
	cancel()
}

func lanIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}
