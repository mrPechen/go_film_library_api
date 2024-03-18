package film

import (
	"encoding/json"
	"log"
	"os"
)

func WriteFilmToFile(body []byte) {

	// Открываем файл для записи (существующий файл будет перезаписан)
	file, err := os.OpenFile("json/film.json", os.O_WRONLY|os.O_TRUNC, 0644)
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

func GetFilmData() map[string]interface{} {
	fileContent, err := os.ReadFile("json/film.json")
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
