package entities

import (
	"time"

	"github.com/google/uuid"
)

// UserDataModel represents an application user in the database.
// Users are provisioned/authenticated through Google Sign-In; GoogleID
// holds the Google account subject ("sub") claim and is unique per user.
type UserDataModel struct {
	ID            string    `json:"id" bson:"id,omitempty"`
	Email         string    `json:"email" bson:"email,omitempty"`
	Name          string    `json:"name" bson:"name,omitempty"`
	Picture       string    `json:"picture" bson:"picture,omitempty"`
	GoogleID      string    `json:"google_id" bson:"google_id,omitempty"`
	Provider      string    `json:"provider" bson:"provider,omitempty"` // e.g. "google" or "password"
	PasswordHash  string    `json:"-" bson:"password_hash,omitempty"`   // bcrypt hash; never serialized to clients
	EmailVerified bool      `json:"email_verified" bson:"email_verified,omitempty"`
	LastLoginAt   time.Time `json:"last_login_at" bson:"last_login_at,omitempty"`
	CreatedAt     time.Time `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at" bson:"updated_at,omitempty"`
}

// NewUser initializes a new UserDataModel with a UUIDv4 ID and current timestamps.
func NewUser() UserDataModel {
	now := time.Now().UTC()
	return UserDataModel{
		ID:        uuid.NewString(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// UserFilter is used to query users by an optional set of fields.
type UserFilter struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	GoogleID string `json:"google_id"`
	Provider string `json:"provider"`
}

// SignUpRequest is the body posted to the email/password registration endpoint.
type SignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// SignInRequest is the body posted to the email/password login endpoint.
type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// GoogleSignInRequest is the body posted to the Google sign-in endpoint.
// IDToken is the Google ID token (credential) obtained on the frontend.
type GoogleSignInRequest struct {
	IDToken string `json:"id_token"`
}

// GoogleTokenInfo maps the response from Google's tokeninfo endpoint.
// email_verified is returned as a string ("true"/"false") by Google.
type GoogleTokenInfo struct {
	Aud           string `json:"aud"`
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Exp           string `json:"exp"`
}

// AuthResponse is returned to the client after a successful sign-in.
type AuthResponse struct {
	Token     string        `json:"token"`
	ExpiresIn int64         `json:"expires_in"`
	User      UserDataModel `json:"user"`
}
