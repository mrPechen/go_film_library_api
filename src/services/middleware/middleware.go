package middleware

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"src/src/config"
	"strings"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем токен из заголовка Authorization
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Authorization token is required", http.StatusUnauthorized)
			return
		}
		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
		allowedRoles := []string{"admin", "user"}
		signingKey := []byte(config.EnvVariable("SIGN_KEY"))
		// Проверяем валидность access токена
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return signingKey, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid access token", http.StatusUnauthorized)
			log.Println(err)
			return
		}

		// Проверяем роль пользователя
		roleAllowed := false
		for _, allowedRole := range allowedRoles {
			if token.Claims.(jwt.MapClaims)["role"] == allowedRole {
				roleAllowed = true
				break
			}
		}

		if !roleAllowed {
			http.Error(w, "Unauthorized access", http.StatusForbidden)
			return
		}

		// Проверяем разрешения на методы
		if !checkMethodPermissions(r.Method, token.Claims.(jwt.MapClaims)["role"].(string)) {
			http.Error(w, "Method not allowed for the user role", http.StatusMethodNotAllowed)
			return
		}

		// Добавляем информацию о пользователе в контекст запроса
		ctx := context.WithValue(r.Context(), "user_id", token.Claims.(jwt.MapClaims)["user_id"])

		// Продолжаем обработку запроса с учетом контекста
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func checkMethodPermissions(method string, role string) bool {
	// Пользователи с ролью "admin" имеют доступ ко всем методам
	if role == "admin" {
		return true
	}

	// Пользователи с ролью "user" разрешено только GET
	if role == "user" && method != http.MethodGet {
		return false
	}

	return true
}
