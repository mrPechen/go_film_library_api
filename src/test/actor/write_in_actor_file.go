package actor

import (
	"encoding/json"
	"log"
	"os"
)

type Actor struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Gender    string `json:"gender"`
	BirthDate string `json:"birth_date"`
}

type FilmForActor struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ReleaseDate string `json:"release_date"`
	Rate        int    `json:"rate"`
}

func WriteActorToFile(body []byte) {

	// Открываем файл для записи (существующий файл будет перезаписан)
	file, err := os.OpenFile("json/actor.json", os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Записываем данные в файл
	_, err = file.Write(body)
	if err != nil {
		log.Fatal(err)
	}

}

func GetActorData() map[string]interface{} {
	fileContent, err := os.ReadFile("json/actor.json")
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Декодируем данные из файла
	var data map[string]interface{}
	if err := json.Unmarshal(fileContent, &data); err != nil {
		log.Fatalf("Failed to unmarshal data from file: %v", err)
	}

	return data
}
