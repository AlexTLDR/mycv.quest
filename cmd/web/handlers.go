package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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

	// Get available templates
	availableTemplates, err := app.cvService.GetAvailableTemplates()
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
		})
	}

	err = response.Component(w, http.StatusOK, templates.CVBuilderPage(data, templateData))
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) templates(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	// Get available templates
	availableTemplates, err := app.cvService.GetAvailableTemplates()
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

func (app *application) getTemplateForm(w http.ResponseWriter, r *http.Request) {
	templateID := r.PathValue("id")
	if templateID == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Return a simple form structure - for now just return success
	// In the future this could return template-specific form fields
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"template_id": templateID,
		"success":     true,
	}
	json.NewEncoder(w).Encode(response)
}

func (app *application) generateCV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form data
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Get template ID
	templateID := r.FormValue("template_id")
	if templateID == "" {
		http.Error(w, "Template ID is required", http.StatusBadRequest)
		return
	}

	// Debug: Log raw form data
	app.logger.Info("Raw form data received:")
	for key, values := range r.Form {
		app.logger.Info("Form field", "key", key, "values", values)
	}

	// Convert form data to simple map
	formData := app.convertFormToMap(r)

	// Debug: Log converted form data
	app.logger.Info("Converted form data", "data", formData)

	// Generate CV
	request := cv.GenerationRequest{
		TemplateID: templateID,
		Data:       formData,
		Format:     "pdf",
	}

	result, err := app.cvService.GenerateCV(request)
	if err != nil {
		app.logger.Error("Failed to generate CV", "error", err)
		http.Error(w, "Failed to generate CV: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if !result.Success {
		app.logger.Error("CV generation failed", "message", result.Message)
		http.Error(w, "CV generation failed: "+result.Message, http.StatusInternalServerError)
		return
	}

	// Set headers for PDF download
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", result.Filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(result.Data)))

	// Write PDF data
	w.Write(result.Data)
}

func (app *application) generatePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form data
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Get template ID
	templateID := r.FormValue("template_id")
	if templateID == "" {
		http.Error(w, "Template ID is required", http.StatusBadRequest)
		return
	}

	// Convert form data to simple map
	formData := app.convertFormToMap(r)

	// Generate CV (same as regular generation for now)
	request := cv.GenerationRequest{
		TemplateID: templateID,
		Data:       formData,
		Format:     "pdf",
	}

	result, err := app.cvService.GenerateCV(request)
	if err != nil {
		app.logger.Error("Failed to generate preview", "error", err)
		http.Error(w, "Failed to generate preview: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if !result.Success {
		app.logger.Error("Preview generation failed", "message", result.Message)
		http.Error(w, "Preview generation failed: "+result.Message, http.StatusInternalServerError)
		return
	}

	// Return JSON response with success status
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success":  true,
		"message":  "Preview generated successfully",
		"filename": result.Filename,
	}
	json.NewEncoder(w).Encode(response)
}

func (app *application) convertFormToMap(r *http.Request) map[string]interface{} {
	data := make(map[string]interface{})

	// Handle photo upload
	usePhoto := r.FormValue("use_photo") == "on"
	data["use_photo"] = usePhoto

	// Simple form field mapping
	for key, values := range r.Form {
		if key == "template_id" || key == "profile_photo" {
			continue // Skip these special fields
		}

		if len(values) == 0 {
			continue
		}

		value := values[0]
		if value == "" {
			continue
		}

		// Handle comma-separated fields that should become arrays
		if app.isCommaSeparatedField(key) {
			arrayValue := app.parseCommaSeparatedValue(value)
			data[key] = arrayValue
		} else if strings.Contains(key, ".") {
			// Handle nested fields like "contacts.name"
			app.setNestedField(data, key, value)
		} else if strings.Contains(key, "[") && strings.Contains(key, "]") {
			// Handle array fields like "jobs[0].position" or "languages[0].name"
			app.setArrayField(data, key, value)
		} else {
			data[key] = value
		}
	}

	return data
}

func (app *application) setNestedField(data map[string]interface{}, key string, value interface{}) {
	parts := strings.Split(key, ".")
	current := data

	// Navigate to the parent
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if _, exists := current[part]; !exists {
			current[part] = make(map[string]interface{})
		}
		if nested, ok := current[part].(map[string]interface{}); ok {
			current = nested
		}
	}

	// Set the final value
	current[parts[len(parts)-1]] = value
}

func (app *application) isCommaSeparatedField(fieldName string) bool {
	commaSeparatedFields := []string{
		"skills", "tools", "methodology",
	}

	for _, field := range commaSeparatedFields {
		if fieldName == field {
			return true
		}
	}
	return false
}

func (app *application) parseCommaSeparatedValue(value string) []interface{} {
	if value == "" {
		return []interface{}{}
	}

	parts := strings.Split(value, ",")
	result := make([]interface{}, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func (app *application) setArrayField(data map[string]interface{}, key string, value interface{}) {
	// Parse array field like "jobs[0].position"
	start := strings.Index(key, "[")
	end := strings.Index(key, "]")
	if start == -1 || end == -1 {
		return
	}

	arrayName := key[:start]
	indexStr := key[start+1 : end]

	var index int
	if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
		return
	}

	remainder := key[end+1:]
	fieldName := strings.TrimPrefix(remainder, ".")

	// Ensure array exists
	if _, exists := data[arrayName]; !exists {
		data[arrayName] = make([]interface{}, 0)
	}

	// Ensure array is large enough
	arr := data[arrayName].([]interface{})
	for len(arr) <= index {
		arr = append(arr, make(map[string]interface{}))
	}
	data[arrayName] = arr

	// Set the field
	if item, ok := arr[index].(map[string]interface{}); ok {
		if fieldName == "" {
			arr[index] = value
		} else {
			item[fieldName] = value
		}
	}
}
