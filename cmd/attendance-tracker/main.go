package main

import (
	"context"
	"log"
	"net/http"

	"github.com/TheSeaGiraffe/attendance-tracker/config"
	"github.com/TheSeaGiraffe/attendance-tracker/controllers"
	"github.com/TheSeaGiraffe/attendance-tracker/database/queries"
	"github.com/TheSeaGiraffe/attendance-tracker/services"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

func main() {
	// Setup database connection
	// Basic one for now with no pooling. Just want to make sure everything is working first.
	dbConfig := config.DefaultConfig()
	dbConn, err := pgx.Connect(context.Background(), dbConfig.String())
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close(context.Background())
	sqlcQuery := queries.New(dbConn)

	// Init services and controllers
	userService := &services.UserService{
		DB: sqlcQuery,
	}

	usersC := controllers.Users{
		UserService: userService,
	}

	// Setup routes
	fs := http.FileServer(http.Dir("static"))

	r := chi.NewRouter()
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	// What I might do here is create a fake home page handler that just redirects to the login page.
	// Will leave it like this for now.
	r.Get("/", usersC.LogIn)
	r.Post("/", usersC.ProcessLogIn)

	r.Route("/users", func(r chi.Router) {
		r.Get("/", usersC.UserHome)
	})

	r.Route("/admin", func(r chi.Router) {
		r.Get("/", usersC.AdminHome)
	})

	// Start server
	err = http.ListenAndServe(":3000", r)
	if err != nil {
		log.Fatal(err)
	}
}
