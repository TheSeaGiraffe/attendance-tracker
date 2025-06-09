package main

import (
	"log"
	"net/http"

	"github.com/TheSeaGiraffe/attendance-tracker/views/pages"
	"github.com/go-chi/chi/v5"
)

func main() {
	fs := http.FileServer(http.Dir("static"))

	r := chi.NewRouter()
	r.Handle("/static/*", http.StripPrefix("/static/", fs))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		err := pages.LoginPage().Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	err := http.ListenAndServe(":3000", r)
	if err != nil {
		log.Fatal(err)
	}
}
