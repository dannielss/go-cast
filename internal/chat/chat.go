package chat

import (
	"log"
	"sync"
)

type Client struct {
	ID   string
	Send chan []byte
}

type Hub struct {
	mu    sync.Mutex
	rooms map[string]map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Join(streamId string, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[streamId] == nil {
		h.rooms[streamId] = make(map[*Client]bool)
	}

	h.rooms[streamId][client] = true
}

func (h *Hub) Leave(streamId string, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.rooms[streamId], client)
	close(client.Send)
}

func (h *Hub) Broadcast(streamId string, message []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	log.Printf("Broadcasting to streamId=%s: %s", streamId, message)
	for client := range h.rooms[streamId] {
		select {
		case client.Send <- message:
		default:
			log.Printf("Client %s channel full or dead, removing", client.ID)
			delete(h.rooms[streamId], client)
			close(client.Send)
		}
	}
}
