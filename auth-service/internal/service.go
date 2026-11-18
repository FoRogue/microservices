package internal

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var (
	errInvalidRegistration = errors.New("username is required and password must be at least 6 characters")
	errUsernameTaken       = errors.New("username already taken")
	errHashPassword        = errors.New("failed to hash password")
	errInvalidCredentials  = errors.New("invalid credentials")
	errIssueToken          = errors.New("failed to issue token")
	errInvalidToken        = errors.New("invalid or expired token")
	errUserNotFound        = errors.New("user not found")
)

type Service struct {
	repo      *Repository
	jwtSecret []byte
}

func NewService(repo *Repository, jwtSecret string) *Service {
	return &Service{repo: repo, jwtSecret: []byte(jwtSecret)}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (User, error) {
	username := strings.TrimSpace(req.Username)
	if username == "" || len(req.Password) < 6 {
		return User{}, errInvalidRegistration
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, errHashPassword
	}
	user, err := s.repo.CreateUser(ctx, username, string(hash))
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return User{}, errUsernameTaken
	}
	return user, err
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (string, error) {
	user, passwordHash, err := s.repo.Credentials(ctx, req.Username)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errInvalidCredentials
	}
	if err != nil {
		return "", err
	}
	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)) != nil {
		return "", errInvalidCredentials
	}
	token, err := s.issueToken(user)
	if err != nil {
		return "", errIssueToken
	}
	return token, nil
}

func (s *Service) CurrentUser(ctx context.Context, tokenStr string) (User, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(*jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return User{}, errInvalidToken
	}

	username, _ := claims["username"].(string)
	user, err := s.repo.FindUser(ctx, username)
	if err != nil {
		return User{}, errUserNotFound
	}
	return user, nil
}

func (s *Service) issueToken(user User) (string, error) {
	claims := jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
