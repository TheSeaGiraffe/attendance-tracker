package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/TheSeaGiraffe/attendance-tracker/database/queries"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	DB *queries.Queries
}

func (us UserService) Authenticate(email, password string) (*queries.User, error) {
	email = strings.ToLower(email)
	user, err := us.DB.GetUserByEmail(context.Background(), email)
	if err != nil {
		return nil, fmt.Errorf("authenticate: could not retrieve user from DB: %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("authenticate: could not hash password: %w", err)
	}
	return &user, nil
}
