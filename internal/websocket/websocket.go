// internal/websocket/websocket.go
package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for simplicity.
		// In production, you'd restrict this.
		return true
	},
}

// client represents a single websocket connection.
type client struct {
	conn *websocket.Conn
	send chan []byte
}

// hub maintains the set of active clients and broadcasts messages.
type hub struct {
	clients    map[*client]bool
	broadcast  chan []byte
	register   chan *client
	unregister chan *client
	mu         sync.RWMutex
}

var h = &hub{
	broadcast:  make(chan []byte),
	register:   make(chan *client),
	unregister: make(chan *client),
	clients:    make(map[*client]bool),
}

// StartBroadcaster runs the central hub logic
func (h *hub) StartBroadcaster() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.clients[c] = true
			h.mu.Unlock()
			log.Printf("Client connected. Total clients: %d", len(h.clients))
			// Send initial full status on connect?
			// status := api.GetFullStatus() // This creates a circular dependency
			// h.broadcastStatus(status) // Need to rethink this
		case c := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
			}
			h.mu.Unlock()
			log.Printf("Client disconnected. Total clients: %d", len(h.clients))
		case message := <-h.broadcast:
			h.mu.RLock()
			for c := range h.clients {
				select {
				case c.send <- message:
				default:
					// Failed to send, assume client is dead
					close(c.send)
					delete(h.clients, c)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// writer pumps messages from the hub to the websocket connection.
func (c *client) writer() {
	defer func() {
		c.conn.Close()
	}()
	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Error writing to client: %v", err)
			return
		}
	}
}

// reader pumps messages from the websocket connection to the hub.
// For now, it just handles disconnects.
func (c *client) reader() {
	defer func() {
		h.unregister <- c
		c.conn.Close()
	}()
	// Set read deadline, etc. if needed
	c.conn.SetReadLimit(512)
	// c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	// c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); return nil })
	for {
		// Read message from browser
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Client read error: %v", err)
			}
			break
		}
		// We don't expect messages from client yet
	}
}

// HandleConnections handles websocket requests from the peer.
func HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade websocket: %v", err)
		return
	}
	c := &client{conn: conn, send: make(chan []byte, 256)}
	h.register <- c

	go c.writer()
	go c.reader()
}

// BroadcastStatus is called by the status updater to send new status to all clients
func BroadcastStatus(status interface{}) {
	message, err := json.Marshal(status)
	if err != nil {
		log.Printf("Error marshalling status: %v", err)
		return
	}
	h.broadcast <- message
}

// GetClientCount returns the number of active websocket clients
func GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// StartBroadcaster is the public function to start the hub's run loop
func StartBroadcaster() {
	h.StartBroadcaster()
}

