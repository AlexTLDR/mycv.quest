package main

import (
	"net/http"

	"github.com/AlexTLDR/mycv.quest/assets"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.FS(assets.EmbeddedFiles))
	mux.Handle("GET /static/", fileServer)

	mux.HandleFunc("GET /{$}", app.home)

	mux.Handle("GET /basic-auth-protected", app.requireBasicAuthentication(http.HandlerFunc(app.protected)))

	// CV API routes
	mux.HandleFunc("GET /api/cv/templates", app.listTemplates)
	mux.HandleFunc("GET /api/cv/templates/{id}", app.getTemplate)
	mux.HandleFunc("GET /api/cv/templates/{id}/form", app.getTemplateForm)
	mux.HandleFunc("GET /api/cv/templates/{id}/sample", app.getTemplateSample)
	mux.HandleFunc("GET /api/cv/templates/{id}/metadata", app.getTemplateMetadata)
	mux.HandleFunc("GET /api/cv/templates/{id}/debug", app.debugTemplate)
	mux.HandleFunc("POST /api/cv/templates/{id}/validate", app.validateTemplateData)
	mux.HandleFunc("POST /api/cv/templates/{id}/generate", app.generateCV)
	mux.HandleFunc("POST /api/cv/templates/{id}/preview", app.generateCVPreview)
	mux.HandleFunc("GET /api/cv/templates/{id}/quick", app.quickGenerate)
	mux.HandleFunc("POST /api/cv/convert", app.convertData)
	mux.HandleFunc("GET /api/cv/system/typst", app.validateTypstInstallation)

	return app.logAccess(app.recoverPanic(app.securityHeaders(app.sessionManager.LoadAndSave(mux))))
}
