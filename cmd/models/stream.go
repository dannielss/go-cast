package models

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type StreamManager struct {
	ID          string
	Broadcaster *websocket.Conn
	Viewers     map[string]*websocket.Conn
	Offer       []byte
	Mu          sync.Mutex
}

func NewStreamManager() *StreamManager {
	return &StreamManager{
		ID:      uuid.NewString(),
		Viewers: make(map[string]*websocket.Conn),
	}
}

func (sm *StreamManager) AddBroadcaster(conn *websocket.Conn) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()
	sm.Broadcaster = conn
}

func (sm *StreamManager) AddViewer(id string, conn *websocket.Conn) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()
	sm.Viewers[id] = conn

	if sm.Offer != nil {
		conn.WriteMessage(websocket.TextMessage, sm.Offer)
	}
}

func (sm *StreamManager) RemoveViewer(id string) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()
	delete(sm.Viewers, id)
}

func (sm *StreamManager) BroadcastToViewers(msg []byte) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()
	for _, conn := range sm.Viewers {
		conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func (sm *StreamManager) SendToBroadcaster(msg []byte) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()
	if sm.Broadcaster != nil {
		sm.Broadcaster.WriteMessage(websocket.TextMessage, msg)
	}
}
