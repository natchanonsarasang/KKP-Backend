package services

import (
	"errors"
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"go-fiber-template/src/client"
	"go-fiber-template/src/middlewares"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type usersService struct {
	UsersRepository      repositories.IUsersRepository
	GoogleOAuthClient    client.IGoogleOAuthClient
	MicrosoftOAuthClient client.IMicrosoftOAuthClient
}

type IUsersService interface {
	// Email/password registration: creates the user with a bcrypt-hashed
	// password and issues an application JWT.
	Register(req entities.SignUpRequest) (*entities.AuthResponse, error)
	// Email/password login: verifies the password and issues an application JWT.
	Login(req entities.SignInRequest) (*entities.AuthResponse, error)

	// Google Sign-In: verifies the Google ID token, upserts the user, and
	// issues an application JWT.
	GoogleSignIn(idToken string) (*entities.AuthResponse, error)

	// Microsoft Sign-In: verifies the Microsoft ID token, upserts the user, and
	// issues an application JWT.
	MicrosoftSignIn(idToken string) (*entities.AuthResponse, error)

	// CRUD
	CreateUser(data entities.UserDataModel) (*entities.UserDataModel, error)
	GetUserByID(id string) (*entities.UserDataModel, error)
	GetAllUsers(filter entities.UserFilter) (*[]entities.UserDataModel, error)
	UpdateUser(id string, data entities.UserDataModel) error
	DeleteUser(id string) error
}

func NewUsersService(repo repositories.IUsersRepository, googleClient client.IGoogleOAuthClient, microsoftClient client.IMicrosoftOAuthClient) IUsersService {
	return &usersService{
		UsersRepository:      repo,
		GoogleOAuthClient:    googleClient,
		MicrosoftOAuthClient: microsoftClient,
	}
}

// validateUser runs business-logic validations on a UserDataModel.
func (sv *usersService) validateUser(data *entities.UserDataModel) error {
	if data.Email == "" {
		return errors.New("email must not be empty")
	}
	return nil
}

// issueAuthResponse mints an application JWT for the user and wraps it with the
// user profile for return to the client.
func (sv *usersService) issueAuthResponse(user *entities.UserDataModel) (*entities.AuthResponse, error) {
	token, err := middlewares.GenerateJWTToken(user.ID, user.ID)
	if err != nil {
		return nil, err
	}

	resp := &entities.AuthResponse{
		Token: *token.Token,
		User:  *user,
	}
	if token.ExpiresIn != nil {
		resp.ExpiresIn = *token.ExpiresIn
	}
	return resp, nil
}

// ===== Email / Password =====

func (sv *usersService) Register(req entities.SignUpRequest) (*entities.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		return nil, errors.New("email must not be empty")
	}
	if len(req.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	existing, err := sv.UsersRepository.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("a user with this email already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	user := entities.NewUser()
	user.Email = email
	user.Name = req.Name
	user.Provider = "password"
	user.PasswordHash = string(hash)
	user.LastLoginAt = now

	if err := sv.UsersRepository.InsertUser(user); err != nil {
		return nil, err
	}

	return sv.issueAuthResponse(&user)
}

func (sv *usersService) Login(req entities.SignInRequest) (*entities.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.Password == "" {
		return nil, errors.New("email and password are required")
	}

	user, err := sv.UsersRepository.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	// Use a uniform error to avoid leaking which emails are registered.
	if user == nil || user.PasswordHash == "" {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	now := time.Now().UTC()
	user.LastLoginAt = now
	user.UpdatedAt = now
	if err := sv.UsersRepository.UpdateUser(user.ID, *user); err != nil {
		return nil, err
	}

	return sv.issueAuthResponse(user)
}

// ===== Google Sign-In =====

func (sv *usersService) GoogleSignIn(idToken string) (*entities.AuthResponse, error) {
	if idToken == "" {
		return nil, errors.New("id_token must not be empty")
	}

	info, err := sv.GoogleOAuthClient.VerifyIDToken(idToken)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	// Find an existing user by Google subject, falling back to email so accounts
	// created through another flow get linked instead of duplicated.
	user, err := sv.UsersRepository.FindByGoogleID(info.Sub)
	if err != nil {
		return nil, err
	}
	if user == nil {
		user, err = sv.UsersRepository.FindByEmail(info.Email)
		if err != nil {
			return nil, err
		}
	}

	if user == nil {
		// First-time sign-in: provision the account.
		newUser := entities.NewUser()
		newUser.Email = info.Email
		newUser.Name = info.Name
		newUser.Picture = info.Picture
		newUser.GoogleID = info.Sub
		newUser.Provider = "google"
		newUser.EmailVerified = info.EmailVerified == "true"
		newUser.LastLoginAt = now

		if err := sv.UsersRepository.InsertUser(newUser); err != nil {
			return nil, err
		}
		user = &newUser
	} else {
		// Returning user: refresh profile + login metadata.
		user.Name = info.Name
		user.Picture = info.Picture
		user.GoogleID = info.Sub
		user.Provider = "google"
		user.EmailVerified = info.EmailVerified == "true"
		user.LastLoginAt = now
		user.UpdatedAt = now

		if err := sv.UsersRepository.UpdateUser(user.ID, *user); err != nil {
			return nil, err
		}
	}

	return sv.issueAuthResponse(user)
}

// ===== Microsoft Sign-In =====

func (sv *usersService) MicrosoftSignIn(idToken string) (*entities.AuthResponse, error) {
	if idToken == "" {
		return nil, errors.New("id_token must not be empty")
	}

	info, err := sv.MicrosoftOAuthClient.VerifyIDToken(idToken)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	// Find an existing user by Microsoft subject, falling back to email so accounts
	// created through another flow get linked instead of duplicated.
	user, err := sv.UsersRepository.FindByMicrosoftID(info.Sub)
	if err != nil {
		return nil, err
	}
	if user == nil {
		user, err = sv.UsersRepository.FindByEmail(info.Email)
		if err != nil {
			return nil, err
		}
	}

	if user == nil {
		// First-time sign-in: provision the account.
		newUser := entities.NewUser()
		newUser.Email = info.Email
		newUser.Name = info.Name
		newUser.MicrosoftID = info.Sub
		newUser.Provider = "microsoft"
		newUser.EmailVerified = true // Microsoft verified account by default
		newUser.LastLoginAt = now

		if err := sv.UsersRepository.InsertUser(newUser); err != nil {
			return nil, err
		}
		user = &newUser
	} else {
		// Returning user: refresh profile + login metadata.
		user.Name = info.Name
		user.MicrosoftID = info.Sub
		user.Provider = "microsoft"
		user.EmailVerified = true
		user.LastLoginAt = now
		user.UpdatedAt = now

		if err := sv.UsersRepository.UpdateUser(user.ID, *user); err != nil {
			return nil, err
		}
	}

	return sv.issueAuthResponse(user)
}

// ===== CRUD =====

func (sv *usersService) CreateUser(data entities.UserDataModel) (*entities.UserDataModel, error) {
	if data.ID == "" {
		data.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	data.CreatedAt = now
	data.UpdatedAt = now

	if err := sv.validateUser(&data); err != nil {
		return nil, err
	}

	// Enforce email uniqueness.
	existing, err := sv.UsersRepository.FindByEmail(data.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("a user with this email already exists")
	}

	if err := sv.UsersRepository.InsertUser(data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (sv *usersService) GetUserByID(id string) (*entities.UserDataModel, error) {
	if id == "" {
		return nil, errors.New("id must not be empty")
	}
	return sv.UsersRepository.FindByID(id)
}

func (sv *usersService) GetAllUsers(filter entities.UserFilter) (*[]entities.UserDataModel, error) {
	return sv.UsersRepository.FindByFilter(filter)
}

func (sv *usersService) UpdateUser(id string, data entities.UserDataModel) error {
	if id == "" {
		return errors.New("id must not be empty")
	}

	existing, err := sv.UsersRepository.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("user not found")
	}

	data.ID = id // ID cannot be changed
	data.UpdatedAt = time.Now().UTC()

	return sv.UsersRepository.UpdateUser(id, data)
}

func (sv *usersService) DeleteUser(id string) error {
	if id == "" {
		return errors.New("id must not be empty")
	}

	existing, err := sv.UsersRepository.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("user not found")
	}

	return sv.UsersRepository.DeleteUser(id)
}
