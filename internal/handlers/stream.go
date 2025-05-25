package handlers

import (
	"fmt"
	"go-cast/internal/stream"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Stream struct {
	Manager *stream.StreamManager
}

func NewStreamHandler(streamManager *stream.StreamManager) *Stream {
	return &Stream{Manager: streamManager}
}

func (s *Stream) StreamHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamId := vars["streamId"]
	role := vars["role"]
	clientId := vars["clientId"]
	fmt.Printf("WebSocket connection: streamId=%s, role=%s, clientId=%s\n", streamId, role, clientId)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	// Register connection
	if role == "broadcaster" {
		s.Manager.RegisterBroadcaster(streamId, conn)
	} else if role == "viewer" {
		s.Manager.RegisterViewer(streamId, clientId, conn)
	}

	defer s.Manager.Unregister(streamId, role, clientId)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("%s %s disconnected: %v\n", role, clientId, err)
			break
		}
		s.Manager.RouteMessage(streamId, role, clientId, msg)
	}
}
