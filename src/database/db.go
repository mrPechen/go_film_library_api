package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"src/src/config"
)

func DBConnect() (*pgxpool.Pool, error) {

	pgUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		config.EnvVariable("POSTGRES_USER"),
		config.EnvVariable("POSTGRES_PASSWORD"),
		config.EnvVariable("POSTGRES_HOST"),
		config.EnvVariable("POSTGRES_PORT"),
		config.EnvVariable("POSTGRES_DB"),
	)

	// Создание конфигурации пула подключений
	pgConfig, err := pgxpool.ParseConfig(pgUrl)
	if err != nil {
		log.Fatalf("Unable to parse connection pool config: %v\n", err)
		return nil, err
	}

	// Установка параметров сеанса
	pgConfig.ConnConfig.RuntimeParams = map[string]string{
		"datestyle": "ISO, DMY",
	}

	// Создание пула подключений к базе данных
	dbpool, err := pgxpool.NewWithConfig(context.Background(), pgConfig)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
		return nil, err
	}

	// Создание таблиц
	if err := CreateTables(dbpool); err != nil {
		dbpool.Close()
		return nil, err
	}

	return dbpool, nil
}

// Функция создания таблиц
func CreateTables(dbpool *pgxpool.Pool) error {

	// Создание таблицы "User"
	_, err := dbpool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS "user" (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL CHECK (email ~* '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'),
			password VARCHAR(100) NOT NULL,
			role VARCHAR(50) NOT NULL CHECK (role IN ('user', 'admin'))
		);
	`)
	if err != nil {
		log.Println("Unable to create User table:", err)
		return err
	}

	// Создание таблицы "Actor"
	_, err = dbpool.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS actor (
            id SERIAL PRIMARY KEY,
            gender VARCHAR(50) NOT NULL,
            birth_date TIMESTAMP WITH TIME ZONE NOT NULL,
            name VARCHAR(255) NOT NULL
        );
    `)
	if err != nil {
		log.Println("Unable to create Actor table:", err)
		return err
	}

	// Создание таблицы "Film"
	_, err = dbpool.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS film (
            id SERIAL PRIMARY KEY,
            name VARCHAR(150) NOT NULL,
            description VARCHAR(1000) NOT NULL,
            release_date TIMESTAMP WITH TIME ZONE NOT NULL,
            rate INTEGER NOT NULL
        );
    `)
	if err != nil {
		log.Println("Unable to create Film table:", err)
		return err
	}

	_, err = dbpool.Exec(context.Background(), `
    	CREATE TABLE IF NOT EXISTS film_actor (
			film_id INTEGER,
			actor_id INTEGER,
			CONSTRAINT film_fk FOREIGN KEY (film_id) REFERENCES film(id) ON DELETE CASCADE,
			CONSTRAINT actor_fk FOREIGN KEY (actor_id) REFERENCES actor(id) ON DELETE CASCADE,
			CONSTRAINT film_actor_pk PRIMARY KEY (film_id, actor_id)
    	);
	`)

	if err != nil {
		log.Println("Unable to create film_actor table:", err)
		return err
	}

	return nil
}
