package main

import (
	"log"
	"net/http"

	"github.com/TheSeaGiraffe/attendance-tracker/controllers"
	"github.com/go-chi/chi/v5"
)

func main() {
	// Init controllers
	usersC := controllers.Users{}

	// Setup routes
	fs := http.FileServer(http.Dir("static"))

	r := chi.NewRouter()
	r.Handle("/static/*", http.StripPrefix("/static/", fs))
	r.Get("/", usersC.LogIn)

	// Start server
	err := http.ListenAndServe(":3000", r)
	if err != nil {
		log.Fatal(err)
	}
}
