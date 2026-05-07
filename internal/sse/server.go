package sse

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"mf935-telemetry/internal/events"
)

type client struct {
	ch chan []byte
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*client]bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*client]bool),
	}
}

func (h *Hub) Publish(event events.Event) {
	if !events.NotificationEvents[event.Type] {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("sse: marshal error: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for c := range h.clients {
		select {
		case c.ch <- data:
		default:
		}
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	c := &client{ch: make(chan []byte, 32)}

	h.mu.Lock()
	h.clients[c] = true
	h.mu.Unlock()

	log.Printf("sse: client connected — total: %d", h.count())

	defer func() {
		h.mu.Lock()
		delete(h.clients, c)
		h.mu.Unlock()
		log.Printf("sse: client disconnected — total: %d", h.count())
	}()

	for {
		select {
		case <-r.Context().Done():
			return
		case data := <-c.ch:
			_, err := w.Write(append(append([]byte("data: "), data...), '\n', '\n'))
			if err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (h *Hub) count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
