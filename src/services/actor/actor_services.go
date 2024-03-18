package actor

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"src/src/database"
	"strconv"
	"strings"
)

// @Summary Создать актера
// @Description Создать актера
// @Tags actor
// @Accept json
// @Param actor body database.Actor true "Данные для создания актера"
// @Success 201
// @Router /api/actor-create/ [post]
func CreateActor(w http.ResponseWriter, r *http.Request, dbpool *pgxpool.Pool) {
	var newActor database.Actor
	if actorErr := json.NewDecoder(r.Body).Decode(&newActor); actorErr != nil {
		http.Error(w, actorErr.Error(), http.StatusBadRequest)
		log.Printf("Error decoding JSON: %v\n", actorErr)
		return
	}

	defer r.Body.Close()

	var actorID int
	actorIDErr := dbpool.QueryRow(context.Background(), `
        INSERT INTO actor (name, gender, birth_date) VALUES ($1, $2, $3) RETURNING id
    `, newActor.Name, newActor.Gender, newActor.BirthDate).Scan(&actorID)
	if actorIDErr != nil {
		http.Error(w, "Failed to create actor", http.StatusInternalServerError)
		log.Printf("Error inserting actor into the database: %v\n", actorIDErr)
		return
	}

	log.Printf("New actor created with ID: %d\n", actorID)
	w.WriteHeader(http.StatusCreated)
}

// @Summary Обновить информацию об актере
// @Description Обновляет информацию об актере на основе предоставленных данных
// @Tags actor
// @Accept json
// @Param id path int true "ID актера для обновления"
// @Param actor body database.Actor true "Данные для обновления актера"
// @Success 201
// @Router /api/actor-update [patch]
func UpdateActor(w http.ResponseWriter, r *http.Request, dbpool *pgxpool.Pool) {

	// Получаем значение параметра из URL запроса
	actorIDStr := r.URL.Query().Get("id")
	if actorIDStr == "" {
		http.Error(w, "Missing required parameter 'id'", http.StatusBadRequest)
		return
	}
	actorID, actorIDErr := strconv.Atoi(actorIDStr)
	if actorIDErr != nil {
		log.Println(actorIDErr)
		http.Error(w, "Invalid actor ID", http.StatusBadRequest)
		return
	}

	// Декодируем JSON-данные из тела запроса
	var updatedActor database.Actor
	if updatedActorErr := json.NewDecoder(r.Body).Decode(&updatedActor); updatedActorErr != nil {
		log.Println(updatedActorErr)
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	// Подготавливаем запрос обновления актера
	var queryBuilder strings.Builder
	var args []interface{}
	queryBuilder.WriteString("UPDATE actor SET ")

	// Счетчик для отслеживания номера текущего параметра
	paramCounter := 1

	// Проверяем, есть ли какие-либо поля для обновления
	if updatedActor.Name != "" {
		queryBuilder.WriteString("name = $")
		queryBuilder.WriteString(strconv.Itoa(paramCounter))
		queryBuilder.WriteString(", ")
		args = append(args, updatedActor.Name)
		paramCounter++
	}
	if updatedActor.Gender != "" {
		queryBuilder.WriteString("gender = $")
		queryBuilder.WriteString(strconv.Itoa(paramCounter))
		queryBuilder.WriteString(", ")
		args = append(args, updatedActor.Gender)
		paramCounter++
	}
	if !updatedActor.BirthDate.IsZero() {
		queryBuilder.WriteString("birth_date = $")
		queryBuilder.WriteString(strconv.Itoa(paramCounter))
		queryBuilder.WriteString(", ")
		args = append(args, updatedActor.BirthDate)
		paramCounter++
	}

	// Удаляем последнюю запятую и пробел из запроса перед оператором SET
	query := strings.TrimSuffix(queryBuilder.String(), ", ")

	// Добавляем условие WHERE для указания актера по ID
	query += " WHERE id = $"
	query += strconv.Itoa(paramCounter)
	args = append(args, actorID)

	// Выполняем запрос обновления в базе данных
	_, err := dbpool.Exec(context.Background(), query, args...)
	if err != nil {
		log.Println("Error updating actor:", err)
		http.Error(w, "Error updating actor", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// getActor обработчик для получения информации об актере
// @Summary Получить информацию об актере
// @Description Получает информацию об актере
// @Tags actor
// @Param id path int true "ID актера для обновления"
// @Produce json
// @Success 200 {object} database.Actor
// @Router /api/actor [get]
func GetActor(w http.ResponseWriter, r *http.Request, dbpool *pgxpool.Pool) {

	// Получаем значение параметра из URL запроса
	actorIDStr := r.URL.Query().Get("id")
	if actorIDStr == "" {
		http.Error(w, "Missing required parameter 'id'", http.StatusBadRequest)
		return
	}

	// Преобразуем ID актера в число
	actorID, actorIDErr := strconv.Atoi(actorIDStr)
	if actorIDErr != nil {
		http.Error(w, "Invalid actor ID", http.StatusBadRequest)
		log.Printf("Error converting actor ID to int: %v\n", actorIDErr)
		return
	}

	// Подготавливаем запрос для получения информации об актере и его фильмах
	query := `
		SELECT 
			a.id, a.name, a.gender, a.birth_date,
			COALESCE(json_agg(json_build_object(
				'id', f.id, 
				'name', f.name, 
				'description', f.description, 
				'release_date', f.release_date, 
				'rate', f.rate
			)) FILTER (WHERE f.id IS NOT NULL), '[]') AS films
		FROM 
			actor a
		LEFT JOIN 
			film_actor fa ON a.id = fa.actor_id
		LEFT JOIN 
			film f ON f.id = fa.film_id
		WHERE 
			a.id = $1
		GROUP BY 
			a.id, a.name, a.gender, a.birth_date
	`

	// Выполняем запрос к базе данных
	row := dbpool.QueryRow(context.Background(), query, actorID)

	// Инициализируем переменные для хранения данных об актере и его фильмах
	var actor database.ActorWithFilms

	// Сканируем результаты запроса
	if scanErr := row.Scan(&actor.Actor.ID, &actor.Actor.Name, &actor.Actor.Gender, &actor.Actor.BirthDate, &actor.Films); scanErr != nil {
		// Если произошла ошибка при сканировании, отправляем клиенту сообщение об ошибке
		http.Error(w, "Actor not found", http.StatusNotFound)
		log.Printf("Error scanning actor data: %v\n", scanErr)
		return
	}

	// Кодируем данные об актере и его фильмах в формат JSON и отправляем клиенту
	w.Header().Set("Content-Type", "application/json")
	if encodeErr := json.NewEncoder(w).Encode(actor); encodeErr != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Error encoding actor data: %v\n", encodeErr)
		return
	}
}

// getAllActors обработчик для получения информации об актерах
// @Summary Получить информацию об актерах
// @Description Получает информацию об актерах
// @Tags actor
// @Produce json
// @Success 200 {array} database.Actor
// @Router /api/actors/ [get]
func GetAllActors(w http.ResponseWriter, r *http.Request, dbpool *pgxpool.Pool) {

	// Подготавливаем запрос для получения всех актеров
	actorsQuery := `
		SELECT 
			a.id, a.name, a.gender, a.birth_date,
			COALESCE(json_agg(json_build_object(
				'id', f.id, 
				'name', f.name, 
				'description', f.description, 
				'release_date', f.release_date, 
				'rate', f.rate
			)) FILTER (WHERE f.id IS NOT NULL), '[]') AS films
		FROM 
			actor a
		LEFT JOIN 
			film_actor fa ON a.id = fa.actor_id
		LEFT JOIN 
			film f ON f.id = fa.film_id
		GROUP BY 
			a.id
	`

	// Выполняем запрос к базе данных
	rows, rowsErr := dbpool.Query(context.Background(), actorsQuery)
	if rowsErr != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Error in query: %v\n", rowsErr)
		return
	}
	defer rows.Close()

	// Инициализируем слайс для хранения данных об актерах
	var actors []database.ActorWithFilms

	// Сканируем результаты запроса для каждого актера
	for rows.Next() {
		var actor database.ActorWithFilms
		var filmsJSON []byte
		if iterErr := rows.Scan(&actor.Actor.ID, &actor.Actor.Name, &actor.Actor.Gender, &actor.Actor.BirthDate, &filmsJSON); iterErr != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Printf("Error iter rows: %v\n", iterErr)
			return
		}

		// Распаковываем JSON с фильмами во временную переменную
		var films []database.FilmForActor
		if unmarshalErr := json.Unmarshal(filmsJSON, &films); unmarshalErr != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Printf("Error unmarshaling films JSON: %v\n", unmarshalErr)
			return
		}

		// Устанавливаем фильмы в поле Films структуры ActorWithFilms
		actor.Films = films

		actors = append(actors, actor)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Error: %v\n", err)
		return
	}

	// Кодируем данные об актерах и их фильмах в формат JSON и отправляем клиенту
	w.Header().Set("Content-Type", "application/json")
	if encodeErr := json.NewEncoder(w).Encode(actors); encodeErr != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Error encoding: %v\n", encodeErr)
		return
	}
}

// @Summary Удаление актера по идентификатору
// @Description Удаляет актера из базы данных по его идентификатору
// @Tags actor
// @Accept json
// @Produce json
// @Param id query int true "Идентификатор актера для удаления"
// @Success 204 {string} string "Успешное удаление актера"
// @Failure 400 {string} string "Отсутствует обязательный параметр 'id'"
// @Failure 404 {string} string "Актер не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /actor-delete [delete]
func DeleteActor(w http.ResponseWriter, r *http.Request, dbpool *pgxpool.Pool) {
	// Получаем ID актера из параметров запроса
	actorIDStr := r.URL.Query().Get("id")
	if actorIDStr == "" {
		http.Error(w, "Missing required parameter 'id'", http.StatusBadRequest)
		return
	}

	// Преобразуем строковый ID актера в целочисленный формат
	actorID, idErr := strconv.Atoi(actorIDStr)
	if idErr != nil {
		http.Error(w, "Invalid actor ID", http.StatusBadRequest)
		log.Printf("Error converting actor ID to integer: %v\n", idErr)
		return
	}

	// Выполняем запрос на удаление актера из базы данных
	_, err := dbpool.Exec(context.Background(), "DELETE FROM actor WHERE id = $1", actorID)
	if err != nil {
		log.Println("Error deleting actor:", err)
		http.Error(w, "Error deleting actor", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
