package routers

import (
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
	"src/src/services/actor"
	"src/src/services/film"
	"src/src/services/login"
	"src/src/services/middleware"
	"src/src/services/refresh"
	"src/src/services/registration"
)

func Handlers(dbpool *pgxpool.Pool, router *http.ServeMux) {
	router.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir("docs"))))
	router.Handle("/", httpSwagger.Handler(
		httpSwagger.URL("/docs/swagger.json"), // URL для Swagger JSON-документа
	))
	createActor := middleware.AuthMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		actor.CreateActor(writer, request, dbpool)
	}))

	updateActor := middleware.AuthMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPatch {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		actor.UpdateActor(writer, request, dbpool)
	}))

	deleteActor := middleware.AuthMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodDelete {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		actor.DeleteActor(writer, request, dbpool)
	}))

	getActor := middleware.AuthMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		actor.GetActor(writer, request, dbpool)
	}))

	getActors := middleware.AuthMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		actor.GetAllActors(writer, request, dbpool)
	}))

	createFilm := middleware.AuthMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		film.CreateFilm(writer, request, dbpool)
	}))

	getFilm := middleware.AuthMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		film.SearchFilm(writer, request, dbpool)
	}))

	getFilms := middleware.AuthMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		film.GetAllFilms(writer, request, dbpool)
	}))

	deleteFilm := middleware.AuthMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodDelete {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		film.DeleteFilm(writer, request, dbpool)
	}))

	updateFilm := middleware.AuthMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPatch {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		film.UpdateFilm(writer, request, dbpool)
	}))

	router.HandleFunc("/api/login/", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		login.Login(writer, request, dbpool)
	})
	router.HandleFunc("/api/registration/", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		registration.RegisterUser(writer, request, dbpool)
	})
	router.HandleFunc("/api/refresh-tokens/", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		refresh.RefreshTokens(writer, request, dbpool)
	})
	router.Handle("/api/actor-create/", createActor)
	router.Handle("/api/actor-update", updateActor)
	router.Handle("/api/actor-delete", deleteActor)
	router.Handle("/api/actor", getActor)
	router.Handle("/api/actors/", getActors)
	router.Handle("/api/film-create/", createFilm)
	router.Handle("/api/film", getFilm)
	router.Handle("/api/films", getFilms)
	router.Handle("/api/film-delete", deleteFilm)
	router.Handle("/api/film-update", updateFilm)
}
