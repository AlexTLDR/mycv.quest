package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/AlexTLDR/mycv.quest/pkg/generator"
	"github.com/AlexTLDR/mycv.quest/templates"
)

type Server struct {
	generator      *generator.CVGenerator
	sessionManager *SessionManager
}

func New(gen *generator.CVGenerator) *Server {
	return &Server{
		generator:      gen,
		sessionManager: NewSessionManager(),
	}
}

func (s *Server) SetupRoutes() {
	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("."))))

	// Home page
	http.HandleFunc("/", s.HandleIndex)

	// Form endpoints
	http.HandleFunc("/form/", s.HandleForm)

	// Generate CV endpoint
	http.HandleFunc("/generate/", s.HandleGenerate)

	// Serve session-specific generated PDFs
	http.HandleFunc("/cv/", s.HandleSessionPDF)
}

func (s *Server) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	templateData := s.generator.GetTemplateData()
	if err := templates.Index(templateData).Render(r.Context(), w); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

func (s *Server) HandleForm(w http.ResponseWriter, r *http.Request) {
	templateKey := strings.TrimPrefix(r.URL.Path, "/form/")

	switch templateKey {
	case "basic":
		if err := templates.BasicForm().Render(r.Context(), w); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	case "modern":
		if err := templates.ModernForm().Render(r.Context(), w); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	case "vantage":
		if err := templates.VantageForm().Render(r.Context(), w); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) HandleGenerate(w http.ResponseWriter, r *http.Request) {
	templateKey := strings.TrimPrefix(r.URL.Path, "/generate/")

	if r.Method == http.MethodGet {
		// Redirect to form page
		http.Redirect(w, r, fmt.Sprintf("/form/%s", templateKey), http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		// Get or create session
		session := s.sessionManager.GetOrCreateSession(r)
		s.sessionManager.SetSessionCookie(w, session)

		// Generate CV in memory
		pdfData, err := s.generator.GenerateFromForm(templateKey, r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error generating CV: %v", err), http.StatusInternalServerError)
			return
		}

		// Store PDF in session
		s.sessionManager.StorePDF(session.ID, templateKey, pdfData)

		// Redirect to the session-specific generated PDF
		http.Redirect(w, r, fmt.Sprintf("/cv/%s/%s.pdf", session.ID, templateKey), http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) HandleSessionPDF(w http.ResponseWriter, r *http.Request) {
	// Parse URL path: /cv/{sessionID}/{template}.pdf
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/cv/"), "/")
	if len(pathParts) != 2 {
		http.NotFound(w, r)
		return
	}

	sessionID := pathParts[0]
	templateFile := pathParts[1]

	// Extract template key from filename
	templateKey := strings.TrimSuffix(templateFile, ".pdf")

	// Get PDF data from session
	pdfData, exists := s.sessionManager.GetPDF(sessionID, templateKey)
	if !exists {
		http.NotFound(w, r)
		return
	}

	// Set headers for PDF response
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"cv-%s.pdf\"", templateKey))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfData)))

	// Write PDF data
	if _, err := w.Write(pdfData); err != nil {
		http.Error(w, "Failed to write PDF data", http.StatusInternalServerError)
	}
}
