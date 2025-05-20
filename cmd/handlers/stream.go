package handlers

import (
	"encoding/json"
	"go-cast/cmd/models"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type StreamRegistry struct {
	mu      sync.Mutex
	streams map[string]*models.StreamManager
}

func NewStreamRegistry() *StreamRegistry {
	return &StreamRegistry{
		streams: make(map[string]*models.StreamManager),
	}
}

func (sr *StreamRegistry) getOrCreate(streamID string) *models.StreamManager {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	stream, ok := sr.streams[streamID]
	if !ok {
		stream = models.NewStreamManager()
		sr.streams[streamID] = stream
	}
	return stream
}

func (sr *StreamRegistry) HandleSignal(w http.ResponseWriter, r *http.Request) {
	streamID := r.URL.Query().Get("stream")
	role := r.URL.Query().Get("role")   // "broadcaster" or "viewer"
	viewerID := r.URL.Query().Get("id") // required for viewers

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to websocket", http.StatusBadRequest)
		return
	}

	manager := sr.getOrCreate(streamID)

	switch role {
	case "broadcaster":
		manager.AddBroadcaster(conn)
		log.Printf("Broadcaster connected to stream %s", streamID)
	case "viewer":
		if viewerID == "" {
			conn.Close()
			http.Error(w, "Viewer ID required", http.StatusBadRequest)
			return
		}
		manager.AddViewer(viewerID, conn)
		log.Printf("Viewer %s connected to stream %s", viewerID, streamID)

		notify := []byte(`{"type":"viewer-joined", "id":"` + viewerID + `"}`)
		manager.SendToBroadcaster(notify)
	default:
		conn.Close()
		http.Error(w, "Missing or invalid role", http.StatusBadRequest)
		return
	}

	go handleMessages(manager, conn, role, viewerID)
}

func handleMessages(manager *models.StreamManager, conn *websocket.Conn, role, viewerID string) {
	defer func() {
		conn.Close()
		if role == "viewer" {
			manager.RemoveViewer(viewerID)
		}
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket closed:", err)
			break
		}

		log.Printf("Received from %s: %s", role, string(msg))

		if role == "broadcaster" {
			manager.Mu.Lock()
			manager.Offer = msg
			manager.Mu.Unlock()

			manager.BroadcastToViewers(msg)
		} else {
			manager.SendToBroadcaster(msg)
		}
	}
}

func (sr *StreamRegistry) GetAllStreams(w http.ResponseWriter, r *http.Request) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	streams := make([]*models.StreamManager, 0, len(sr.streams))
	for _, stream := range sr.streams {
		streams = append(streams, stream)
	}

	log.Printf("Active streams: %v", streams)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(streams)
}
