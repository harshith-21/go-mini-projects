package main

import (
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

type Entry struct {
	ShortURL  string
	TargetURL string
}

func homePageHandler(data *[]Entry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hometmpl := template.Must(template.ParseFiles("templates/home.html"))
		err := hometmpl.Execute(w, *data)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func shortenHandler(data *[]Entry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetURL := r.FormValue("TargetUrl")
		shortURL := r.FormValue("ShortUrl")

		entry := Entry{
			ShortURL:  shortURL,
			TargetURL: targetURL,
		}

		*data = append(*data, entry)

		tmpl := template.Must(
			template.ParseFiles("templates/url.html"),
		)

		if err := tmpl.Execute(w, entry); err != nil {
			http.Error(
				w,
				"Internal server error",
				http.StatusInternalServerError,
			)
		}
	}
}

func redirectHandler(data *[]Entry, redirectSleep int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortURL := r.PathValue("shortURL")

		for _, entry := range *data {
			if entry.ShortURL == shortURL {
				w.Header().Set("Content-Type", "text/html")

				fmt.Fprintf(w, `
					<!DOCTYPE html>
					<html>
					<head>
						<title>Redirecting...</title>
					</head>
					<body>
						<h1>Redirecting...</h1>
						<p>You'll be redirected in %d seconds...</p>

						<div
							hx-get="/redirect/%s"
							hx-trigger="load delay:%ds"
							hx-swap="none">
						</div>

						<script src="https://cdn.jsdelivr.net/npm/htmx.org@2.0.11/dist/htmx.min.js"></script>
					</body>
					</html>
				`, redirectSleep, shortURL, redirectSleep)

				return
			}
		}

		http.NotFound(w, r)
	}
}

func performRedirectHandler(data *[]Entry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortURL := r.PathValue("shortURL")

		for _, entry := range *data {
			if entry.ShortURL == shortURL {
				w.Header().Set("HX-Redirect", entry.TargetURL)
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		http.NotFound(w, r)
	}
}

func deleteHandler(data *[]Entry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortURL := r.PathValue("shortURL")
		for i, entry := range *data {
			if entry.ShortURL == shortURL {
				*data = append((*data)[:i], (*data)[i+1:]...)
				return
			}
		}
		http.NotFound(w, r)
	}
}

func main() {

	data := []Entry{
		{
			ShortURL:  "goog",
			TargetURL: "https://google.com",
		},
	}

	RedirectSleepStr := os.Getenv("REDIR_SLEEP")
	var RedirectSleep int = 0;
	if len(RedirectSleepStr) > 0 {
		RedirectSleeptemp, err := strconv.Atoi(RedirectSleepStr)
		if err != nil {
			log.Println("Failed to convert RedirectSleep to int, defaulting to 0")
		} else {
			RedirectSleep = RedirectSleeptemp
			log.Println("REDIR_SLEEP Env var detected, and Redirect sleep is now set to", RedirectSleep)
		}
	} 

	mux := http.NewServeMux()

	mux.Handle(
		"GET /static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(http.Dir("./static")),
		),
	)

	mux.HandleFunc("POST /shorten", shortenHandler(&data))
	mux.HandleFunc("GET /", homePageHandler(&data))
	mux.HandleFunc("GET /{shortURL}", redirectHandler(&data, RedirectSleep))
	mux.HandleFunc("GET /redirect/{shortURL}", performRedirectHandler(&data))
	mux.HandleFunc("DELETE /{shortURL}", deleteHandler(&data))



	if err := http.ListenAndServe(":8080", mux); err != nil {
		slog.Error("server stopped", "error", err)
	}
}
