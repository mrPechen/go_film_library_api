package actor

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"src/src/test/test_auth"
	"testing"
)

func CreateActor(t *testing.T) {
	requestBody := map[string]string{
		"name":       "Tom Hanks",
		"gender":     "male",
		"birth_date": "1975-10-02T00:00:00Z",
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	url := "http://server:8080/api/actor-create/"

	// Выполнение POST запроса
	client := &http.Client{}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	token := test_auth.GetTokenData()
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token["user@test.ru"].AccessToken)
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("Failed to make POST request: %v", err)
	}
	response.Body.Close()
	// Проверка статуса ответа
	assert.Equal(t, http.StatusMethodNotAllowed, response.StatusCode, "Expected status code 405, but got %d", response.StatusCode)

	requestTwo, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	requestTwo.Header.Set("Content-Type", "application/json")
	requestTwo.Header.Set("Authorization", "Bearer "+token["admin@test.ru"].AccessToken)
	responseTwo, err := client.Do(requestTwo)
	if err != nil {
		t.Fatalf("Failed to make POST request: %v", err)
	}
	defer response.Body.Close()

	assert.Equal(t, http.StatusCreated, responseTwo.StatusCode, "Expected status code 201, but got %d", responseTwo.StatusCode)

	WriteActorToFile(requestBodyBytes)
}
