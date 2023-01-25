package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
)

var (
	stravaClientId     = os.Getenv("STRAVA_CLIENT_ID")
	stravaClientSecret = os.Getenv("STRAVA_CLIENT_SECRET")
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	tmpl_files := []string{
		"templates/base.html",
		"templates/login.html",
	}
	tmpl := template.Must(template.ParseFiles(tmpl_files...))
	err := tmpl.Execute(w, nil)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/", indexHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
