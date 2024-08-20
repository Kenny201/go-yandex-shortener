package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"

	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/auth/entity"
	"github.com/Kenny201/go-yandex-shortener.git/internal/utils/secret"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
)

// Константы для ключей JWT claims
const (
	UserIDClaimKey = "user_id"
	ExpClaimKey    = "exp"
)

// UserRepository определяет интерфейс для работы с хранилищем сокращённых ссылок.
type UserRepository interface {
	GetByEmail(email string) (*entity.User, error)
	Create(username, email, password string) (*entity.User, error)
}

// Auth представляет собой основной сервис для работы с сокращёнными ссылками.
type Auth struct {
	userRepository UserRepository
	SecretKey      []byte
}

// New создает новый экземпляр сервиса Auth с заданным репозиторием.
func New(userRepository UserRepository) *Auth {
	secretToken := []byte(viper.GetString("JWT_SECRET"))
	return &Auth{userRepository: userRepository, SecretKey: secretToken}
}

func (a *Auth) Register(username, email, password string) (*entity.User, error) {
	user, err := a.userRepository.GetByEmail(email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}
	if user != nil {
		return nil, errors.New("user already exists")
	}

	hashedPassword, err := secret.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user, err = a.userRepository.Create(username, email, hashedPassword)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (a *Auth) Login(email string, password string) (*entity.User, error) {
	user, err := a.userRepository.GetByEmail(email)

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (a *Auth) GenerateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		UserIDClaimKey: userID.String(),
		ExpClaimKey:    time.Now().Add(1 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(a.SecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
