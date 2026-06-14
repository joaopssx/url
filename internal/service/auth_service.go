package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"urlshortener/internal/model"
	"urlshortener/internal/repository"
)

type AuthService interface {
	Register(email, password string) (*model.User, error)
	Login(email, password string) (string, time.Time, error)
	ParseToken(tokenString string) (*jwt.MapClaims, error)
}

type authService struct {
	userRepo  repository.UserRepository
	jwtSecret []byte
}

func NewAuthService(userRepo repository.UserRepository, secret string) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: []byte(secret),
	}
}

func (s *authService) Register(email, password string) (*model.User, error) {
	existingUser, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now().UTC(),
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(email, password string) (string, time.Time, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", time.Time{}, err
	}
	if user == nil {
		return "", time.Time{}, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", time.Time{}, errors.New("invalid credentials")
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"exp":   expirationTime.Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}

func (s *authService) ParseToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, errors.New("invalid token")
}
