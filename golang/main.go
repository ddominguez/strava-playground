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

func indexHandler(response http.ResponseWriter, request *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/base.html"))
	response.Header().Set("Content-Type", "text/html")
	err := tmpl.Execute(response, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/", indexHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
