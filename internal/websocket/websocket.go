// internal/websocket/websocket.go
package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	// "nixon/internal/gstreamer" // A3 Fix: Removed direct import
	"sync"
	// "time" // Removed, polling is no longer done here

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins
	}
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan []byte) // Channel carries marshalled JSON bytes
	mutex     sync.RWMutex      // Use RWMutex
)

// HandleConnections upgrades HTTP to WebSocket and manages connection lifecycle.
func HandleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer ws.Close()

	// Register client
	mutex.Lock()
	clients[ws] = true
	mutex.Unlock()
	log.Println("WebSocket client connected:", ws.RemoteAddr())

	// Send initial state upon connection (Requires GStreamer Manager to be ready)
	// This might fail if called before GStreamer Init completes fully.
	// Consider adding a slight delay or a ready check.
	/* // Removed initial state send here, rely on first broadcast/poll after connect
	if manager := gstreamer.GetManager(); manager != nil { // Check if manager exists
		initialStatus := manager.GetStatus() // A3: Get status from GStreamer Manager
		initialPayload, err := json.Marshal(initialStatus)
		if err == nil {
			if writeErr := ws.WriteMessage(websocket.TextMessage, initialPayload); writeErr != nil {
				log.Printf("Error sending initial state to %s: %v", ws.RemoteAddr(), writeErr)
			}
		} else {
			log.Printf("Error marshalling initial state: %v", err)
		}
	} else {
		log.Println("Warning: Cannot send initial state, GStreamer manager not ready.")
	}
	*/

	// Read loop (keep connection alive, handle client messages if any)
	for {
		// ReadMessage blocks until a message is received or an error occurs
		_, _, readErr := ws.ReadMessage()
		if readErr != nil {
			// Log specific close errors if needed (e.g., websocket.IsCloseError)
			log.Printf("WebSocket client read error/disconnect: %v", readErr)
			break // Exit loop on disconnect/error
		}
		// Handle client messages here if the protocol requires it
		// log.Printf("Received message from client: %s", message)
	}

	// Unregister client
	mutex.Lock()
	delete(clients, ws)
	mutex.Unlock()
	log.Println("WebSocket client disconnected:", ws.RemoteAddr())
}

// BroadcastUpdate is called by external packages (like gstreamer via callback) (A3 Fix)
// It accepts the status map, marshals it, and sends to the broadcast channel.
func BroadcastUpdate(status map[string]interface{}) {
	payload, err := json.Marshal(status)
	if err != nil {
		log.Printf("Error marshalling status for broadcast: %v", err)
		return
	}
	// Use non-blocking send to prevent blocking the caller (e.g., GStreamer bus handler)
	select {
	case broadcast <- payload:
		// log.Println("Status update queued for broadcast.") // Can be noisy
	default:
		log.Println("Warning: WebSocket broadcast channel full, dropping status update.")
	}
}

// StartBroadcaster listens on the broadcast channel and sends messages to all connected clients.
func StartBroadcaster() {
	log.Println("Starting WebSocket broadcaster...")
	for {
		payload := <-broadcast // Wait for a message to broadcast

		mutex.RLock() // Lock for reading client list
		// Copy client pointers to avoid holding lock during writes
		currentClients := make([]*websocket.Conn, 0, len(clients))
		for client := range clients {
			currentClients = append(currentClients, client)
		}
		mutex.RUnlock()

		// log.Printf("Broadcasting update to %d clients", len(currentClients)) // Can be noisy

		// Iterate over the copied list
		for _, client := range currentClients {
			// Consider setting a write deadline
			// client.SetWriteDeadline(time.Now().Add(5 * time.Second))
			err := client.WriteMessage(websocket.TextMessage, payload)
			// client.SetWriteDeadline(time.Time{}) // Clear deadline

			if err != nil {
				log.Printf("WebSocket write error to %s: %v. Removing client.", client.RemoteAddr(), err)
				client.Close() // Ensure connection is closed
				// Remove the failed client from the main map (requires full Lock)
				mutex.Lock()
				delete(clients, client)
				mutex.Unlock()
			}
		}
	}
}

// PollAndBroadcast is NO LONGER NEEDED (A3 Fix)
// GStreamer manager now pushes updates via the callback.
/*
func PollAndBroadcast() {
	// ... implementation removed ...
}
*/

