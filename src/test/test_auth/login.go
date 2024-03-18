package test_auth

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func Login(email string, t *testing.T) {
	// Создание тела запроса
	requestBody := map[string]string{
		"email":    email,
		"password": "123",
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	url := "http://server:8080/api/login/"

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

	tokens := Tokens{
		AccessToken:  responseBody["access_token"].(string),
		RefreshToken: responseBody["refresh_token"].(string),
	}

	WriteTokensToFile(email, tokens)

}
