package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"mf935-telemetry/internal/events"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type client struct {
	conn      *websocket.Conn
	send      chan []byte
	closeOnce sync.Once
}

func (c *client) close() {
	c.closeOnce.Do(func() {
		close(c.send)
		c.conn.Close()
	})
}

type SSEPublisher interface {
	Publish(event events.Event)
}

type Hub struct {
	mu          sync.RWMutex
	clients     map[*client]bool
	broadcast   chan []byte
	register    chan *client
	unregister  chan *client
	snapshot    []byte
	unreachable []byte
	sse         SSEPublisher
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *client),
		unregister: make(chan *client),
	}
}

func (h *Hub) SetSSE(p SSEPublisher) {
	h.sse = p
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = true
			if h.unreachable != nil {
				select {
				case c.send <- h.snapshot:
				default:
				}
				select {
				case c.send <- h.unreachable:
				default:
				}
			} else if h.snapshot != nil {
				select {
				case c.send <- h.snapshot:
				default:
				}
			}
			log.Printf("ws: client connected — total: %d", len(h.clients))

		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				c.close()
				log.Printf("ws: client disconnected — total: %d", len(h.clients))
			}

		case msg := <-h.broadcast:
			for c := range h.clients {
				select {
				case c.send <- msg:
				default:
					delete(h.clients, c)
					c.close()
				}
			}
		}
	}
}

func (h *Hub) Broadcast(event events.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("ws: marshal error: %v", err)
		return
	}

	switch event.Type {
	case events.EventDeviceUnreachable, events.EventDeviceMismatch:
		h.unreachable = data
	case events.EventDeviceRecovered:
		h.unreachable = nil
	}

	h.broadcast <- data

	if h.sse != nil {
		h.sse.Publish(event)
	}
}

func (h *Hub) StoreSnapshot(event events.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("ws: snapshot marshal error: %v", err)
		return
	}
	h.snapshot = data
	h.unreachable = nil
	h.broadcast <- data
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws: upgrade error: %v", err)
		return
	}

	c := &client{
		conn: conn,
		send: make(chan []byte, 256),
	}

	h.register <- c

	// writer
	go func() {
		defer func() { h.unregister <- c }()
		for msg := range c.send {
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}()

	go func() {
		defer func() { h.unregister <- c }()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}
