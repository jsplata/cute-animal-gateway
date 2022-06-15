package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type application struct {
	auth struct {
		username string
		password string
	}
}

type FoxJSON struct {
	Image string
	Link  string
}

type DogJSON struct {
	Message string
	Status  string
}

type CatJSON struct {
	File string
}

func main() {
	app := new(application)

	app.auth.username = os.Getenv("AUTH_USERNAME")
	app.auth.password = os.Getenv("AUTH_PASSWORD")

	if app.auth.username == "" {
		log.Fatal("basic auth username must be provided")
	}

	if app.auth.password == "" {
		log.Fatal("basic auth password must be provided")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/fox", app.basicAuth(app.foxHandler))
	mux.HandleFunc("/dog", app.basicAuth(app.dogHandler))
	mux.HandleFunc("/cat", app.basicAuth(app.catHandler))

	srv := &http.Server{
		Addr:    ":4000",
		Handler: mux,
	}

	log.Printf("starting server on %s", srv.Addr)
	err := srv.ListenAndServe()
	log.Fatal(err)
}

func (app *application) foxHandler(w http.ResponseWriter, r *http.Request) {
	fox := new(FoxJSON)
	getJSON("https://randomfox.ca/floof/", fox)
	fmt.Fprintln(w, "<html><img src=\""+fox.Image+"\"></html>")
}

func (app *application) dogHandler(w http.ResponseWriter, r *http.Request) {
	dog := new(DogJSON)
	getJSON("https://dog.ceo/api/breeds/image/random", dog)
	fmt.Fprintln(w, "<html><img src=\""+dog.Message+"\"></html>")
}

func (app *application) catHandler(w http.ResponseWriter, r *http.Request) {
	cat := new(CatJSON)
	getJSON("https://aws.random.cat/meow", cat)
	fmt.Fprintln(w, "<html><img src=\""+cat.File+"\"></html>")
}

func getJSON(url string, target interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func (app *application) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {

			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte("your expected username"))
			expectedPasswordHash := sha256.Sum256([]byte("your expected password"))

			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}
