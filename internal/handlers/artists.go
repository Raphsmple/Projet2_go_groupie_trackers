package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Struct API principale
type API struct {
	Artists   string `json:"artists"`
	Locations string `json:"locations"`
	Dates     string `json:"dates"`
	Relation  string `json:"relation"`
}

// Struct artiste
type Artist struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	Image        string   `json:"image"`
	FirstAlbum   string   `json:"firstAlbum"`
	CreationDate int      `json:"creationDate"`
}

// Struct relation concerts
type RelationData struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

type RelationResponse struct {
	Index []RelationData `json:"index"`
}

// Struct pour passer les données à artist.html
type ArtistPage struct {
	Artist   Artist
	Concerts RelationData
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/artists", http.StatusFound)
}

func ArtistsHandler(w http.ResponseWriter, r *http.Request) {
	// 1) API principale
	respAPI, err := http.Get("https://groupietrackers.herokuapp.com/api")
	if err != nil {
		http.Error(w, "Erreur API: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer respAPI.Body.Close()

	bodyAPI, err := io.ReadAll(respAPI.Body)
	if err != nil {
		http.Error(w, "Lecture API: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var api API
	if err := json.Unmarshal(bodyAPI, &api); err != nil {
		http.Error(w, "Parse API: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2) Liste artistes
	respArtists, err := http.Get(api.Artists)
	if err != nil {
		http.Error(w, "Erreur artistes: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer respArtists.Body.Close()

	bodyArtists, err := io.ReadAll(respArtists.Body)
	if err != nil {
		http.Error(w, "Lecture artistes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var artists []Artist
	if err := json.Unmarshal(bodyArtists, &artists); err != nil {
		http.Error(w, "Parse artistes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3) Filtrage
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	yearMin, _ := strconv.Atoi(r.URL.Query().Get("year_min"))
	yearMax, _ := strconv.Atoi(r.URL.Query().Get("year_max"))
	membersMin, _ := strconv.Atoi(r.URL.Query().Get("members_min"))

	var filtered []Artist
	for _, a := range artists {
		match := true
		if q != "" {
			match = strings.Contains(strings.ToLower(a.Name), q)
			if !match {
				for _, m := range a.Members {
					if strings.Contains(strings.ToLower(m), q) {
						match = true
						break
					}
				}
			}
		}
		if !match {
			continue
		}
		if yearMin != 0 && a.CreationDate < yearMin {
			continue
		}
		if yearMax != 0 && a.CreationDate > yearMax {
			continue
		}
		if membersMin != 0 && len(a.Members) < membersMin {
			continue
		}
		filtered = append(filtered, a)
	}

	// 4) Template
	tmpl, err := template.ParseFiles("templates/artiste.html")
	if err != nil {
		http.Error(w, "Template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Artists    []Artist
		Query      string
		YearMin    int
		YearMax    int
		MembersMin int
	}{
		Artists:    filtered,
		Query:      q,
		YearMin:    yearMin,
		YearMax:    yearMax,
		MembersMin: membersMin,
	}

	tmpl.Execute(w, data)
}

func ArtistHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID manquant", http.StatusBadRequest)
		return
	}

	// API principale
	respAPI, err := http.Get("https://groupietrackers.herokuapp.com/api")
	if err != nil {
		http.Error(w, "Erreur API: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer respAPI.Body.Close()

	bodyAPI, err := io.ReadAll(respAPI.Body)
	if err != nil {
		http.Error(w, "Lecture API: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var api API
	if err := json.Unmarshal(bodyAPI, &api); err != nil {
		http.Error(w, "Parse API: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Liste artistes
	respArtists, err := http.Get(api.Artists)
	if err != nil {
		http.Error(w, "Erreur artistes: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer respArtists.Body.Close()

	bodyArtists, err := io.ReadAll(respArtists.Body)
	if err != nil {
		http.Error(w, "Lecture artistes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var artists []Artist
	if err := json.Unmarshal(bodyArtists, &artists); err != nil {
		http.Error(w, "Parse artistes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Chercher l'artiste
	var selected Artist
	found := false
	for _, a := range artists {
		if fmt.Sprint(a.ID) == id {
			selected = a
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "Artiste introuvable", http.StatusNotFound)
		return
	}

	// Relations concerts
	respRel, err := http.Get(api.Relation)
	if err != nil {
		http.Error(w, "Erreur relations: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer respRel.Body.Close()

	bodyRel, err := io.ReadAll(respRel.Body)
	if err != nil {
		http.Error(w, "Lecture relations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var relations RelationResponse
	if err := json.Unmarshal(bodyRel, &relations); err != nil {
		http.Error(w, "Parse relations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var rel RelationData
	for _, r := range relations.Index {
		if r.ID == selected.ID {
			rel = r
			break
		}
	}

	// Template
	tmpl, err := template.ParseFiles("templates/present/artist.html")
	if err != nil {
		http.Error(w, "Template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := ArtistPage{
		Artist:   selected,
		Concerts: rel,
	}
	tmpl.Execute(w, data)
}
