package film

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"src/src/test/test_auth"
	"testing"
)

func GetFilms(t *testing.T) {
	url := "http://server:8080/api/films?sort_by=release_date"

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
	var films []map[string]interface{}
	UnmarshalErr := json.Unmarshal(bodyBytes, &films)
	if UnmarshalErr != nil {
		fmt.Println("Error:", err)
		return
	}

	// Получаем первый элемент массива
	filmDataResponse := films[0]
	actorDataResponse := filmDataResponse["actors"].([]interface{})
	actor := actorDataResponse[0].(map[string]interface{})
	filmDataFile := GetFilmData()
	actorIdInFilm := filmDataFile["actor_ids"].([]interface{})
	// Проверка соответствия всех полей кроме ID
	assert.Equal(t, filmDataFile["name"].(string), filmDataResponse["name"], "Film name does not match expected value")
	assert.Equal(t, filmDataFile["description"].(string), filmDataResponse["description"], "Film description does not match expected value")
	assert.Equal(t, actorIdInFilm[0].(float64), actor["id"], "Actor id does not match expected value")

	var elements []json.RawMessage
	if err := json.Unmarshal(bodyBytes, &elements); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	WriteFilmToFile(elements[0])
}
