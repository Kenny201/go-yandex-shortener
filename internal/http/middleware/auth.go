package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

// Определение уникального типа для ключей контекста
type contextKey string

// UserIDContextKey Объявление ключей контекста как констант
const (
	UserIDContextKey contextKey = "user_id"
)

func NewTokenMiddleware(userID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				jwtSecret := viper.GetString("JWT_SECRET")
				token, err := generateAuthToken(userID, jwtSecret)
				if err != nil {
					slog.Error("Failed to generate auth token", slog.String("error", err.Error()))
					http.Error(w, `{"error":"Failed to generate auth token"}`, http.StatusInternalServerError)
					return
				}

				// Устанавливаем куку с токеном
				http.SetCookie(w, &http.Cookie{
					Name:     "auth_token",
					Value:    token,
					Path:     "/",
					HttpOnly: true,
					SameSite: http.SameSiteStrictMode,
					Secure:   true,
				})

				// Устанавливаем заголовок Authorization
				w.Header().Set("Authorization", "Bearer "+token)
				slog.Info("Generated new auth token for POST request", slog.String("userID", userID))

				// Добавляем userID в контекст и продолжаем выполнение запроса
				ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Если не POST-запрос, просто продолжаем выполнение
			next.ServeHTTP(w, r)
		})
	}
}

func AuthCheckMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// Пытаемся получить токен из куки, если заголовка нет
				cookie, err := r.Cookie("auth_token")
				if err == nil && cookie.Value != "" {
					authHeader = "Bearer " + cookie.Value
				} else {
					slog.Warn("Missing Authorization header and auth_token cookie")
					http.Error(w, "", http.StatusUnauthorized)
					return
				}
			}

			// Парсинг токена из заголовка Authorization
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			jwtSecret := viper.GetString("JWT_SECRET")

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				slog.Warn("Invalid token detected")
				http.Error(w, "", http.StatusUnauthorized)
				return
			}

			// Извлечение userID из токена и добавление в контекст
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				userID, ok := claims["user_id"].(string)
				if !ok || userID == "" {
					slog.Warn("Token claims missing or invalid userID")
					http.Error(w, "", http.StatusUnauthorized)
					return
				}
				slog.Info("Valid token found", slog.String("userID", userID))
				ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				slog.Warn("Invalid token claims")
				http.Error(w, "", http.StatusUnauthorized)
				return
			}
		})
	}
}

func generateAuthToken(userID, secret string) (string, error) {
	// Создание нового токена
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Устанавливаем срок действия токена на 24 часа
	})

	// Подписываем токен
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
