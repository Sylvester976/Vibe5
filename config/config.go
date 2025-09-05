package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	ClientID     string
	ClientSecret string
	RedirectURI  string
)

func LoadConfig() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	ClientID = os.Getenv("SPOTIFY_CLIENT_ID")
	ClientSecret = os.Getenv("SPOTIFY_CLIENT_SECRET")
	RedirectURI = os.Getenv("SPOTIFY_REDIRECT_URI")

	if ClientID == "" || ClientSecret == "" || RedirectURI == "" {
		log.Fatal("Missing required Spotify environment variables")
	}
}
