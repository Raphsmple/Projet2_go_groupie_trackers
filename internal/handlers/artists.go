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

// Struct pour un artiste
type Artist struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	Image        string   `json:"image"`
	FirstAlbum   string   `json:"firstAlbum"`
	CreationDate int      `json:"creationDate"`
}

// Struct pour les concerts / relations
type RelationData struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

type RelationResponse struct {
	Index []RelationData `json:"index"`
}

// ==================== HANDLER LISTE DES ARTISTES ====================
func ArtistsHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	if err != nil {
		http.Error(w, "Erreur API: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Erreur lecture API: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var artists []Artist
	if err := json.Unmarshal(body, &artists); err != nil {
		http.Error(w, "Erreur parse JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

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

	tmpl, err := template.ParseFiles("templates/artiste.html")
	if err != nil {
		http.Error(w, "Erreur template: "+err.Error(), http.StatusInternalServerError)
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

// ==================== HANDLER ARTISTE INDIVIDUEL ====================
func ArtistHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID manquant", http.StatusBadRequest)
		return
	}

	resp, _ := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	var artists []Artist
	json.Unmarshal(body, &artists)

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

	respRel, _ := http.Get("https://groupietrackers.herokuapp.com/api/relation")
	bodyRel, _ := io.ReadAll(respRel.Body)
	defer respRel.Body.Close()
	var relations RelationResponse
	json.Unmarshal(bodyRel, &relations)

	var rel RelationData
	for _, r := range relations.Index {
		if r.ID == selected.ID {
			rel = r
			break
		}
	}

	tmpl, _ := template.ParseFiles("templates/present/artist.html")

	data := struct {
		Artist   Artist
		Concerts RelationData
	}{
		Artist:   selected,
		Concerts: rel,
	}

	tmpl.Execute(w, data)
}
