package test_auth

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// Tokens представляет токены доступа и обновления
type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func WriteTokensToFile(email string, tokens Tokens) {
	// Сначала читаем содержимое файла
	fileContent, err := ioutil.ReadFile("json/tokens.json")
	if err != nil {
		log.Fatalf("Failed to read tokens file: %v", err)
	}

	// Декодируем содержимое файла в map
	var tokensData map[string]Tokens
	if len(fileContent) > 0 {
		if err := json.Unmarshal(fileContent, &tokensData); err != nil {
			log.Fatalf("Failed to unmarshal tokens data from file: %v", err)
		}
	} else {
		tokensData = make(map[string]Tokens)
	}

	// Добавляем новые токены
	tokensData[email] = tokens

	// Кодируем обновленные данные обратно в JSON
	updatedTokensJSON, err := json.Marshal(tokensData)
	if err != nil {
		log.Fatalf("Failed to marshal updated tokens to JSON: %v", err)
	}

	// Создаем или перезаписываем файл "tokens.json"
	file, err := os.Create("json/tokens.json")
	if err != nil {
		log.Fatalf("Failed to create tokens file: %v", err)
	}
	defer file.Close()

	// Записываем данные в файл
	if _, err := file.Write(updatedTokensJSON); err != nil {
		log.Fatalf("Failed to write tokens to file: %v", err)
	}
}

func GetTokenData() map[string]Tokens {
	// Читаем содержимое файла JSON
	fileContent, err := ioutil.ReadFile("json/tokens.json")
	if err != nil {
		log.Println("Reading file error:", err)
		return nil
	}

	// Создаем переменную для хранения данных из файла JSON
	var data map[string]Tokens

	// Распаковываем данные из файла JSON в структуру данных
	if err := json.Unmarshal(fileContent, &data); err != nil {
		log.Println("Unmarshal error:", err)
		return nil
	}

	return data
}
