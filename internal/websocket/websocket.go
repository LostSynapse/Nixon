package websocket

import (
	"nixon/internal/logger"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins
	}
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan []byte)
	mutex     = &sync.RWMutex{}
)

// Handler upgrades HTTP to WebSocket and manages connection lifecycle.
func Handler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}
	defer ws.Close()

	// Register client
	mutex.Lock()
	clients[ws] = true
	mutex.Unlock()
	logger.Log.Info().Str("remote_addr", r.RemoteAddr).Msg("WebSocket client connected")

	// This loop is necessary to detect when a client disconnects.
	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			logger.Log.Info().Str("remote_addr", r.RemoteAddr).Msg("WebSocket client disconnected")
			mutex.Lock()
			delete(clients, ws)
			mutex.Unlock()
			break
		}
	}
}

// HandleMessages listens to the broadcast channel and forwards messages to clients.
// This must be started as a goroutine.
func HandleMessages() {
	for {
		msg := <-broadcast
		mutex.RLock()
		// Send message to all clients
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				logger.Log.Warn().Err(err).Str("remote_addr", client.RemoteAddr().String()).Msg("WebSocket write error, closing client")
				client.Close()
				// Safely remove the client inside the read-lock is tricky.
				// For simplicity, we let the read loop handle removal.
			}
		}
		mutex.RUnlock()
	}
}

// Broadcast sends a message to all connected WebSocket clients.
func Broadcast(message string) {
	broadcast <- []byte(message)
}
