package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"sort"
)

type Track struct {
	Name   string
	Artist string
	Album  string
}

type Artist struct {
	Name   string
	Genres []string
}

type VibeData struct {
	TopTracks  []Track
	TopArtists []Artist
	TopAlbums  []string
	TopGenres  []string
}

func VibeHandler(w http.ResponseWriter, r *http.Request) {
	if userAccessToken == "" {
		log.Println("⚠️ No access token found, redirecting to /")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	client := &http.Client{}
	vibe := VibeData{}

	// ----- Fetch Top Tracks -----
	reqTracks, _ := http.NewRequest("GET", "https://api.spotify.com/v1/me/top/tracks?limit=5&time_range=medium_term", nil)
	reqTracks.Header.Set("Authorization", "Bearer "+userAccessToken)
	respTracks, _ := client.Do(reqTracks)
	defer respTracks.Body.Close()

	var tracksResp struct {
		Items []struct {
			Name  string `json:"name"`
			Album struct {
				Name string `json:"name"`
			} `json:"album"`
			Artists []struct {
				Name string `json:"name"`
			} `json:"artists"`
		} `json:"items"`
	}
	json.NewDecoder(respTracks.Body).Decode(&tracksResp)

	for _, t := range tracksResp.Items {
		vibe.TopTracks = append(vibe.TopTracks, Track{
			Name:   t.Name,
			Artist: t.Artists[0].Name,
			Album:  t.Album.Name,
		})
		vibe.TopAlbums = append(vibe.TopAlbums, t.Album.Name)
	}

	// ----- Fetch Top Artists -----
	reqArtists, _ := http.NewRequest("GET", "https://api.spotify.com/v1/me/top/artists?limit=5&time_range=medium_term", nil)
	reqArtists.Header.Set("Authorization", "Bearer "+userAccessToken)
	respArtists, _ := client.Do(reqArtists)
	defer respArtists.Body.Close()

	var artistsResp struct {
		Items []struct {
			Name   string   `json:"name"`
			Genres []string `json:"genres"`
		} `json:"items"`
	}
	json.NewDecoder(respArtists.Body).Decode(&artistsResp)

	genreCount := map[string]int{}
	for _, a := range artistsResp.Items {
		vibe.TopArtists = append(vibe.TopArtists, Artist{
			Name:   a.Name,
			Genres: a.Genres,
		})
		for _, g := range a.Genres {
			genreCount[g]++
		}
	}

	// ----- Rank Top Genres -----
	type kv struct {
		Key   string
		Value int
	}
	var sortedGenres []kv
	for k, v := range genreCount {
		sortedGenres = append(sortedGenres, kv{k, v})
	}
	sort.Slice(sortedGenres, func(i, j int) bool {
		return sortedGenres[i].Value > sortedGenres[j].Value
	})

	for i := 0; i < 5 && i < len(sortedGenres); i++ {
		vibe.TopGenres = append(vibe.TopGenres, sortedGenres[i].Key)
	}

	// ----- Pass to template -----
	jsonData, err := json.Marshal(vibe) // convert Go struct to JSON
	if err != nil {
		http.Error(w, "Failed to marshal data", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/vibe.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	// Pass JSON string instead of struct
	tmpl.Execute(w, string(jsonData))

}
