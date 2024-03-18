package registration

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"src/src/config"
	"src/src/database"
	"src/src/services/jwt"
	"src/src/services/login"
	"strings"
	"time"
)

// @Summary Регистрация пользователя
// @Description Регистрирует нового пользователя в системе и выдает ему токены доступа
// @Tags authentication
// @Accept json
// @Produce json
// @Param email formData string true "Email пользователя"
// @Param password formData string true "Пароль пользователя"
// @Param role formData string true "Роль пользователя ('user' или 'admin')"
// @Success 200 "Успешная регистрация и вход"
// @Failure 400 "Некорректные данные для регистрации"
// @Failure 500 "Ошибка сервера"
// @Router /api/registration/ [post]
func RegisterUser(w http.ResponseWriter, r *http.Request, dbpool *pgxpool.Pool) {
	// Парсим данные запроса
	var newUser database.User
	decodeErr := json.NewDecoder(r.Body).Decode(&newUser)
	if decodeErr != nil {
		http.Error(w, decodeErr.Error(), http.StatusBadRequest)
		log.Printf("Error decoding request body: %v", decodeErr)
		return
	}

	// Проверяем, существует ли пользователь с таким email в базе данных
	exists, existsErr := login.UserExists(newUser.Email, dbpool)
	if existsErr != nil {
		http.Error(w, "Failed to check user existence", http.StatusInternalServerError)
		log.Printf("Error checking user existence: %v", existsErr)
		return
	}

	if exists {
		http.Error(w, "User with this email already exists", http.StatusBadRequest)
		log.Println("User with this email already exists")
		return
	}

	// Хешируем пароль
	hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if hashErr != nil {
		http.Error(w, hashErr.Error(), http.StatusInternalServerError)
		log.Printf("Error hashing password: %v", hashErr)
		return
	}

	_, insertErr := dbpool.Exec(context.Background(), `
        INSERT INTO "user" (email, password, role)
        VALUES ($1, $2, $3)
    `, newUser.Email, string(hashedPassword), newUser.Role)
	if insertErr != nil {
		if strings.Contains(insertErr.Error(), "violates check constraint \"user_role_check\"") {
			http.Error(w, "Invalid user role", http.StatusBadRequest)
			log.Println("Error: Invalid user role")
			return
		}
		if strings.Contains(insertErr.Error(), "violates check constraint \"user_email_check\"") {
			http.Error(w, "Invalid email format", http.StatusBadRequest)
			log.Println("Error: Invalid email format")
			return
		}
		http.Error(w, insertErr.Error(), http.StatusInternalServerError)
		log.Printf("Error inserting user into database: %v", insertErr)
		return
	}

	log.Printf("User with email %s registered successfully", newUser.Email)

	// Создаем access и refresh токены для нового пользователя
	var (
		signingKey   = []byte(config.EnvVariable("SIGN_KEY"))
		refreshKey   = []byte(config.EnvVariable("REFRESH_KEY"))
		tokenExpires = time.Minute * 30
	)

	// Создаем access и refresh токены для нового пользователя
	accessToken, accessTokenErr := jwt.CreateAccessToken(&newUser, signingKey, tokenExpires)
	if accessTokenErr != nil {
		http.Error(w, accessTokenErr.Error(), http.StatusInternalServerError)
		log.Println("Error creating access token:", accessTokenErr)
		return
	}

	refreshToken, refreshTokenErr := jwt.CreateRefreshToken(&newUser, refreshKey)
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
