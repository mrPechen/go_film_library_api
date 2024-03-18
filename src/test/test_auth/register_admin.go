package test_auth

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func RegisterAdmin(t *testing.T) {
	// Создание тела запроса
	requestBodyUser := map[string]string{
		"email":    "admin@test.ru",
		"password": "123",
		"role":     "admin",
	}

	requestBodyBytes, err := json.Marshal(requestBodyUser)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	url := "http://server:8080/api/registration/"

	// Выполнение POST запроса
	response, err := http.Post(url, "application/json", bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		t.Fatalf("Failed to make POST request: %v", err)
	}
	defer response.Body.Close()

	// Проверка статуса ответа
	assert.Equal(t, http.StatusOK, response.StatusCode, "Expected status code 200, but got %d", response.StatusCode)

	// Проверка наличия ключей в теле ответа
	var responseBody map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&responseBody)
	assert.NoError(t, err, "Failed to decode response body: %v", err)

	_, accessTokenExists := responseBody["access_token"]
	assert.True(t, accessTokenExists, "Expected 'access_token' key in response body, but not found")

	_, refreshTokenExists := responseBody["refresh_token"]
	assert.True(t, refreshTokenExists, "Expected 'refresh_token' key in response body, but not found")
}
