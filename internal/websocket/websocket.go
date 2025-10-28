package websocket

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
	"nixon/internal/common"
	"nixon/internal/slogger"
	"sync"
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
		slogger.Log.Error("Failed to upgrade WebSocket connection", "err", err)
		return
	}
	defer ws.Close()

	// Register client
	mutex.Lock()
	clients[ws] = true
	mutex.Unlock()
	slogger.Log.Info("WebSocket client connected", "remote_addr", r.RemoteAddr)

	// This loop is necessary to detect when a client disconnects.
	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			slogger.Log.Info("WebSocket client disconnected", "remote_addr", r.RemoteAddr)
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
				slogger.Log.Warn("WebSocket write error, closing client", "err", err, "remote_addr", client.RemoteAddr().String())
				client.Close()
				// Safely remove the client inside the read-lock is tricky.
				// For simplicity, we let the read loop handle removal.
			}
		}
		mutex.RUnlock()
	}
}

// BroadcastStatus sends the current AudioStatus to all connected clients.
func BroadcastStatus(status common.AudioStatus) {
	// Create a message envelope.
	msg := struct {
		Type    string             `json:"type"`
		Payload common.AudioStatus `json:"payload"`
	}{
		Type:    "status_update",
		Payload: status,
	}

	payloadBytes, err := json.Marshal(msg)
	if err != nil {
		slogger.Log.Error("Failed to marshal status for broadcast", "err", err)
		return
	}

	// Send the marshaled message to the broadcast channel.
	broadcast <- payloadBytes
}

// Broadcast sends a message to all connected WebSocket clients.
func Broadcast(message string) {
	broadcast <- []byte(message)
}
