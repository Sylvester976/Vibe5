package main

import (
	"Vibe5/config"
	"Vibe5/routes"
	"log"
	"net/http"
)

func main() {
	// Load env vars
	config.LoadConfig()
	port := config.Port

	// Setup routes
	routes.SetupRoutes()
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	log.Println("Server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
