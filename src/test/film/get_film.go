package film

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"src/src/test/test_auth"
	"testing"
)

func GetFilm(t *testing.T) {

	url := "http://server:8080/api/film?film_name=fil"

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
	// Чтение тела ответа
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	var responseBodyJSON []map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &responseBodyJSON); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}
	filmData := GetFilmData()
	assert.Equal(t, http.StatusOK, response.StatusCode, "Expected status code 200, but got %d", response.StatusCode)
	assert.Equal(t, filmData, responseBodyJSON[0], "Data from file film.json and response body should match")

}
