package main

import (
	"fmt"
	"net/http"

	templates "github.com/AlexTLDR/mycv.quest/assets/templates/templ"
	"github.com/AlexTLDR/mycv.quest/internal/cv"
	"github.com/AlexTLDR/mycv.quest/internal/response"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	err := response.Component(w, http.StatusOK, templates.HomePage(data))
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) protected(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is a protected handler"))
}

func (app *application) cvBuilder(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	// Get available templates dynamically
	availableTemplates, err := app.getAvailableTemplates()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Convert to template struct for the template
	var templateData []templates.Template
	for _, t := range availableTemplates {
		templateData = append(templateData, templates.Template{
			ID:           t.ID,
			Name:         t.Name,
			Description:  t.Description,
			Version:      t.Version,
			Author:       t.Author,
			PreviewImage: t.PreviewImage,
			Repository:   t.Repository,
		})
	}

	err = response.Component(w, http.StatusOK, templates.CVBuilderPage(data, templateData))
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) templates(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	// Get available templates dynamically
	availableTemplates, err := app.getAvailableTemplates()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Convert to TemplateInfo struct for the template
	var templateData []templates.TemplateInfo
	for _, t := range availableTemplates {
		templateData = append(templateData, templates.TemplateInfo{
			ID:           t.ID,
			Name:         t.Name,
			Description:  t.Description,
			Version:      t.Version,
			Author:       t.Author,
			PreviewImage: t.PreviewImage,
			Repository:   t.Repository,
		})
	}

	err = response.Component(w, http.StatusOK, templates.TemplatesPage(data, templateData))
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) templatePreview(w http.ResponseWriter, r *http.Request) {
	// Get template ID from URL path
	templateID := r.PathValue("id")
	if templateID == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Create dummy data for preview
	dummyData := map[string]interface{}{
		"name":     "John Doe",
		"email":    "john.doe@example.com",
		"phone":    "+1 (555) 123-4567",
		"location": "New York, NY",
		"website":  "https://johndoe.dev",
		"linkedin": "https://linkedin.com/in/johndoe",
		"github":   "https://github.com/johndoe",
		"summary":  "Experienced software engineer with 5+ years of expertise in full-stack development, specializing in modern web technologies and cloud infrastructure.",
		"experience": []map[string]interface{}{
			{
				"title":       "Senior Software Engineer",
				"company":     "Tech Solutions Inc.",
				"location":    "New York, NY",
				"date_start":  "2022-01",
				"date_end":    "Present",
				"description": "Led development of microservices architecture serving 1M+ users. Implemented CI/CD pipelines reducing deployment time by 60%.",
			},
			{
				"title":       "Software Engineer",
				"company":     "StartupCorp",
				"location":    "San Francisco, CA",
				"date_start":  "2020-06",
				"date_end":    "2021-12",
				"description": "Developed responsive web applications using React and Node.js. Collaborated with cross-functional teams to deliver high-quality features.",
			},
		},
		"education": []map[string]interface{}{
			{
				"degree":      "Bachelor of Science in Computer Science",
				"institution": "University of Technology",
				"location":    "Boston, MA",
				"date_start":  "2016-09",
				"date_end":    "2020-05",
				"gpa":         "3.8/4.0",
			},
		},
		"skills": []string{
			"JavaScript", "TypeScript", "React", "Node.js", "Go", "Python",
			"AWS", "Docker", "Kubernetes", "PostgreSQL", "MongoDB", "Git",
		},
		"projects": []map[string]interface{}{
			{
				"name":         "E-Commerce Platform",
				"description":  "Built a scalable e-commerce platform using microservices architecture",
				"technologies": "React, Node.js, PostgreSQL, AWS",
				"url":          "https://github.com/johndoe/ecommerce",
			},
			{
				"name":         "Task Management App",
				"description":  "Developed a collaborative task management application with real-time updates",
				"technologies": "Vue.js, Express, Socket.io, MongoDB",
				"url":          "https://github.com/johndoe/taskmanager",
			},
		},
	}

	// Generate preview PDF
	if app.cvService != nil {
		templateData := cv.TemplateData{
			TemplateID: templateID,
			Data:       dummyData,
		}
		result, err := app.cvService.GeneratePreview(templateID, templateData)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		// Set appropriate headers for PDF
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s-preview.pdf\"", templateID))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(result.Data)))

		// Write the PDF content
		w.WriteHeader(http.StatusOK)
		w.Write(result.Data)
		return
	}

	// Fallback if CV service is not available
	app.serverError(w, r, fmt.Errorf("CV service not available"))
}
