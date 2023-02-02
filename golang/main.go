package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

var (
	stravaClientId     = os.Getenv("STRAVA_CLIENT_ID")
	stravaClientSecret = os.Getenv("STRAVA_CLIENT_SECRET")
)

type StravaAuthorizeCodeBody struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
	GrantType    string `json:"grant_type"`
}

type StravaAuthorizedUser struct {
	TokenType    string `json:"token_type"`
	ExpiresAt    uint32 `json:"expires_at"`
	ExpiresIn    uint32 `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	Athlete      struct {
		Id        int    `json:"id"`
		FirstName string `json:"firstname"`
		LastName  string `json:"lastname"`
		Profile   string `json:"profile"`
	} `json:"athlete"`
}

type StravaActivity struct {
	Id          uint64  `json:"id"`
	Name        string  `json:"name"`
	Distance    float32 `json:"distance"`
	StartDate   string  `json:"start_date"`
	ElapsedTime uint32  `json:"elapsed_time"`
	MovingTime  uint32  `json:"moving_time"`
	// Pace        string  `json:"pace"`
}

// Store this here for now :)
var authorizedUser StravaAuthorizedUser

func httpInternalServerError(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if (StravaAuthorizedUser{}) == authorizedUser {
		log.Println("User is not authorized")
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	} else {
		log.Printf("Found authorized user %s %s", authorizedUser.Athlete.FirstName, authorizedUser.Athlete.LastName)
		http.Redirect(w, r, "/activities", http.StatusTemporaryRedirect)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/login.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		log.Println("Failed to execute templates", err)
		httpInternalServerError(w, r)
	}
}

func activitiesHandler(w http.ResponseWriter, r *http.Request) {
	if (StravaAuthorizedUser{}) == authorizedUser {
		log.Println("User is not authorized")
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// New GET request to fetch a strava athletes activites
	client := &http.Client{}
	u, err := url.Parse("https://www.strava.com/api/v3/activities")
	if err != nil {
		log.Println("Failed to parse activities api url", err)
		httpInternalServerError(w, r)
		return
	}
	q := u.Query()
	q.Set("per_page", "10")
	u.RawQuery = q.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Println("Failed to created new request", err)
		httpInternalServerError(w, r)
	}
	req.URL.Query().Add("per_page", "10")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authorizedUser.AccessToken))
	resp, err := client.Do(req)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to complete request to %s", u.String()))
		httpInternalServerError(w, r)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Unable to read response body", err)
		httpInternalServerError(w, r)
		return
	}
	var activities []StravaActivity
	if err := json.Unmarshal(respBody, &activities); err != nil {
		log.Println("Unable to unmarshal response", err)
		httpInternalServerError(w, r)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/activities.html"))
	if err := tmpl.Execute(w, activities); err != nil {
		log.Println("Failed to execute templates", err)
		httpInternalServerError(w, r)
	}
}

func authorizeHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse("https://www.strava.com/oauth/authorize")
	if err != nil {
		log.Println("Failed to parse strava authorize url", err)
		httpInternalServerError(w, r)
		return
	}
	q := u.Query()
	q.Set("client_id", stravaClientId)
	// TODO: How do we get the scheme??
	q.Set("redirect_uri", fmt.Sprintf("http://%s/strava_callback", r.Host))
	q.Set("response_type", "code")
	q.Set("scope", "activity:read_all")
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		log.Println(fmt.Sprintf("Request URL (%s) is missing code param", r.URL))
		http.Error(w, "Missing code param", http.StatusBadRequest)
		return
	}
	authorizeBody := StravaAuthorizeCodeBody{
		ClientId:     stravaClientId,
		ClientSecret: stravaClientSecret,
		Code:         code,
		GrantType:    "authorization_code",
	}
	requestBody, err := json.Marshal(authorizeBody)
	if err != nil {
		log.Println("Unable to build authorize request body", err)
		httpInternalServerError(w, r)
		return
	}
	response, err := http.Post("https://www.strava.com/oauth/token", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println("Unable to request strava token", err)
		httpInternalServerError(w, r)
		return
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println("Unable to read response body", err)
		httpInternalServerError(w, r)
		return
	}
	var userData StravaAuthorizedUser
	if err := json.Unmarshal(responseBody, &userData); err != nil {
		log.Println("Unable to unmarshal response", err)
		httpInternalServerError(w, r)
		return
	}
	log.Printf("Authorized athlete %s %s", userData.Athlete.FirstName, userData.Athlete.LastName)
	authorizedUser = userData
	http.Redirect(w, r, "/activities", http.StatusTemporaryRedirect)
}

func main() {
	http.HandleFunc("/strava_callback", callbackHandler)
	http.HandleFunc("/strava_authorize", authorizeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/activities", activitiesHandler)
	http.HandleFunc("/", indexHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
