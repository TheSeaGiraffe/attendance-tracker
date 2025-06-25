package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/TheSeaGiraffe/attendance-tracker/config"
	"github.com/TheSeaGiraffe/attendance-tracker/controllers"
	"github.com/TheSeaGiraffe/attendance-tracker/database/queries"
	"github.com/TheSeaGiraffe/attendance-tracker/services"
	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("could not load .env file: %v", err)
	}

	// TODO: need to check that the app is actually connected to the DB
	// Set up database connection pool. Use defaults for now.
	dbConfig := config.DefaultConfig()
	dbPool, err := pgxpool.New(context.Background(), dbConfig.String())
	if err != nil {
		log.Fatalf("could not create database connection pool: %v", err)
	}
	defer dbPool.Close()
	sqlcQuery := queries.New(dbPool)

	// Init services and controllers
	userService := &services.UserService{
		DB: sqlcQuery,
	}

	passwordResetService := &services.PasswordResetService{
		DB: sqlcQuery,
	}

	sessionManager := scs.New() // this is practically a service
	sessionManager.Store = pgxstore.New(dbPool)
	sessionManager.Cookie.Secure = true

	// TODO: add error handling for the case where the API key isn't provided. Should probably
	// either panic or otherwise force the server to exit
	emailService := services.NewEmailService(os.Getenv("MAILERSEND_API_KEY"))

	usersC := controllers.Users{
		UserService:          userService,
		SessionManager:       sessionManager,
		PasswordResetService: passwordResetService,
		EmailService:         emailService,
	}

	umw := controllers.UserMiddleware{
		SessionManager: sessionManager,
	}

	// Disable for now. Will add CSRF protection once all of the user account stuff is working.
	// TODO: add error handling for the case where neither the csrf key or the "CSRF_SECURE"
	// value are set. For the csrf key, should probably either panic or just force the server
	// to exit
	// csrfMw := csrf.Protect(
	// 	[]byte(os.Getenv("CSRF_KEY")),
	// 	csrf.Secure(os.Getenv("CSRF_SECURE") == "true"),
	// )

	// Setup routes
	fs := http.FileServer(http.Dir("static"))

	r := chi.NewRouter()

	// r.Use(csrfMw)
	r.Use(sessionManager.LoadAndSave)

	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	r.Get("/", usersC.Home) // Not sure if this is the proper way to do it.

	// Both of the sign up endpoints are "hidden" in the sense that none of the pages
	// ever call them. Might have to think about adding a flag that enables them only
	// for debugging.
	r.Get("/signup", usersC.SignUp)
	r.Post("/signup", usersC.ProcessSignUp)

	r.Get("/signin", usersC.SignIn)
	r.Post("/signin", usersC.ProcessSignIn)
	r.Post("/signout", usersC.ProcessSignOut)
	r.Get("/forgot-pw", usersC.ForgotPassword)
	r.Post("/forgot-pw", usersC.ProcessForgotPassword)
	r.Get("/reset-pw", usersC.ResetPassword)
	r.Post("/reset-pw", usersC.ProcessResetPassword)

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
