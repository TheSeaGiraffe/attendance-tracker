package controllers

import (
	"net/http"

	"github.com/TheSeaGiraffe/attendance-tracker/services"
	"github.com/TheSeaGiraffe/attendance-tracker/views/pages"
)

type Users struct {
	UserService *services.UserService
}

func (u Users) LogIn(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	err := pages.LoginPage(email).Render(r.Context(), w)
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
	data.email = r.FormValue("email")
	data.password = r.FormValue("password")
	// TODO: add user to the session.
	user, err := u.UserService.Authenticate(data.email, data.password)
	if err != nil {
		// TODO: log this error and add better http error handling
		http.Error(w, "Problem authenticating user", http.StatusInternalServerError)
	}
	// TODO: Redirect user to user page. Make sure to check if user is admin
	if user.IsAdmin {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/users", http.StatusFound)
}

func (u Users) UserHome(w http.ResponseWriter, r *http.Request) {
	err := pages.UserHome().Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Not sure I should be combining user and admin handlers in the same controller. Will
// think about separating them later. Keep like this for now.
func (u Users) AdminHome(w http.ResponseWriter, r *http.Request) {
	err := pages.AdminHome().Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
