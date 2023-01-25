package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
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

func authorizeHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse("https://www.strava.com/oauth/authorize")
	if err != nil {
		log.Fatal(err)
	}
	q := u.Query()
	q.Set("client_id", stravaClientId)
	q.Set("redirect_uri", fmt.Sprintf("http://%s/strava_callback", r.Host))
	q.Set("response_type", "code")
	q.Set("scope", "activity:read_all")
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "strava redirect callback handler.")
}

func main() {
	http.HandleFunc("/strava_callback", callbackHandler)
	http.HandleFunc("/strava_authorize", authorizeHandler)
	http.HandleFunc("/", indexHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
