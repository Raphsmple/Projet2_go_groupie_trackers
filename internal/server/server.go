package server

import (
	"fmt"
	"net/http"
	"Projet2_go_groupie_trackers/internal/handlers"
)

// Démarre le serveur HTTP
func Start() {
	// Servir les fichiers statiques (CSS)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Routes
	http.HandleFunc("/", handlers.RootHandler)
	http.HandleFunc("/artists", handlers.ArtistsHandler)
	http.HandleFunc("/artist", handlers.ArtistHandler)

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
