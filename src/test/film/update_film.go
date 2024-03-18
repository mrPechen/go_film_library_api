package film

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"src/src/test/test_auth"
	"testing"
)

func UpdateFilm(t *testing.T) {
	requestBody := map[string]string{
		"name":        "film 1.1",
		"description": "desc film 1.1",
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	filmData := GetFilmData()
	film := filmData["id"]
	url := fmt.Sprintf("http://server:8080/api/film-update?id=%d", int(film.(float64)))

	// Выполнение POST запроса
	client := &http.Client{}
	request, err := http.NewRequest("PATCH", url, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	token := test_auth.GetTokenData()
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token["user@test.ru"].AccessToken)
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("Failed to make PATCH request: %v", err)
	}
	response.Body.Close()
	// Проверка статуса ответа
	assert.Equal(t, http.StatusMethodNotAllowed, response.StatusCode, "Expected status code 405, but got %d", response.StatusCode)

	requestTwo, err := http.NewRequest("PATCH", url, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	requestTwo.Header.Set("Content-Type", "application/json")
	requestTwo.Header.Set("Authorization", "Bearer "+token["admin@test.ru"].AccessToken)
	responseTwo, err := client.Do(requestTwo)
	if err != nil {
		t.Fatalf("Failed to make PATCH request: %v", err)
	}
	defer response.Body.Close()

	assert.Equal(t, http.StatusCreated, responseTwo.StatusCode, "Expected status code 201, but got %d", responseTwo.StatusCode)

	filmData["name"] = "film 1.1"
	filmData["description"] = "desc film 1.1"

	requestBodyJSON, err := json.Marshal(filmData)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	requestBodyBytes = []byte(requestBodyJSON)

	WriteFilmToFile(requestBodyBytes)
}
