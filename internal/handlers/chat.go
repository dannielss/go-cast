package handlers

import (
	"go-cast/internal/chat"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type ChatServer struct {
	Hub *chat.Hub
}

func NewChatHandler(hub *chat.Hub) *ChatServer {
	return &ChatServer{
		Hub: hub,
	}
}

func (cs *ChatServer) ChatHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamId := vars["streamId"]
	clientId := vars["clientid"]

	log.Printf("ChatHandler called for streamId=%s clientId=%s", streamId, clientId)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	client := &chat.Client{
		ID:   clientId,
		Send: make(chan []byte, 256),
	}

	cs.Hub.Join(streamId, client)

	go func() {
		defer func() {
			cs.Hub.Leave(streamId, client)
			conn.Close()
		}()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("ReadMessage error: %v", err)
				break
			}

			cs.Hub.Broadcast(streamId, message)
		}
	}()

	go func() {
		for msg := range client.Send {
			log.Printf("Sending message to client %s: %s", clientId, string(msg))
			err := conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Printf("WriteMessage error: %v", err)
				break
			}
		}
	}()
}
