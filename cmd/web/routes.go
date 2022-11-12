package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})
	// Leave the static files route unchanged.
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	dynamic := alice.New(app.sessionManager.LoadAndSave)

	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.main))
	router.Handler(http.MethodGet, "/home", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/note/view/:id", dynamic.ThenFunc(app.victim))
	router.Handler(http.MethodGet, "/note/create", dynamic.ThenFunc(app.create))
	router.Handler(http.MethodPost, "/note/create", dynamic.ThenFunc(app.createPost))

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	return standard.Then(router)
}
