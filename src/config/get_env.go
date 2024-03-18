package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

// Функция получения данных из файла .env
func EnvVariable(key string) string {

	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error loading .env file, %v", err)
	}

	return os.Getenv(key)
}
