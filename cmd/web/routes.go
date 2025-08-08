package main

import (
	"net/http"
	"strings"

	"github.com/AlexTLDR/mycv.quest/assets"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// Create a wrapper that fixes MIME types for JavaScript files
	fileServer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set correct MIME type for JavaScript files
		if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		}

		// Create a custom ResponseWriter that prevents the file server from overriding our Content-Type
		wrappedWriter := &mimePreservingWriter{
			ResponseWriter: w,
			isJS:           strings.HasSuffix(r.URL.Path, ".js"),
		}

		http.FileServer(http.FS(assets.EmbeddedFiles)).ServeHTTP(wrappedWriter, r)
	})

	mux.Handle("GET /static/", fileServer)
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", fileServer))

	// Serve template static files (preview images, etc.)
	templateFileServer := http.StripPrefix("/static/templates/", http.FileServer(http.Dir("assets/templates/typst/")))
	mux.Handle("GET /static/templates/", templateFileServer)

	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /cv-builder", app.cvBuilder)
	mux.HandleFunc("GET /templates", app.templates)
	mux.HandleFunc("GET /templates/{id}/preview", app.templatePreview)
	mux.HandleFunc("GET /api/templates/{id}/form", app.getTemplateForm)

	// API routes for CV generation
	mux.HandleFunc("POST /api/cv/generate", app.generateCV)
	mux.HandleFunc("POST /api/cv/preview", app.generatePreview)

	mux.Handle("GET /basic-auth-protected", app.requireBasicAuthentication(http.HandlerFunc(app.protected)))

	return app.logAccess(app.recoverPanic(app.securityHeaders(app.sessionManager.LoadAndSave(mux))))
}

// mimePreservingWriter prevents the file server from overriding our MIME type for JS files
type mimePreservingWriter struct {
	http.ResponseWriter
	isJS          bool
	headerWritten bool
}

func (w *mimePreservingWriter) WriteHeader(statusCode int) {
	if !w.headerWritten && w.isJS {
		// Ensure our JavaScript MIME type is preserved
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	}
	w.headerWritten = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *mimePreservingWriter) Write(data []byte) (int, error) {
	if !w.headerWritten {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}
