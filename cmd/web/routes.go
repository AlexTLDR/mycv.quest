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

	// Serve template static files (preview images, etc.)
	templateFileServer := http.StripPrefix("/static/templates/", http.FileServer(http.Dir("assets/templates/typst/")))
	mux.Handle("GET /static/templates/", templateFileServer)

	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /cv-builder", app.cvBuilder)
	mux.HandleFunc("GET /templates", app.templates)
	mux.HandleFunc("GET /templates/{id}/preview", app.templatePreview)

	mux.Handle("GET /basic-auth-protected", app.requireBasicAuthentication(http.HandlerFunc(app.protected)))

	return app.logAccess(app.recoverPanic(app.securityHeaders(app.sessionManager.LoadAndSave(mux))))
}
