package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/AlexTLDR/mycv.quest/pkg/generator"
	"github.com/AlexTLDR/mycv.quest/templates"
)

type Server struct {
	generator *generator.CVGenerator
}

func New(gen *generator.CVGenerator) *Server {
	return &Server{
		generator: gen,
	}
}

func (s *Server) SetupRoutes() {
	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("."))))

	// Home page
	http.HandleFunc("/", s.handleIndex)

	// Form endpoints
	http.HandleFunc("/form/", s.handleForm)

	// Generate CV endpoint
	http.HandleFunc("/generate/", s.handleGenerate)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	templateData := s.generator.GetTemplateData()
	templates.Index(templateData).Render(r.Context(), w)
}

func (s *Server) handleForm(w http.ResponseWriter, r *http.Request) {
	templateKey := strings.TrimPrefix(r.URL.Path, "/form/")

	switch templateKey {
	case "basic":
		templates.BasicForm().Render(r.Context(), w)
	case "modern":
		templates.ModernForm().Render(r.Context(), w)
	case "vantage":
		templates.VantageForm().Render(r.Context(), w)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	templateKey := strings.TrimPrefix(r.URL.Path, "/generate/")

	if r.Method == "GET" {
		// Redirect to form page
		http.Redirect(w, r, fmt.Sprintf("/form/%s", templateKey), http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		if err := s.generator.GenerateFromForm(templateKey, r); err != nil {
			http.Error(w, fmt.Sprintf("Error generating CV: %v", err), http.StatusInternalServerError)
			return
		}

		// Redirect to the generated PDF
		http.Redirect(w, r, fmt.Sprintf("/static/output/cv-%s.pdf", templateKey), http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
