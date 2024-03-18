package actor

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"src/src/test/test_auth"
	"testing"
)

func DeleteActor(t *testing.T) {
	actorData := GetActorData()
	actor, ok := actorData["actor"].(map[string]interface{})
	if !ok {
		fmt.Println("Actor data not found")
		return
	}
	url := fmt.Sprintf("http://server:8080/api/actor-delete?id=%d", int(actor["id"].(float64)))

	// Выполнение POST запроса
	client := &http.Client{}
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	token := test_auth.GetTokenData()
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token["user@test.ru"].AccessToken)
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("Failed to make DELETE request: %v", err)
	}
	response.Body.Close()
	// Проверка статуса ответа
	assert.Equal(t, http.StatusMethodNotAllowed, response.StatusCode, "Expected status code 405, but got %d", response.StatusCode)

	requestTwo, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	requestTwo.Header.Set("Content-Type", "application/json")
	requestTwo.Header.Set("Authorization", "Bearer "+token["admin@test.ru"].AccessToken)
	responseTwo, err := client.Do(requestTwo)
	if err != nil {
		t.Fatalf("Failed to make DELETE request: %v", err)
	}
	defer response.Body.Close()

	assert.Equal(t, http.StatusNoContent, responseTwo.StatusCode, "Expected status code 204, but got %d", responseTwo.StatusCode)

}
