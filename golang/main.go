package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

var (
	stravaClientId     = os.Getenv("STRAVA_CLIENT_ID")
	stravaClientSecret = os.Getenv("STRAVA_CLIENT_SECRET")
)

type stravaAuthorizeBody struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
	GrantType    string `json:"grant_type"`
}

type stravaAuthorizedUser struct {
	TokenType    string        `json:"token_type"`
	ExpiresAt    uint32        `json:"expires_at"`
	ExpiresIn    uint32        `json:"expires_in"`
	RefreshToken string        `json:"refresh_token"`
	AccessToken  string        `json:"access_token"`
	Athlete      stravaAthlete `json:"athlete"`
}

func (u *stravaAuthorizedUser) hasToken() bool {
	return u.AccessToken != ""
}

type stravaAthlete struct {
	Id        int    `json:"id"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Profile   string `json:"profile"`
}

func (a *stravaAthlete) fullName() string {
	return fmt.Sprintf("%s %s", a.FirstName, a.LastName)
}

type stravaActivity struct {
	Id          uint64  `json:"id"`
	Name        string  `json:"name"`
	Distance    float32 `json:"distance"`
	StartDate   string  `json:"start_date"`
	ElapsedTime uint32  `json:"elapsed_time"`
	MovingTime  uint32  `json:"moving_time"`
}

var (
	authorizedUser   stravaAuthorizedUser
	athleteActivites []stravaActivity
)

// Trigger an Internal Server Error
func httpInternalServerError(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// indexHandler will redirect an authorized Strava user to the activities page.
// If user is not authorized, then user is redirected to the login page.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if !authorizedUser.hasToken() {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	log.Printf("Found authorized user %s", authorizedUser.Athlete.fullName())
	http.Redirect(w, r, "/activities", http.StatusTemporaryRedirect)
}

// Renders the login page.
func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/login.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		log.Println("Failed to execute templates", err)
		httpInternalServerError(w, r)
	}
}

// Renders the activities page for authorized Strava users.
func activitiesHandler(w http.ResponseWriter, r *http.Request) {
	if !authorizedUser.hasToken() {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

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
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authorizedUser.AccessToken))
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Unable to complete request to %s", u.String())
		httpInternalServerError(w, r)
		return
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&athleteActivites); err != nil {
		log.Println("Failed to decode response body.")
		httpInternalServerError(w, r)
		return
	}

	data := struct {
		Activities       []stravaActivity
		SelectedActivity stravaActivity
	}{
		Activities:       athleteActivites,
		SelectedActivity: athleteActivites[0],
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/activities.html", "templates/activity.html"))
	if err := tmpl.Execute(w, data); err != nil {
		log.Println("Failed to execute templates", err)
		httpInternalServerError(w, r)
		return
	}
}

// Parses the actvitiy id from a request path.
// It is currently meant for the getActivityHandler.
// The expected request path is /activities/:activity_id.
func parseIdFromPath(s string) (int, error) {
	split := strings.Split(strings.Trim(s, "/"), "/")
	splitLen := len(split)
	if splitLen != 2 {
		return -1, errors.New("invalid path")
	}

	id, err := strconv.Atoi(split[1])
	if err != nil {
		return -1, err
	}

	return id, nil
}

// Will fetch and render a single strave activity.
func getActivityHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseIdFromPath(r.URL.Path)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to parse id from path %s", r.URL.Path), err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	index := slices.IndexFunc(athleteActivites, func(a stravaActivity) bool {
		return a.Id == uint64(id)
	})

	if index == -1 {
		log.Printf("Acivity ID %d not found", id)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.New("activity").ParseFiles("templates/activity.html"))
	if err := tmpl.Execute(w, athleteActivites[index]); err != nil {
		log.Println("Failed to execute templates", err)
		httpInternalServerError(w, r)
		return
	}
}

// Will attempt to authorize a Strava user and redirect
// to a callback handler if successful.
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

// callbackHandler will handle the token exchange for an authorized Strava user.
// If successful, a short lived token and refresh token will be created.
func callbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		log.Printf("Request URL (%s) is missing code param", r.URL)
		http.Error(w, "Missing code param", http.StatusBadRequest)
		return
	}
	authorizeBody := stravaAuthorizeBody{
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
	if err := json.NewDecoder(response.Body).Decode(&authorizedUser); err != nil {
		log.Println("Failed to decode response body.")
		httpInternalServerError(w, r)
		return
	}
	log.Printf("Authorized athlete %s", authorizedUser.Athlete.fullName())
	http.Redirect(w, r, "/activities", http.StatusTemporaryRedirect)
}

func main() {
	http.HandleFunc("/strava_callback", callbackHandler)
	http.HandleFunc("/strava_authorize", authorizeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/activities/", getActivityHandler)
	http.HandleFunc("/activities", activitiesHandler)
	http.HandleFunc("/", indexHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
