package endpoints

import (
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sort"
	"time"

	"../config"
	"../models"
)

// Index creates a request handler that serves index.html, the main scoreboard page with all of
// the teams and their scores.
func Index(db *sql.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		funcs := template.FuncMap{
			"increment": func(a int) string { return fmt.Sprintf("%d", a+1) },
		}
		templateFile, err := os.Open(path.Join(cfg.HTMLDir, "index.html"))
		if err != nil {
			fmt.Println("Error parsing template", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("We done goofed! Try again in a few minutes."))
			return
		}
		templateCode, readErr := ioutil.ReadAll(templateFile)
		if readErr != nil {
			fmt.Println("Error reading template for index.html", readErr)
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("We done goofed! Try again in a few minutes."))
			return
		}
		teamInfo, err := models.FindTeams(db)
		if err != nil {
			fmt.Println("Error finding teams", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("We done goofed! Try again in a few minutes."))
			return
		}
		data := struct {
			CTF   string
			Teams []models.TeamScore
		}{
			cfg.CTFName,
			[]models.TeamScore{},
		}
		fmt.Println("Got teams", teamInfo)
		sort.Sort(models.TeamByScore(teamInfo))
		for _, team := range teamInfo {
			lastSubmitted := "No flags submitted yet"
			if team.LastSubmission.After(time.Date(2015, time.January, 1, 1, 0, 0, 0, time.UTC)) {
				lastSubmitted = team.LastSubmission.Local().Format(time.UnixDate)
			}
			data.Teams = append(data.Teams, models.TeamScore{
				team.Name,
				team.Members,
				team.Score,
				lastSubmitted,
			})
		}
		t, err := template.New("index").Funcs(funcs).Parse(string(templateCode))
		if err != nil {
			fmt.Println("Error building template", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("We done goofed! Try again in a few minutes."))
			return
		}
		t.Execute(w, data)
	}
}
