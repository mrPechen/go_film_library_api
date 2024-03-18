package test

import (
	"src/src/test/actor"
	"src/src/test/film"
	"src/src/test/test_auth"
	"testing"
)

func Test(t *testing.T) {
	test_auth.RegisterUser(t)
	test_auth.RegisterAdmin(t)
	test_auth.Login("user@test.ru", t)
	test_auth.Login("admin@test.ru", t)
	actor.CreateActor(t)
	actor.GetActorsWithoutFilms(t)
	actor.GetActorWithoutFilms(t)
	actor.UpdateActor(t)
	actor.GetActorWithoutFilms(t)
	film.CreateFilm(t)
	film.GetFilms(t)
	film.GetFilm(t)
	film.UpdateFilm(t)
	film.GetFilm(t)
	film.DeleteFilm(t)
	actor.DeleteActor(t)
}
