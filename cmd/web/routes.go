package main

import (
	"net/http"

	"github.com/AlexTLDR/mycv.quest/assets"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.FS(assets.EmbeddedFiles))
	mux.Handle("GET /static/", fileServer)
	mux.Handle("GET /assets/", fileServer)

	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /cv-builder", app.cvBuilder)
	mux.HandleFunc("GET /templates", app.templates)

	mux.Handle("GET /basic-auth-protected", app.requireBasicAuthentication(http.HandlerFunc(app.protected)))

	return app.logAccess(app.recoverPanic(app.securityHeaders(app.sessionManager.LoadAndSave(mux))))
}
