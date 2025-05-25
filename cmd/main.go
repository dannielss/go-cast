package main

import (
	"go-cast/internal/chat"
	"go-cast/internal/handlers"
	"go-cast/internal/stream"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	tmpl, err := template.ParseGlob(filepath.Join("static", "*.html"))
	if err != nil {
		log.Fatal("Error parsing pages:", err)
	}
	handlers.SetTemplate(tmpl)

	chatHub := chat.NewHub()
	cs := handlers.NewChatHandler(chatHub)

	streamManager := stream.NewStreamManager()
	sh := handlers.NewStreamHandler(streamManager)

	// Page routes
	r.HandleFunc("/", handlers.HomePage)
	r.HandleFunc("/broadcaster", handlers.StreamPage)
	r.HandleFunc("/viewer", handlers.ViewerPage)

	// WebSocket handler
	r.HandleFunc("/ws/chat/{streamId}/{clientId}", cs.ChatHandler)
	r.HandleFunc("/ws/{streamId}/{role}/{clientId}", sh.StreamHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
