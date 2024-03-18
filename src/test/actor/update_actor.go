package actor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"src/src/test/test_auth"
	"testing"
)

func UpdateActor(t *testing.T) {
	requestBody := map[string]string{
		"name":       "Zak Galifianakis",
		"birth_date": "1969-01-02T00:00:00Z",
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	actorData := GetActorData()
	actor, ok := actorData["actor"].(map[string]interface{})
	if !ok {
		fmt.Println("Actor data not found")
		return
	}
	url := fmt.Sprintf("http://server:8080/api/actor-update?id=%d", int(actor["id"].(float64)))

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

	requestBodyTwo := map[string]interface{}{
		"actor": map[string]interface{}{
			"id":         actor["id"],
			"name":       "Zak Galifianakis",
			"gender":     "male",
			"birth_date": "1969-01-02T00:00:00Z",
		},
		"films": []interface{}{},
	}
	requestTwoBodyBytes, err := json.Marshal(requestBodyTwo)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	WriteActorToFile(requestTwoBodyBytes)
}
