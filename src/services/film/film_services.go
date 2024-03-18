package film

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"src/src/database"
	"strconv"
	"strings"
)

// @Summary Создание нового фильма
// @Description Создание нового фильма
// @Tags film
// @Accept json
// @Param request body database.Film true "Данные для создания фильма"
// @Success 201 {string} string "Film created successfully"
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
// @Router /api/film-create/ [post]
func CreateFilm(w http.ResponseWriter, r *http.Request, dbpool *pgxpool.Pool) {
	var newFilm database.Film
	if err := json.NewDecoder(r.Body).Decode(&newFilm); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var filmID int
	// Выполняем запрос INSERT для создания нового фильма
	insertErr := dbpool.QueryRow(context.Background(), `
		INSERT INTO film (name, description, release_date, rate) VALUES ($1, $2, $3, $4) RETURNING id
	`, newFilm.Name, newFilm.Description, newFilm.ReleaseDate, newFilm.Rate).Scan(&filmID)
	if insertErr != nil {
		http.Error(w, "Failed to create film", http.StatusInternalServerError)
		log.Println("Error creating film:", insertErr)
		return
	}

	// Добавляем записи в таблицу film_actor для каждого актера
	for _, actorID := range newFilm.ActorIDs {
		_, addErr := dbpool.Exec(context.Background(), `
			INSERT INTO film_actor (film_id, actor_id) VALUES ($1, $2)
		`, filmID, actorID)
		if addErr != nil {
			http.Error(w, "Failed to add actor to film", http.StatusInternalServerError)
			log.Printf("Error adding actor to film: %v\n", addErr)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func searchFunc(dbpool *pgxpool.Pool, searchFragment string, searchBy string) ([]database.FilmWithActors, error) {
	var films []database.FilmWithActors
	var query string

	switch searchBy {
	case "film":
		query = `
			SELECT f.id, f.name, f.description, f.release_date, f.rate, 
			   CASE WHEN COUNT(fa.actor_id) > 0
					THEN json_agg(json_build_object('id', a.id, 'name', a.name, 'gender', a.gender, 'birth_date', a.birth_date))
					ELSE NULL
			   END AS actors
            FROM film f
			LEFT JOIN film_actor fa ON f.id = fa.film_id
			LEFT JOIN actor a ON a.id = fa.actor_id
            WHERE LOWER(f.name) LIKE LOWER($1)
            GROUP BY f.id
        `
	case "actor":
		query = `
            SELECT f.id, f.name, f.description, f.release_date, f.rate, 
			   COALESCE(json_agg(json_build_object('id', a.id, 'name', a.name, 'gender', a.gender, 'birth_date', a.birth_date)), '[]') AS actors
            FROM film f
            JOIN film_actor fa ON f.id = fa.film_id
            JOIN actor a ON a.id = fa.actor_id
            WHERE LOWER(a.name) LIKE LOWER($1)
            GROUP BY f.id
        `
	default:
		return nil, errors.New("Invalid search type")
	}

	rows, err := dbpool.Query(context.Background(), query, "%"+searchFragment+"%")
	if err != nil {
		log.Printf("Error executing query: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var film database.FilmWithActors
		var actorsJSON []byte
		if scanErr := rows.Scan(&film.ID, &film.Name, &film.Description, &film.ReleaseDate, &film.Rate, &actorsJSON); scanErr != nil {
			log.Printf("Error scanning film information: %v\n", scanErr)
			return nil, scanErr
		}
		if actorsJSON == nil {
			actorsJSON = []byte("[]")
		}
		if unmarshalErr := json.Unmarshal(actorsJSON, &film.Actors); unmarshalErr != nil {
			log.Printf("Error unmarshalling actors JSON: %v\n", unmarshalErr)
			return nil, unmarshalErr
		}
		films = append(films, film)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v\n", err)
		return nil, err
	}

	return films, nil
}

// @Summary Поиск фильма по названию фильма или имени актера
// @Description Поиск фильма по названию фильма или имени актера
// @Tags film
// @Accept json
// @Produce json
// @Param film_name query string false "Фрагмент названия фильма"
// @Param actor_name query string false "фрагмент имени актера"
// @Success 200 {array} database.FilmWithActors "Фильм"
// @Failure 400 {string} string "Отсутствует обязательный параметр 'id'"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /api/film [get]
func SearchFilm(w http.ResponseWriter, r *http.Request, dbpool *pgxpool.Pool) {
	// Получаем параметры запроса
	queryParams := r.URL.Query()
	filmNameFragment := queryParams.Get("film_name")
	actorNameFragment := queryParams.Get("actor_name")

	var films []database.FilmWithActors

	// Если передан фрагмент имени фильма, ищем фильмы по имени фильма
	if filmNameFragment != "" {
		filmsByFilm, filmErr := searchFunc(dbpool, filmNameFragment, "film")
		if filmErr != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Printf("Error searching films by film name: %v\n", filmErr)
			return
		}
		films = append(films, filmsByFilm...)
	}

	// Если передан фрагмент имени актера, ищем фильмы по имени актера
	if actorNameFragment != "" {
		filmsByActor, actorErr := searchFunc(dbpool, actorNameFragment, "actor")
		if actorErr != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Printf("Error searching films by actor name: %v\n", actorErr)
			return
		}
		films = append(films, filmsByActor...)
	}

	// Отправляем результаты клиенту
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(films); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Error encoding films: %v\n", err)
		return
	}
}

// @Summary Получить все фильмы
// @Description Получить все фильмы
// @Tags film
// @Accept json
// @Produce json
// @Param sort_by query string false "Значения sort_by (name, release_date, rate)"
// @Success 200 {array} database.FilmWithActors "Список фильмов"
// @Failure 400 {string} string "Отсутствует обязательный параметр 'id'"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /api/films [get]
func GetAllFilms(w http.ResponseWriter, r *http.Request, dbpool *pgxpool.Pool) {
	// Получаем параметры запроса
	queryValues := r.URL.Query()
	sortBy := queryValues.Get("sort_by")

	// Проверяем корректность параметра sort_by
	validSortOptions := map[string]bool{
		"name":         true,
		"release_date": true,
		"rate":         true,
	}
	if !validSortOptions[sortBy] {
		http.Error(w, "Invalid sort_by parameter", http.StatusBadRequest)
		return
	}

	// Формируем запрос к базе данных в зависимости от параметра sort_by
	query := fmt.Sprintf(`
		SELECT f.id, f.name, f.description, f.release_date, f.rate, 
			   CASE WHEN COUNT(fa.actor_id) > 0
					THEN json_agg(json_build_object('id', a.id, 'name', a.name, 'gender', a.gender, 'birth_date', a.birth_date))
					ELSE NULL
			   END AS actors
		FROM film f
		LEFT JOIN film_actor fa ON f.id = fa.film_id
		LEFT JOIN actor a ON a.id = fa.actor_id
		GROUP BY f.id
		ORDER BY %s DESC
	`, sortBy)

	// Выполняем запрос к базе данных
	rows, err := dbpool.Query(context.Background(), query)
	if err != nil {
		http.Error(w, "Failed to fetch films", http.StatusInternalServerError)
		log.Printf("Error fetching films: %v\n", err)
		return
	}
	defer rows.Close()

	// Инициализируем слайс для хранения данных о фильмах
	var films []database.FilmWithActors

	// Сканируем результаты запроса в слайс
	for rows.Next() {
		var film database.FilmWithActors
		var actorsJSON []byte

		if scanErr := rows.Scan(&film.ID, &film.Name, &film.Description, &film.ReleaseDate, &film.Rate, &actorsJSON); scanErr != nil {
			log.Println("Error scanning film information:", scanErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if actorsJSON == nil {
			actorsJSON = []byte("[]")
		}
		if unmarshalErr := json.Unmarshal(actorsJSON, &film.Actors); unmarshalErr != nil {
			log.Println("Error decoding actors JSON:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		films = append(films, film)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error iterating over film rows:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Кодируем данные о фильмах в формат JSON и отправляем клиенту
	w.Header().Set("Content-Type", "application/json")
	if encodeErr := json.NewEncoder(w).Encode(films); encodeErr != nil {
		log.Println("Error encoding films to JSON:", encodeErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// @Summary Удаление фильма по идентификатору
// @Description Удаляет фильм из базы данных по его идентификатору
// @Tags film
// @Accept json
// @Produce json
// @Param id query int true "Идентификатор фильма для удаления"
// @Success 204 {string} string "Успешное удаление фильма"
// @Failure 400 {string} string "Отсутствует обязательный параметр 'id'"
// @Failure 404 {string} string "Фильм не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /api/film-delete [delete]
func DeleteFilm(w http.ResponseWriter, r *http.Request, dbpool *pgxpool.Pool) {
	// Извлекаем ID фильма из параметров запроса
	filmIDStr := r.URL.Query().Get("id")
	if filmIDStr == "" {
		http.Error(w, "Missing required parameter 'id'", http.StatusBadRequest)
		return
	}

	// Преобразуем строковый ID фильма в целочисленный формат
	filmID, idErr := strconv.Atoi(filmIDStr)
	if idErr != nil {
		http.Error(w, "Invalid film ID", http.StatusBadRequest)
		log.Printf("Error converting film ID to integer: %v\n", idErr)
		return
	}

	// Выполняем запрос на удаление фильма из базы данных
	result, err := dbpool.Exec(context.Background(), "DELETE FROM film WHERE id = $1", filmID)
	if err != nil {
		http.Error(w, "Failed to delete film", http.StatusInternalServerError)
		log.Printf("Error deleting film: %v\n", err)
		return
	}

	// Проверяем количество удаленных записей
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Film not found", http.StatusNotFound)
		log.Println("Film not found for deletion")
		return
	}

	// Возвращаем успешный статус удаления
	w.WriteHeader(http.StatusNoContent)
}

// @Summary Обновление фильма
// @Description Обновляет информацию о фильме и его актерах в базе данных
// @Tags film
// @Accept json
// @Produce json
// @Param id query int true "Идентификатор фильма для обновления"
// @Param body body database.UpdateFilmWithActors true "Данные для обновления фильма"
// @Success 201 {string} string "Фильм успешно обновлен"
// @Failure 400 {string} string "Отсутствует обязательный параметр 'id' или некорректные данные JSON"
// @Failure 404 {string} string "Фильм не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /api/film-update [patch]
func UpdateFilm(w http.ResponseWriter, r *http.Request, dbpool *pgxpool.Pool) {
	// Извлекаем ID фильма из параметров запроса
	filmIDStr := r.URL.Query().Get("id")
	if filmIDStr == "" {
		http.Error(w, "Missing required parameter 'id'", http.StatusBadRequest)
		return
	}

	// Преобразуем строковый ID фильма в целочисленный формат
	filmID, idErr := strconv.Atoi(filmIDStr)
	if idErr != nil {
		http.Error(w, "Invalid film ID", http.StatusBadRequest)
		log.Printf("Error converting film ID to integer: %v\n", idErr)
		return
	}

	// Декодируем JSON-данные из тела запроса в структуру UpdateFilmWithActors
	var updatedFilm database.UpdateFilmWithActors
	if decoderErr := json.NewDecoder(r.Body).Decode(&updatedFilm); decoderErr != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		log.Printf("Error decode JSON: %v\n", decoderErr)
		return
	}
	defer r.Body.Close()

	// Подготавливаем запрос обновления фильма
	var queryBuilder strings.Builder
	var args []interface{}
	queryBuilder.WriteString("UPDATE film SET ")

	// Счетчик для отслеживания номера текущего параметра
	paramCounter := 1

	// Проверяем, есть ли какие-либо поля для обновления
	if updatedFilm.Name != "" {
		queryBuilder.WriteString("name = $")
		queryBuilder.WriteString(strconv.Itoa(paramCounter))
		queryBuilder.WriteString(", ")
		args = append(args, updatedFilm.Name)
		paramCounter++
	}
	if updatedFilm.Description != "" {
		queryBuilder.WriteString("description = $")
		queryBuilder.WriteString(strconv.Itoa(paramCounter))
		queryBuilder.WriteString(", ")
		args = append(args, updatedFilm.Description)
		paramCounter++
	}
	if !updatedFilm.ReleaseDate.IsZero() {
		queryBuilder.WriteString("release_date = $")
		queryBuilder.WriteString(strconv.Itoa(paramCounter))
		queryBuilder.WriteString(", ")
		args = append(args, updatedFilm.ReleaseDate)
		paramCounter++
	}
	if updatedFilm.Rate != 0 {
		queryBuilder.WriteString("rate = $")
		queryBuilder.WriteString(strconv.Itoa(paramCounter))
		queryBuilder.WriteString(", ")
		args = append(args, updatedFilm.Rate)
		paramCounter++
	}

	// Удаляем последнюю запятую и пробел из запроса перед оператором SET
	query := strings.TrimSuffix(queryBuilder.String(), ", ")

	// Добавляем условие WHERE для указания фильма по ID
	query += " WHERE id = $"
	query += strconv.Itoa(paramCounter)
	args = append(args, filmID)

	// Выполняем запрос обновления в базе данных
	_, err := dbpool.Exec(context.Background(), query, args...)
	if err != nil {
		log.Println("Error updating film:", err)
		http.Error(w, "Error updating film", http.StatusInternalServerError)
		return
	}

	// Обновляем список актеров фильма
	if len(updatedFilm.Actors) > 0 {
		// Сначала удаляем существующих актеров у фильма
		_, deleteErr := dbpool.Exec(context.Background(), "DELETE FROM film_actor WHERE film_id = $1", filmID)
		if deleteErr != nil {
			log.Println("Error deleting film actors:", deleteErr)
			http.Error(w, "Error updating film", http.StatusInternalServerError)
			return
		}

		// Затем добавляем новых актеров к фильму
		for _, actorID := range updatedFilm.Actors {
			_, insertErr := dbpool.Exec(context.Background(), "INSERT INTO film_actor (film_id, actor_id) VALUES ($1, $2)", filmID, actorID)
			if insertErr != nil {
				log.Println("Error inserting film actors:", insertErr)
				http.Error(w, "Error updating film", http.StatusInternalServerError)
				return
			}
		}
	}

	w.WriteHeader(http.StatusCreated)
}
