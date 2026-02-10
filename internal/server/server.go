package server

import (
	"fmt"
	"net/http"
)

// Démarre le serveur HTTP
func Start() {
	// Route racine "/"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Groupie Tracker server is running")
	})

	// Message dans le terminal
	fmt.Println("Server running on http://localhost:8080")

	// Démarrage du serveur
	http.ListenAndServe(":8080", nil)
}
