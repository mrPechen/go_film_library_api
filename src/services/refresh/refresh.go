package refresh

import (
	"context"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"src/src/config"
	"src/src/database"
	jwt_dir "src/src/services/jwt"
	"time"
)

// @Summary Обновление токенов доступа и обновления
// @Description Обновляет токены доступа и обновления пользователя.
// @Tags authentication
// @Accept json
// @Produce json
// @Param refresh_token query string true "Обновляющий токен"
// @Success 200 "Успешное обновление токенов"
// @Failure 400 "Отсутствует обязательный параметр 'refresh_token'"
// @Failure 401 "Неверный обновляющий токен"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /api/refresh-tokens/ [post]
func RefreshTokens(w http.ResponseWriter, r *http.Request, dbpool *pgxpool.Pool) {
	var (
		signingKey   = []byte(config.EnvVariable("SIGN_KEY"))
		refreshKey   = []byte(config.EnvVariable("REFRESH_KEY"))
		tokenExpires = time.Minute * 30
	)

	// Парсим данные запроса
	var reqBody struct {
		RefreshToken string `json:"refresh_token"`
	}
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println("Error decoding request body:", err)
		return
	}

	// Получаем refresh token из запроса
	refreshToken := reqBody.RefreshToken
	if refreshToken == "" {
		http.Error(w, "Refresh token is required", http.StatusBadRequest)
		return
	}

	// Проверяем валидность refresh token'а
	user, err := ValidateRefreshToken(refreshToken, refreshKey, dbpool)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		log.Println("Error validating refresh token:", err)
		return
	}

	// Создаем новые токены для пользователя
	accessToken, accessTokenErr := jwt_dir.CreateAccessToken(user, signingKey, tokenExpires)
	if accessTokenErr != nil {
		http.Error(w, accessTokenErr.Error(), http.StatusInternalServerError)
		log.Println("Error creating access token:", accessTokenErr)
		return
	}

	newRefreshToken, newRefreshTokenErr := jwt_dir.CreateRefreshToken(user, refreshKey)
	if newRefreshTokenErr != nil {
		http.Error(w, newRefreshTokenErr.Error(), http.StatusInternalServerError)
		log.Println("Error creating refresh token:", newRefreshTokenErr)
		return
	}

	// Формируем JSON с новыми токенами и отправляем клиенту
	tokens := database.Tokens{AccessToken: accessToken, RefreshToken: newRefreshToken}
	jsonTokens, marshalErr := json.Marshal(tokens)
	if marshalErr != nil {
		http.Error(w, marshalErr.Error(), http.StatusInternalServerError)
		log.Println("Error marshaling tokens to JSON:", marshalErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonTokens)
}

func ValidateRefreshToken(tokenString string, refreshKey []byte, dbpool *pgxpool.Pool) (*database.User, error) {
	// Проверяем валидность refresh token'а и получаем его данные
	claims := jwt.MapClaims{}
	refreshToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return refreshKey, nil
	})
	if err != nil || !refreshToken.Valid {
		return nil, err
	}

	// Получаем пользователя из базы данных по идентификатору из токена
	var user database.User
	err = dbpool.QueryRow(context.Background(), "SELECT id, email, role FROM \"user\" WHERE id = $1", claims["user_id"]).Scan(&user.ID, &user.Email, &user.Role)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
