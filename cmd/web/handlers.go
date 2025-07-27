package main

import (
	"fmt"
	"net/http"

	templates "github.com/AlexTLDR/mycv.quest/assets/templates/templ"
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

	// Serve the preview image directly
	imagePath := fmt.Sprintf("assets/templates/typst/%s/preview.png", templateID)

	// Set headers for image
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	// Serve the image file
	http.ServeFile(w, r, imagePath)
}
