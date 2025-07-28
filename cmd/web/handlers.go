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
			PreviewImage: "/static/templates/" + t.ID + "/preview.png",
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
			PreviewImage: "/static/templates/" + t.ID + "/preview.png",
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

func (app *application) getTemplateForm(w http.ResponseWriter, r *http.Request) {
	templateID := r.PathValue("id")
	if templateID == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Get the template form structure
	form, err := app.cvService.GenerateForm(templateID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(form); err != nil {
		app.serverError(w, r, err)
	}
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

	// Convert form data to template data
	templateData, err := app.convertFormToTemplateData(templateID, r)
	if err != nil {
		app.logger.Error("Failed to convert form data", "error", err)
		http.Error(w, "Failed to process form data: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Generate CV
	request := cv.GenerationRequest{
		TemplateID: templateID,
		Data:       *templateData,
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

	// Convert form data to template data
	templateData, err := app.convertFormToTemplateData(templateID, r)
	if err != nil {
		app.logger.Error("Failed to convert form data", "error", err)
		http.Error(w, "Failed to process form data: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Generate preview
	result, err := app.cvService.GeneratePreview(templateID, *templateData)
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

func (app *application) convertFormToTemplateData(templateID string, r *http.Request) (*cv.TemplateData, error) {
	data := make(map[string]interface{})

	// Parse all form values
	for key, values := range r.Form {
		if key == "template_id" {
			continue // Skip template_id as it's handled separately
		}

		if len(values) == 0 {
			continue
		}

		value := values[0]
		if value == "" {
			continue // Skip empty values
		}

		// Handle comma-separated fields that should become arrays
		if app.isCommaSeparatedField(key) {
			arrayValue := app.parseCommaSeparatedValue(value)
			if err := app.setNestedValue(data, key, arrayValue); err != nil {
				return nil, fmt.Errorf("failed to set field %s: %w", key, err)
			}
			continue
		}

		// Handle nested fields (e.g., "contacts.name", "jobs[0].position")
		if err := app.setNestedValue(data, key, value); err != nil {
			return nil, fmt.Errorf("failed to set field %s: %w", key, err)
		}
	}

	// Process arrays (work experience, education, etc.)
	if err := app.processFormArrays(data, r.Form); err != nil {
		return nil, fmt.Errorf("failed to process arrays: %w", err)
	}

	// Post-process specific fields for template compatibility
	app.postProcessTemplateData(data)

	templateData := &cv.TemplateData{
		TemplateID: templateID,
		Data:       data,
	}

	return templateData, nil
}

func (app *application) setNestedValue(data map[string]interface{}, key string, value interface{}) error {
	// Handle array notation like "jobs[0].position"
	if strings.Contains(key, "[") && strings.Contains(key, "]") {
		return nil // Arrays are handled separately
	}

	// Split key by dots for nested objects
	parts := strings.Split(key, ".")
	current := data

	// Navigate to the parent of the target field
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if _, exists := current[part]; !exists {
			current[part] = make(map[string]interface{})
		}
		if nested, ok := current[part].(map[string]interface{}); ok {
			current = nested
		} else {
			return fmt.Errorf("field %s is not an object", strings.Join(parts[:i+1], "."))
		}
	}

	// Set the final value
	finalKey := parts[len(parts)-1]
	current[finalKey] = value

	return nil
}

func (app *application) processFormArrays(data map[string]interface{}, form map[string][]string) error {
	arrays := make(map[string]map[int]map[string]interface{})

	// Collect array fields
	for key, values := range form {
		if !strings.Contains(key, "[") || !strings.Contains(key, "]") {
			continue
		}

		// Parse array field like "jobs[0].position"
		arrayName, index, fieldName, err := app.parseArrayField(key)
		if err != nil {
			continue // Skip malformed array fields
		}

		if _, exists := arrays[arrayName]; !exists {
			arrays[arrayName] = make(map[int]map[string]interface{})
		}
		if _, exists := arrays[arrayName][index]; !exists {
			arrays[arrayName][index] = make(map[string]interface{})
		}

		if len(values) > 0 {
			// Handle nested fields within array items
			if strings.Contains(fieldName, ".") {
				if err := app.setNestedValueInMap(arrays[arrayName][index], fieldName, values[0]); err != nil {
					return err
				}
			} else {
				arrays[arrayName][index][fieldName] = values[0]
			}
		}
	}

	// Convert maps to slices and add to data
	for arrayName, arrayMap := range arrays {
		var arraySlice []interface{}
		maxIndex := -1
		for index := range arrayMap {
			if index > maxIndex {
				maxIndex = index
			}
		}

		for i := 0; i <= maxIndex; i++ {
			if item, exists := arrayMap[i]; exists {
				arraySlice = append(arraySlice, item)
			} else {
				arraySlice = append(arraySlice, make(map[string]interface{}))
			}
		}

		data[arrayName] = arraySlice
	}

	return nil
}

func (app *application) parseArrayField(key string) (arrayName string, index int, fieldName string, err error) {
	// Parse field like "jobs[0].position" -> arrayName="jobs", index=0, fieldName="position"
	start := strings.Index(key, "[")
	end := strings.Index(key, "]")
	if start == -1 || end == -1 || end <= start {
		return "", 0, "", fmt.Errorf("invalid array field format")
	}

	arrayName = key[:start]
	indexStr := key[start+1 : end]

	var indexInt int
	if _, err := fmt.Sscanf(indexStr, "%d", &indexInt); err != nil {
		return "", 0, "", fmt.Errorf("invalid array index")
	}

	remainder := key[end+1:]
	if strings.HasPrefix(remainder, ".") {
		fieldName = remainder[1:]
	} else {
		fieldName = remainder
	}

	return arrayName, indexInt, fieldName, nil
}

func (app *application) setNestedValueInMap(data map[string]interface{}, key string, value interface{}) error {
	parts := strings.Split(key, ".")
	current := data

	// Navigate to the parent of the target field
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if _, exists := current[part]; !exists {
			current[part] = make(map[string]interface{})
		}
		if nested, ok := current[part].(map[string]interface{}); ok {
			current = nested
		} else {
			return fmt.Errorf("field %s is not an object", strings.Join(parts[:i+1], "."))
		}
	}

	// Set the final value
	finalKey := parts[len(parts)-1]
	current[finalKey] = value

	return nil
}

func (app *application) isCommaSeparatedField(fieldName string) bool {
	commaSeparatedFields := []string{
		"skills", "tools", "methodology", "achievements",
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

func (app *application) postProcessTemplateData(data map[string]interface{}) {
	// Convert jobs array tags from comma-separated strings to arrays
	if jobs, ok := data["jobs"].([]interface{}); ok {
		for _, job := range jobs {
			if jobMap, ok := job.(map[string]interface{}); ok {
				if tags, ok := jobMap["tags"].(string); ok && tags != "" {
					jobMap["tags"] = app.parseCommaSeparatedValue(tags)
				}

				// Convert description to array if it's a single string
				if desc, ok := jobMap["description"].([]interface{}); ok && len(desc) == 1 {
					if descStr, ok := desc[0].(string); ok && descStr != "" {
						// Split by newlines or keep as single item
						if strings.Contains(descStr, "\n") {
							lines := strings.Split(descStr, "\n")
							jobMap["description"] = make([]interface{}, 0, len(lines))
							for _, line := range lines {
								trimmed := strings.TrimSpace(line)
								if trimmed != "" {
									jobMap["description"] = append(jobMap["description"].([]interface{}), trimmed)
								}
							}
						}
					}
				}
			}
		}
	}

	// Set display text for social links if URLs are provided
	if contacts, ok := data["contacts"].(map[string]interface{}); ok {
		if linkedin, ok := contacts["linkedin"].(map[string]interface{}); ok {
			if url, ok := linkedin["url"].(string); ok && url != "" {
				if _, exists := linkedin["displayText"]; !exists {
					linkedin["displayText"] = "linkedin"
				}
			}
		}

		if github, ok := contacts["github"].(map[string]interface{}); ok {
			if url, ok := github["url"].(string); ok && url != "" {
				if _, exists := github["displayText"]; !exists {
					// Extract username from GitHub URL
					if strings.Contains(url, "github.com/") {
						parts := strings.Split(url, "github.com/")
						if len(parts) > 1 {
							username := strings.Split(parts[1], "/")[0]
							github["displayText"] = "@" + username
						}
					} else {
						github["displayText"] = "@username"
					}
				}
			}
		}
	}
}
