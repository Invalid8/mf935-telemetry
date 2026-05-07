package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/invalid8/mf935-telemetry/internal/auth"
	"github.com/invalid8/mf935-telemetry/internal/poller"
)

func serveStatic() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
}

type loginRequest struct {
	Password string `json:"password"` // SHA256(plaintext).toUpperCase() — pre-hashed by browser
}

type loginResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

func loginHandler(session *auth.Session, p *poller.Poller, ctx context.Context) http.HandlerFunc {
	var started bool

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(loginResponse{OK: false, Message: "invalid request"})
			return
		}

		if err := session.Login(req.Password); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(loginResponse{OK: false, Message: err.Error()})
			return
		}

		// Set password cookie — expires in 30 days
		http.SetCookie(w, &http.Cookie{
			Name:     "mf_pw",
			Value:    req.Password,
			Path:     "/",
			Expires:  time.Now().Add(30 * 24 * time.Hour),
			HttpOnly: false, // must be readable by JS for re-auth on reconnect
			SameSite: http.SameSiteStrictMode,
		})

		// Start poller only once
		if !started {
			started = true
			go p.Run(ctx)
		}

		json.NewEncoder(w).Encode(loginResponse{OK: true})
	}
}
