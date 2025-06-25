package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/TheSeaGiraffe/attendance-tracker/database/queries"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var ErrEmailTaken = errors.New("email address is already in use")

type UserService struct {
	DB *queries.Queries
}

func (us UserService) New(name, email, password string) (*queries.User, error) {
	email = strings.ToLower(email)
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("create user: could not hash password: %w", err)
	}
	createUserParams := queries.CreateNewUserParams{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashedPass),
		IsAdmin:      false,
	}
	userID, err := us.DB.CreateNewUser(context.Background(), createUserParams)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			if pgError.Code == pgerrcode.UniqueViolation {
				return nil, ErrEmailTaken
			}
		}
		return nil, fmt.Errorf("new: could not create new user: %w", err)
	}

	// Not sure if this is the best way of going about it. Will leave it like this for now.
	user := queries.User{
		ID:           userID,
		Name:         createUserParams.Name,
		Email:        createUserParams.Email,
		PasswordHash: createUserParams.PasswordHash,
		IsAdmin:      createUserParams.IsAdmin,
	}

	return &user, nil
}

func (us UserService) Authenticate(email, password string) (*queries.User, error) {
	email = strings.ToLower(email)
	user, err := us.DB.GetUserByEmail(context.Background(), email)
	if err != nil {
		return nil, fmt.Errorf("authenticate: could not retrieve user from DB: %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("authenticate: comparing passwords: %w", err)
	}
	return &user, nil
}

func (us *UserService) UpdatePassword(userID int, password string) error {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("update password: could not hash password: %w", err)
	}
	updateParams := queries.UpdateUserPasswordParams{
		ID:           int32(userID),
		PasswordHash: string(hashedPass),
	}
	err = us.DB.UpdateUserPassword(context.Background(), updateParams)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}
