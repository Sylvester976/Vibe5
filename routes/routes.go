package routes

import (
	"fmt"
	"net/http"
)

func SetupRoutes() {
	// Landing page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Welcome to Vibe5 🎶 - Your custom Spotify Wrapped")
	})

	// Login (Spotify OAuth)
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Login route (to be implemented)")
	})

	// Callback (Spotify redirects here after login)
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Callback route (to be implemented)")
	})

	// Wrapped page (show top 5 songs + artists)
	http.HandleFunc("/vibe5", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "This will show your top 5 songs + artists")
	})
}
