package actor

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"src/src/test/test_auth"
	"testing"
)

func GetActorsWithoutFilms(t *testing.T) {
	url := "http://server:8080/api/actors/"

	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	token := test_auth.GetTokenData()
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token["user@test.ru"].AccessToken)

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer response.Body.Close()

	// Проверка статуса ответа
	assert.Equal(t, http.StatusOK, response.StatusCode, "Expected status code 200, but got %d", response.StatusCode)

	// Чтение тела ответа
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Преобразуем тело ответа в массив интерфейсов
	var responseJSON []interface{}
	err = json.Unmarshal(bodyBytes, &responseJSON)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	// Проверяем, что ответ содержит хотя бы один элемент
	if len(responseJSON) == 0 {
		t.Fatal("Response body is empty")
	}

	// Получаем первый элемент массива
	actorData := responseJSON[0].(map[string]interface{})
	actor := actorData["actor"].(map[string]interface{})
	// Проверка соответствия всех полей кроме ID
	assert.Equal(t, "Tom Hanks", actor["name"].(string), "Actor name does not match expected value")
	assert.Equal(t, "male", actor["gender"].(string), "Actor gender does not match expected value")

	var elements []json.RawMessage
	if err := json.Unmarshal(bodyBytes, &elements); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	WriteActorToFile(elements[0])
}
