package websocket

import (
	"log"
	"net/http"
	"nixon/internal/state"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan interface{})
	mutex     = &sync.Mutex{}
)

func HandleWebSocket(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()
	mutex.Lock()
	clients[ws] = true
	mutex.Unlock()

	s := state.Get()
	ws.WriteJSON(s)

	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			break
		}
	}

	mutex.Lock()
	delete(clients, ws)
	mutex.Unlock()
}

func broadcastState(s interface{}) {
	go func() {
		broadcast <- s
	}()
}

func HandleBroadcast() {
	for {
		msg := <-broadcast
		mutex.Lock()
		for client := range clients {
			if err := client.WriteJSON(msg); err != nil {
				log.Printf("WebSocket write error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}
}

func PollAndBroadcast() {
	var lastState state.AppStateStruct
	lastState = state.Get()

	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		currentState := state.Get()
		
		// Compare fields manually to avoid issues with comparing the mutex
		if currentState.SRTStreamActive != lastState.SRTStreamActive ||
			currentState.IcecastStreamActive != lastState.IcecastStreamActive ||
			currentState.RecordingActive != lastState.RecordingActive ||
			currentState.CurrentRecordingFile != lastState.CurrentRecordingFile ||
			currentState.DiskUsagePercent != lastState.DiskUsagePercent ||
			currentState.Listeners != lastState.Listeners ||
			currentState.ListenerPeak != lastState.ListenerPeak {
			
			broadcastState(currentState)
			lastState = currentState
		}
	}
}

