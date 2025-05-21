package handlers

import (
	"fmt"
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

var manager = NewStreamManager()

func WSHandler(w http.ResponseWriter, r *http.Request) {
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
		manager.RegisterBroadcaster(streamId, conn)
	} else if role == "viewer" {
		manager.RegisterViewer(streamId, clientId, conn)
	}

	defer manager.Unregister(streamId, role, clientId)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("%s %s disconnected: %v\n", role, clientId, err)
			break
		}
		manager.RouteMessage(streamId, role, clientId, msg)
	}
}
