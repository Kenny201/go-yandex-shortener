package middleware

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type contextKey string

const (
	UserIDContextKey contextKey = "user_id"
)

// AuthMiddleware создает новый токен и устанавливает его в заголовок Authorization и в куку, если токен отсутствует или недействителен.
func AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			jwtSecret := viper.GetString("JWT_SECRET")

			userID, err := validateAuthTokenFromRequest(r, jwtSecret)
			if err != nil {
				// Создаем новый токен
				userID = uuid.New().String()
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
			} else {
				slog.Info("Valid token found", slog.String("userID", userID))
			}

			// Добавляем userID в контекст и продолжаем выполнение запроса
			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthCheckMiddleware проверяет наличие и валидность токена в заголовке Authorization или куке и передает запрос дальше.
func AuthCheckMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			jwtSecret := viper.GetString("JWT_SECRET")

			userID, err := validateAuthTokenFromRequest(r, jwtSecret)
			if err != nil {
				slog.Warn("Missing or invalid token")
				http.Error(w, "", http.StatusUnauthorized)
				return
			}

			// Добавляем userID в контекст и продолжаем выполнение запроса
			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// validateAuthTokenFromRequest извлекает и проверяет токен из заголовка Authorization или куки.
func validateAuthTokenFromRequest(r *http.Request, secret string) (string, error) {
	// Извлечение токена из заголовка Authorization
	authHeader := r.Header.Get("Authorization")
	var tokenStr string

	if authHeader != "" {
		tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		// Пытаемся получить токен из куки, если заголовка нет
		cookie, err := r.Cookie("auth_token")
		if err == nil && cookie.Value != "" {
			tokenStr = cookie.Value
		} else {
			return "", errors.New("missing token in both Authorization header and auth_token cookie")
		}
	}

	return validateAuthToken(tokenStr, secret)
}

// validateAuthToken проверяет валидность токена и извлекает userID.
func validateAuthToken(tokenStr, secret string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			return "", errors.New("userID missing in token claims")
		}
		return userID, nil
	}

	return "", errors.New("invalid token claims")
}

// generateAuthToken создает новый токен с userID.
func generateAuthToken(userID, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Устанавливаем срок действия токена на 24 часа
	})

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
