package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/TheSeaGiraffe/attendance-tracker/database/queries"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	MinBytesPerToken     = 32
	DefaultResetDuration = 1 * time.Hour
)

// Might want to think about spinning up a goroutine that regularly checks for orphaned
// reset tokens and removes them if they have already expired. I'll have to figure out
// an efficient way of polling the database. The polling doesn't have to be too agressive,
// maybe something like once every half-hour.

type PasswordResetService struct {
	DB            *queries.Queries
	BytesPerToken int
	Duration      time.Duration
}

func createResetToken(numBytes int) (string, error) {
	// Create byte slice and ensure that it is at least the length of `MinBytesPerToken`
	b := make([]byte, numBytes)
	nRead, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("could not read random bytes: %w", err)
	}
	if nRead < numBytes {
		return "", fmt.Errorf("didn't read enough random bytes")
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func hash(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}

func (ps *PasswordResetService) Create(email string) (string, error) {
	email = strings.ToLower(email)
	user, err := ps.DB.GetUserByEmail(context.Background(), email)
	if err != nil {
		return "", fmt.Errorf("could not find user with specified email: %w", err)
	}

	bytesPerToken := ps.BytesPerToken
	if bytesPerToken == 0 {
		bytesPerToken = MinBytesPerToken
	}
	token, err := createResetToken(bytesPerToken)
	if err != nil {
		return "", fmt.Errorf("could not create refresh token: %w", err)
	}
	duration := ps.Duration
	if duration == 0 {
		duration = DefaultResetDuration
	}
	createTokenParams := queries.CreateTokenForUserParams{
		UserID:    pgtype.Int4{Int32: user.ID, Valid: true},
		TokenHash: hash(token),
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(duration), InfinityModifier: pgtype.Finite, Valid: true},
	}
	_, err = ps.DB.CreateTokenForUser(context.Background(), createTokenParams)
	if err != nil {
		return "", fmt.Errorf("could not insert token into database: %w", err)
	}

	// Even though we just need the token, will leave like this for now since we
	// might need this info in the future. Will remove if we end up not needing this.
	// pwReset := queries.PasswordReset{
	// 	ID:        tokenID,
	// 	UserID:    createTokenParams.UserID,
	// 	TokenHash: createTokenParams.TokenHash,
	// 	ExpiresAt: createTokenParams.ExpiresAt,
	// }

	return token, nil
}

func (ps *PasswordResetService) Consume(token string) (*queries.User, error) {
	tokenHash := hash(token)
	userPassInfo, err := ps.DB.GetUserForToken(context.Background(), tokenHash)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve user associated with password reset token from database: %w", err)
	}
	if time.Now().After(userPassInfo.ExpiresAt.Time) {
		return nil, fmt.Errorf("password reset token expired: %v", token)
	}
	err = ps.DB.DeleteTokenById(context.Background(), userPassInfo.ResetTokenID)
	if err != nil {
		return nil, fmt.Errorf("could not delete password reset token from database: %w", err)
	}
	user := queries.User{
		ID:           userPassInfo.UserID,
		Name:         userPassInfo.Name,
		Email:        userPassInfo.Email,
		PasswordHash: userPassInfo.PasswordHash,
		IsAdmin:      userPassInfo.IsAdmin,
	}
	return &user, nil
}
