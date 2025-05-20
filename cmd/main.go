package main

import (
	"fmt"
	"go-cast/cmd/handlers"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

func main() {
	tmpl, err := template.ParseGlob(filepath.Join("templates", "*.html"))
	if err != nil {
		log.Fatal("Error parsing templates:", err)
	}
	handlers.SetTemplate(tmpl)

	router := mux.NewRouter()

	// Page routes
	router.HandleFunc("/", handlers.HomePage)
	router.HandleFunc("/stream", handlers.StreamPage)
	router.HandleFunc("/viewer/{id}", handlers.ViewerPage)

	streamRegistry := handlers.NewStreamRegistry()
	router.HandleFunc("/ws/signal", streamRegistry.HandleSignal)

	// rest
	router.HandleFunc("/api/streams", streamRegistry.GetAllStreams).Methods("GET")

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
