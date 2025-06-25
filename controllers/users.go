package controllers

import (
	"context"
	"encoding/gob"
	"errors"
	"net/http"
	"net/url"

	"github.com/TheSeaGiraffe/attendance-tracker/database/queries"
	"github.com/TheSeaGiraffe/attendance-tracker/services"
	"github.com/TheSeaGiraffe/attendance-tracker/views/pages"
	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/csrf"
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
	UserService          *services.UserService
	SessionManager       *scs.SessionManager
	PasswordResetService *services.PasswordResetService
	EmailService         *services.EmailService
}

func (u Users) Home(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/signin", http.StatusFound)
}

func (u Users) SignUp(w http.ResponseWriter, r *http.Request) {
	var data struct {
		csrfToken string
		name      string
		email     string
	}
	data.csrfToken = csrf.Token(r)
	data.name = r.FormValue("name")
	data.email = r.FormValue("email")
	// err := pages.SignUpPage(data.csrfToken, data.name, data.email, false).Render(context.Background(), w)
	err := pages.SignUpPage(data.name, data.email, false).Render(context.Background(), w)
	if err != nil {
		// TODO: Log this error
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (u Users) ProcessSignUp(w http.ResponseWriter, r *http.Request) {
	var data struct {
		csrfToken string
		name      string
		email     string
		password  string
	}

	data.name = r.FormValue("name")
	data.email = r.FormValue("email")
	data.password = r.FormValue("password")

	user, err := u.UserService.New(data.name, data.email, data.password)
	if err != nil {
		// TODO: log this error
		if errors.Is(err, services.ErrEmailTaken) {
			// Find a way to better handle this error. It should be displayed to the user
			// in a manner similar to how errors are displayed in the `lenslocked` app.
			// For now, log the error and send users back to the sign up page.
			data.csrfToken = csrf.Token(r)
			// err = pages.SignUpPage(data.csrfToken, data.name, data.email, false).Render(context.Background(), w)
			err = pages.SignUpPage(data.name, data.email, false).Render(context.Background(), w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gob.Register(queries.User{})
	err = u.SessionManager.RenewToken(r.Context())
	if err != nil {
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		// TODO: log these errors. Find a way to show the user a warning that they couldn't be
		// logged in.
		data.csrfToken = csrf.Token(r)
		// err := pages.SignInPage(data.csrfToken, data.email, false).Render(r.Context(), w)
		err := pages.SignInPage(data.email, false).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	u.SessionManager.Put(r.Context(), sessionUserKey, user)

	if user.IsAdmin {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/users", http.StatusFound)
}

func (u Users) SignIn(w http.ResponseWriter, r *http.Request) {
	var data struct {
		csrfToken string
		email     string
	}
	data.csrfToken = csrf.Token(r)
	data.email = r.FormValue("email")
	// err := pages.SignInPage(data.csrfToken, data.email, false).Render(r.Context(), w)
	err := pages.SignInPage(data.email, false).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (u Users) ProcessSignIn(w http.ResponseWriter, r *http.Request) {
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

	if user.IsAdmin {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/users", http.StatusFound)
}

func (u Users) ProcessSignOut(w http.ResponseWriter, r *http.Request) {
	err := u.SessionManager.Destroy(r.Context())
	if err != nil {
		// TODO: log this error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/signin", http.StatusFound)
}

func (u Users) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	// csrfToken := csrf.Token(r)
	email := r.FormValue("email")
	// err := pages.ForgotPassword(csrfToken, false).Render(r.Context(), w)
	err := pages.ForgotPassword(email, false).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (u Users) ProcessForgotPassword(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	user, err := u.UserService.DB.GetUserByEmail(context.Background(), email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	token, err := u.PasswordResetService.Create(email)
	if err != nil {
		// TODO: Handle other cases in the future. For instance, if a user doesn't exist
		// with the email address. Also, log this error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	vals := url.Values{
		"token": {token},
	}
	// TODO: Make the URL here configurable
	err = u.EmailService.ForgotPassword(user.Name, user.Email, "https://www.agrisoft-attendance.com/reset-pw?"+vals.Encode())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = pages.CheckYourEmail(email, false).Render(context.Background(), w)
	if err != nil {
		// TODO: log this error
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (u Users) ResetPassword(w http.ResponseWriter, r *http.Request) {
	// csrfToken := csrf.Token(r)
	resetToken := r.FormValue("token")
	// err := pages.ResetPassword(csrfToken, resetToken, false).Render(context.Background(), w)
	err := pages.ResetPassword(resetToken, false).Render(context.Background(), w)
	if err != nil {
		// TODO; log this error
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (u Users) ProcessResetPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token    string
		Password string
	}
	data.Token = r.FormValue("token")
	data.Password = r.FormValue("password")

	user, err := u.PasswordResetService.Consume(data.Token)
	if err != nil {
		// TODO: Distinguish between server errors and invalid token errors
		// Also, log this error.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = u.UserService.UpdatePassword(int(user.ID), data.Password)
	if err != nil {
		// TODO: log this error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gob.Register(queries.User{})
	err = u.SessionManager.RenewToken(r.Context())
	if err != nil {
		// TODO: log this error
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

func (u Users) UserHome(w http.ResponseWriter, r *http.Request) {
	err := pages.UserHome(true).Render(r.Context(), w)
	if err != nil {
		// TODO: log this error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Not sure I should be combining user and admin handlers in the same controller. Will
// think about separating them later. Keep like this for now.
func (u Users) AdminHome(w http.ResponseWriter, r *http.Request) {
	err := pages.AdminHome(true).Render(r.Context(), w)
	if err != nil {
		// TODO: log this error
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
			http.Redirect(w, r, "/signin", http.StatusFound)
			return
		}

		user, ok := umw.SessionManager.Get(r.Context(), sessionUserKey).(queries.User)
		if !ok {
			// Should I just redirect here?
			// TODO: log this error
			http.Error(w, "could not hydrate user", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), ctxUserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
