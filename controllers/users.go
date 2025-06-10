package controllers

import (
	"net/http"

	"github.com/TheSeaGiraffe/attendance-tracker/views/pages"
)

type Users struct{}

func (u Users) LogIn(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	err := pages.LoginPage(email).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
