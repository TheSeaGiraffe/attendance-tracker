package main

import (
	"context"
	"log"
	"net/http"

	"github.com/TheSeaGiraffe/attendance-tracker/config"
	"github.com/TheSeaGiraffe/attendance-tracker/controllers"
	"github.com/TheSeaGiraffe/attendance-tracker/database/queries"
	"github.com/TheSeaGiraffe/attendance-tracker/services"
	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// TODO: need to check that the app is actually connected to the DB
	// Set up database connection pool. Use defaults for now.
	dbConfig := config.DefaultConfig()
	dbPool, err := pgxpool.New(context.Background(), dbConfig.String())
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()
	sqlcQuery := queries.New(dbPool)

	// Init services and controllers
	userService := &services.UserService{
		DB: sqlcQuery,
	}

	sessionManager := scs.New() // this is practically a service
	sessionManager.Store = pgxstore.New(dbPool)
	sessionManager.Cookie.Secure = true

	usersC := controllers.Users{
		UserService:    userService,
		SessionManager: sessionManager,
	}

	umw := controllers.UserMiddleware{
		SessionManager: sessionManager,
	}

	// Setup routes
	fs := http.FileServer(http.Dir("static"))

	r := chi.NewRouter()

	r.Use(sessionManager.LoadAndSave)

	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	r.Get("/", usersC.Home) // Not sure if this is the proper way to do it.
	r.Get("/login", usersC.LogIn)
	r.Post("/login", usersC.ProcessLogIn)
	r.Post("/logout", usersC.ProcessLogOut)

	r.Route("/users", func(r chi.Router) {
		r.Use(umw.RequireUser)
		r.Get("/", usersC.UserHome)
	})

	r.Route("/admin", func(r chi.Router) {
		r.Use(umw.RequireUser)
		r.Get("/", usersC.AdminHome)
	})

	// Start server
	err = http.ListenAndServe(":3000", r)
	if err != nil {
		log.Fatal(err)
	}
}
