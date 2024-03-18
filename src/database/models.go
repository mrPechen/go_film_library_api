package database

import (
	"time"
)

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Actor struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Gender    string    `json:"gender"`
	BirthDate time.Time `json:"birth_date"`
}

type Film struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ReleaseDate time.Time `json:"release_date"`
	Rate        int       `json:"rate"`
	ActorIDs    []int     `json:"actor_ids"`
	Actors      []Actor   `json:"actors"`
}

type FilmForActor struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ReleaseDate time.Time `json:"release_date"`
	Rate        int       `json:"rate"`
}

type ActorWithFilms struct {
	Actor Actor          `json:"actor"` // Информация об актере
	Films []FilmForActor `json:"films"` // Список фильмов, в которых участвует актер
}

type FilmWithActors struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ReleaseDate time.Time `json:"release_date"`
	Rate        int       `json:"rate"`
	Actors      []Actor   `json:"actors"`
}

type UpdateFilmWithActors struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ReleaseDate time.Time `json:"release_date"`
	Rate        int       `json:"rate"`
	Actors      []int     `json:"actors"`
}
