package service

import (
	"context"
	"errors"
	"ticTacToe/internal/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	jwtProvider TokenProvider
	userService *UserService
}

var (
	ErrEmptyLogin         = errors.New("login cannot be empty")
	ErrEmptyPassword      = errors.New("password cannot be empty")
	ErrLoginTooShort      = errors.New("login is too short")
	ErrPasswordTooShort   = errors.New("password is too short")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAuthHeader  = errors.New("invalid authorization header")
)

func NewAuthService(
	provider TokenProvider,
	userService *UserService,
) *AuthService {
	return &AuthService{
		jwtProvider: provider,
		userService: userService,
	}
}

func (s *AuthService) Registration(ctx context.Context,
	req model.JwtRequest) (bool, error) {
	if err := validateSignUpRequest(req); err != nil {
		return false, err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost)

	if err != nil {
		return false, err
	}
	user := model.UserModel{
		ID:             uuid.NewString(),
		Login:          req.Login,
		HashedPassword: string(hashedPassword),
	}
	err = s.userService.Create(ctx, user)
	if err != nil {
		return false, err
	}

	return true, nil
}

func validateSignUpRequest(req model.JwtRequest) error {
	if req.Login == "" {
		return ErrEmptyLogin
	}

	if req.Password == "" {
		return ErrEmptyPassword
	}

	if len(req.Login) < 3 {
		return ErrLoginTooShort
	}

	if len(req.Password) < 6 {
		return ErrPasswordTooShort
	}

	return nil
}

func (s *AuthService) Login(req model.JwtRequest) (model.JwtResponse, error) {
	userFromdata, err := s.userService.GetByLogin(context.Background(), req.Login)
	if err != nil {
		return model.JwtResponse{}, ErrInvalidCredentials
	}
	err = bcrypt.CompareHashAndPassword(
		[]byte(userFromdata.HashedPassword),
		[]byte(req.Password),
	)
	if err != nil {
		return model.JwtResponse{}, ErrInvalidCredentials
	}
	UserForToken := User{ID: userFromdata.ID,
		Name: userFromdata.Login}

	accessToken, err := s.GenerateAccessToken(UserForToken)
	if err != nil {
		return model.JwtResponse{}, err
	}
	refreshToken, err := s.GenerateRefreshToken(UserForToken)
	if err != nil {
		return model.JwtResponse{}, err
	}

	return model.JwtResponse{AccessToken: accessToken,
		RefreshToken: refreshToken,
		Type:         "Bearer",
	}, nil

}

func (s *AuthService) RefreshAccessToken(req model.RefreshJwtRequest) (model.JwtResponse, error) {
	err := s.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return model.JwtResponse{}, err
	}
	userID, err := s.jwtProvider.GetUserIDFromRefreshToken(req.RefreshToken)
	if err != nil {
		return model.JwtResponse{}, err
	}
	userFromdata, err := s.userService.GetByID(context.Background(), userID)
	if err != nil {
		return model.JwtResponse{}, errors.New("user not found")
	}
	userForToken := User{ID: userFromdata.ID,
		Name: userFromdata.Login}
	accessToken, err := s.GenerateAccessToken(userForToken)
	if err != nil {
		return model.JwtResponse{}, err
	}

	return model.JwtResponse{AccessToken: accessToken,
		RefreshToken: req.RefreshToken,
		Type:         "Bearer",
	}, nil

}

func (s *AuthService) RefreshRefreshToken(req model.RefreshJwtRequest) (model.JwtResponse, error) {
	err := s.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return model.JwtResponse{}, err
	}
	userID, err := s.jwtProvider.GetUserIDFromRefreshToken(req.RefreshToken)
	if err != nil {
		return model.JwtResponse{}, err
	}
	userFromdata, err := s.userService.GetByID(context.Background(), userID)
	if err != nil {
		return model.JwtResponse{}, errors.New("user not found")
	}
	userForToken := User{ID: userFromdata.ID,
		Name: userFromdata.Login}
	accessToken, err := s.GenerateAccessToken(userForToken)
	if err != nil {
		return model.JwtResponse{}, err
	}

	refreshToken, err := s.GenerateRefreshToken(userForToken)
	if err != nil {
		return model.JwtResponse{}, err
	}

	return model.JwtResponse{AccessToken: accessToken,
		RefreshToken: refreshToken,
		Type:         "Bearer",
	}, nil
}

func (s *AuthService) GenerateAccessToken(user User) (string, error) {
	return s.jwtProvider.GenerateAccessToken(user)
}
func (s *AuthService) GenerateRefreshToken(user User) (string, error) {
	return s.jwtProvider.GenerateRefreshToken(user)
}

func (s *AuthService) ValidateAccessToken(token string) error {
	return s.jwtProvider.ValidateAccessToken(token)
}
func (s *AuthService) ValidateRefreshToken(token string) error {
	return s.jwtProvider.ValidateRefreshToken(token)
}

func (s *AuthService) GetUserID(token string) (string, error) {
	return s.jwtProvider.GetUserID(token)
}
func (s *AuthService) GetUserIDFromRefreshToken(token string) (string, error) {
	return s.jwtProvider.GetUserIDFromRefreshToken(token)
}
