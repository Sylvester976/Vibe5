package routes

import (
	"Vibe5/handlers"
	"net/http"
)

func SetupRoutes() {
	http.HandleFunc("/", handlers.IndexHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/callback", handlers.CallbackHandler)
	http.HandleFunc("/vibe", handlers.VibeHandler)
}
