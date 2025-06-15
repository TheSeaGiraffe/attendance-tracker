package controllers

import (
	"context"
	"encoding/gob"
	"net/http"

	"github.com/TheSeaGiraffe/attendance-tracker/database/queries"
	"github.com/TheSeaGiraffe/attendance-tracker/services"
	"github.com/TheSeaGiraffe/attendance-tracker/views/pages"
	"github.com/alexedwards/scs/v2"
)

type key string

const (
	sessionUserKey     = "user"
	ctxUserKey     key = "user" // Not sure about this; leave for now
)

func getUserFromContext(r *http.Request) queries.User {
	return r.Context().Value(sessionUserKey).(queries.User)
}

type Users struct {
	UserService    *services.UserService
	SessionManager *scs.SessionManager
}

func (u Users) Home(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (u Users) LogIn(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	err := pages.LoginPage(email, false).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (u Users) ProcessLogIn(w http.ResponseWriter, r *http.Request) {
	var data struct {
		email    string
		password string
	}
	// TODO: Add email and password validation
	data.email = r.FormValue("email")
	data.password = r.FormValue("password")
	user, err := u.UserService.Authenticate(data.email, data.password)
	if err != nil {
		// TODO: log this error and add better http error handling
		http.Error(w, "Problem authenticating user", http.StatusInternalServerError)
	}

	gob.Register(queries.User{})
	err = u.SessionManager.RenewToken(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	u.SessionManager.Put(r.Context(), sessionUserKey, user)

	// TODO: Redirect user to user page. Make sure to check if user is admin
	if user.IsAdmin {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/users", http.StatusFound)
}

func (u Users) ProcessLogOut(w http.ResponseWriter, r *http.Request) {
	err := u.SessionManager.Destroy(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (u Users) UserHome(w http.ResponseWriter, r *http.Request) {
	err := pages.UserHome(true).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Not sure I should be combining user and admin handlers in the same controller. Will
// think about separating them later. Keep like this for now.
func (u Users) AdminHome(w http.ResponseWriter, r *http.Request) {
	err := pages.AdminHome(true).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type UserMiddleware struct {
	SessionManager *scs.SessionManager
}

func (umw UserMiddleware) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !umw.SessionManager.Exists(r.Context(), sessionUserKey) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		user, ok := umw.SessionManager.Get(r.Context(), sessionUserKey).(queries.User)
		if !ok {
			// Should I just redirect here?
			http.Error(w, "could not hydrate user", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), ctxUserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
