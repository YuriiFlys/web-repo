package service

import (
	"context"
	"errors"
	"strings"

	"project-management/internal/auth"
	"project-management/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrEmailInUse = errors.New("email already in use")

type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (model.User, string, error)
	Login(ctx context.Context, input LoginInput) (model.User, string, error)
	GetByID(ctx context.Context, id uint) (model.User, error)
}

type PasswordManager interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type TokenIssuer interface {
	Issue(user model.User) (string, error)
}

type AuthRepository interface {
	FindByEmail(ctx context.Context, email string) (model.User, error)
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uint) (model.User, error)
}

type passwordManager struct{}

func (passwordManager) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (passwordManager) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

type jwtTokenIssuer struct{}

func (jwtTokenIssuer) Issue(user model.User) (string, error) {
	return auth.IssueToken(user)
}

type authService struct {
	repo   AuthRepository
	hasher PasswordManager
	tokens TokenIssuer
}

func NewAuthService(repo AuthRepository) AuthService {
	return &authService{repo: repo, hasher: passwordManager{}, tokens: jwtTokenIssuer{}}
}

func NewAuthServiceWithDeps(repo AuthRepository, hasher PasswordManager, tokens TokenIssuer) AuthService {
	return &authService{repo: repo, hasher: hasher, tokens: tokens}
}

func (s *authService) Register(ctx context.Context, input RegisterInput) (model.User, string, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	name := strings.TrimSpace(input.Name)

	if _, err := s.repo.FindByEmail(ctx, email); err == nil {
		return model.User{}, "", ErrEmailInUse
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, "", err
	}

	hash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return model.User{}, "", err
	}

	user := model.User{Email: email, Name: name, PasswordHash: hash}
	if err := s.repo.Create(ctx, &user); err != nil {
		return model.User{}, "", err
	}

	token, err := s.tokens.Issue(user)
	if err != nil {
		return model.User{}, "", err
	}

	return user, token, nil
}

func (s *authService) Login(ctx context.Context, input LoginInput) (model.User, string, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return model.User{}, "", err
	}

	if err := s.hasher.Compare(user.PasswordHash, input.Password); err != nil {
		return model.User{}, "", err
	}

	token, err := s.tokens.Issue(user)
	if err != nil {
		return model.User{}, "", err
	}

	return user, token, nil
}

func (s *authService) GetByID(ctx context.Context, id uint) (model.User, error) {
	return s.repo.GetByID(ctx, id)
}
