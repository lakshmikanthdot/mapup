package realtime

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"mapup/backend/internal/models"
)

// client wraps a websocket connection with its own write mutex.
// gorilla/websocket only permits a single concurrent writer per connection,
// so serialising writes here keeps concurrent Broadcast calls from
// interleaving frames and corrupting the stream.
type client struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]*client
}

func NewHub() *Hub { return &Hub{clients: make(map[*websocket.Conn]*client)} }

func (h *Hub) Add(c *websocket.Conn) {
	h.mu.Lock()
	h.clients[c] = &client{conn: c}
	h.mu.Unlock()
}

func (h *Hub) Remove(c *websocket.Conn) {
	h.mu.Lock()
	cl := h.clients[c]
	delete(h.clients, c)
	h.mu.Unlock()
	if cl != nil {
		_ = cl.conn.Close()
	}
}

func (h *Hub) Broadcast(a models.Alert) {
	payload, err := json.Marshal(a)
	if err != nil {
		return
	}

	h.mu.RLock()
	clients := make([]*client, 0, len(h.clients))
	for _, cl := range h.clients {
		clients = append(clients, cl)
	}
	h.mu.RUnlock()

	for _, cl := range clients {
		cl.mu.Lock()
		err := cl.conn.WriteMessage(websocket.TextMessage, payload)
		cl.mu.Unlock()
		if err != nil {
			h.Remove(cl.conn)
		}
	}
}
