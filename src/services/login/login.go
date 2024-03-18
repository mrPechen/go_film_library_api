package login

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"src/src/config"
	"src/src/database"
	"src/src/services/jwt"
	"time"
)

// @Summary Вход пользователя
// @Description Аутентификация пользователя по электронной почте и паролю.
// @Tags authentication
// @Accept json
// @Produce json
// @Param email formData string true "Email пользователя"
// @Param password formData string true "Пароль пользователя"
// @Success 200 "Успешная аутентификация и выдача токенов"
// @Failure 400 "Ошибка декодирования запроса"
// @Failure 404 "Пользователь не найден"
// @Failure 500 "Ошибка сервера"
// @Router /api/login/ [post]
func Login(w http.ResponseWriter, r *http.Request, dbpool *pgxpool.Pool) {

	var (
		signingKey   = []byte(config.EnvVariable("SIGN_KEY"))
		refreshKey   = []byte(config.EnvVariable("REFRESH_KEY"))
		tokenExpires = time.Minute * 30
	)

	// Парсим данные запроса
	var creds database.User
	decodeErr := json.NewDecoder(r.Body).Decode(&creds)
	if decodeErr != nil {
		http.Error(w, decodeErr.Error(), http.StatusBadRequest)
		log.Println("Error decoding request body:", decodeErr)
		return
	}

	// Проверяем, существует ли пользователь в базе данных
	exists, existsErr := UserExists(creds.Email, dbpool)
	if existsErr != nil {
		http.Error(w, "Failed to check user existence", http.StatusInternalServerError)
		log.Println("Error checking user existence:", existsErr)
		return
	}
	if !exists {
		http.Error(w, "User does not exist", http.StatusNotFound)
		log.Println("User does not exist")
		return
	}

	// Получаем пользователя из базы данных
	var user database.User
	err := dbpool.QueryRow(context.Background(), "SELECT id, email, password, role FROM \"user\" WHERE email = $1", creds.Email).
		Scan(&user.ID, &user.Email, &user.Password, &user.Role)
	if err != nil {
		http.Error(w, "Failed to fetch user from database", http.StatusInternalServerError)
		log.Println("Failed to fetch user from database:", err)
		return
	}

	// Создаем access и refresh токены для пользователя
	accessToken, accessTokenErr := jwt.CreateAccessToken(&user, signingKey, tokenExpires)
	if accessTokenErr != nil {
		http.Error(w, accessTokenErr.Error(), http.StatusInternalServerError)
		log.Println("Error creating access token:", accessTokenErr)
		return
	}

	refreshToken, refreshTokenErr := jwt.CreateRefreshToken(&user, refreshKey)
	if refreshTokenErr != nil {
		http.Error(w, refreshTokenErr.Error(), http.StatusInternalServerError)
		log.Println("Error creating refresh token:", refreshTokenErr)
		return
	}

	// Формируем JSON с токенами и отправляем клиенту
	tokens := database.Tokens{AccessToken: accessToken, RefreshToken: refreshToken}
	jsonTokens, marshalErr := json.Marshal(tokens)
	if marshalErr != nil {
		http.Error(w, marshalErr.Error(), http.StatusInternalServerError)
		log.Println("Error marshaling tokens to JSON:", marshalErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonTokens)
}

func UserExists(email string, dbpool *pgxpool.Pool) (bool, error) {
	var exists bool
	err := dbpool.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM \"user\" WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
