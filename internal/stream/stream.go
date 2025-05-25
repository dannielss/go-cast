package stream

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type StreamManager struct {
	streams map[string]*Stream
	mu      sync.Mutex
}

type Stream struct {
	Broadcaster *websocket.Conn
	Viewers     map[string]*websocket.Conn
}

func NewStreamManager() *StreamManager {
	return &StreamManager{
		streams: make(map[string]*Stream),
	}
}

func (m *StreamManager) RegisterBroadcaster(streamId string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.streams[streamId] = &Stream{
		Broadcaster: conn,
		Viewers:     make(map[string]*websocket.Conn),
	}
	log.Printf("Broadcaster registered: %s\n", streamId)
}

func (m *StreamManager) RegisterViewer(streamId, clientId string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	stream, ok := m.streams[streamId]
	if !ok {
		log.Printf("No broadcaster for stream %s\n", streamId)
		conn.Close()
		return
	}
	stream.Viewers[clientId] = conn
	log.Printf("Viewer %s joined stream %s\n", clientId, streamId)
	log.Printf("Total viewers: %d\n", len(stream.Viewers))
}

func (m *StreamManager) Unregister(streamId, role, clientId string) {
	m.mu.Lock()
	stream, ok := m.streams[streamId]
	if !ok {
		m.mu.Unlock()
		return
	}

	if role == "viewer" {
		if conn, ok := stream.Viewers[clientId]; ok {
			conn.Close()
			delete(stream.Viewers, clientId)
		}
		m.mu.Unlock()
		return
	}

	if role == "broadcaster" {
		for clientId, viewerConn := range stream.Viewers {
			err := viewerConn.WriteJSON(map[string]interface{}{
				"type": "stream-closed",
				"msg":  "Stream has ended by broadcaster",
			})
			if err != nil {
				log.Printf("Failed to send stream-closed to viewer %s: %v", clientId, err)
			}
			viewerConn.Close()
		}

		if stream.Broadcaster != nil {
			stream.Broadcaster.Close()
		}

		delete(m.streams, streamId)
	}

	m.mu.Unlock()
}

func (m *StreamManager) RouteMessage(streamId, role, clientId string, msg []byte) {
	m.mu.Lock()
	stream, ok := m.streams[streamId]
	if !ok {
		m.mu.Unlock()
		return
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(msg, &payload); err != nil {
		m.mu.Unlock()
		log.Println("Invalid message format")
		return
	}

	payload["from"] = clientId

	var broadcasterConn *websocket.Conn
	var viewerConn *websocket.Conn
	broadcasterConn = stream.Broadcaster

	to, toExists := payload["to"].(string)
	if toExists {
		viewerConn = stream.Viewers[to]
	}
	m.mu.Unlock()

	switch role {
	case "viewer":
		if broadcasterConn != nil {
			if err := broadcasterConn.WriteJSON(payload); err != nil {
				log.Printf("Failed to send message to broadcaster: %v", err)
			}
		}
	case "broadcaster":
		if viewerConn != nil {
			if err := viewerConn.WriteJSON(payload); err != nil {
				log.Printf("Failed to send message to viewer: %v", err)
			}
		}
	}
}
