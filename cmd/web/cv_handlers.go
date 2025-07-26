package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AlexTLDR/mycv.quest/internal/cv"
	"github.com/AlexTLDR/mycv.quest/internal/response"
)

// initializeCVService initializes the CV service
func (app *application) initializeCVService() error {
	templatesDir := filepath.Join("assets", "templates", "typst")
	outputDir := filepath.Join("tmp", "cv_output")

	cvService, err := cv.NewService(templatesDir, outputDir)
	if err != nil {
		return fmt.Errorf("failed to initialize CV service: %w", err)
	}

	app.cvService = cvService
	return nil
}

// listTemplates returns all available CV templates
func (app *application) listTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := app.cvService.ListTemplates()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
	})
}

// getTemplate returns details for a specific template
func (app *application) getTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := r.PathValue("id")
	if templateID == "" {
		response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "template ID is required",
		})
		return
	}

	template, err := app.cvService.GetTemplate(templateID)
	if err != nil {
		response.JSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("template not found: %v", err),
		})
		return
	}

	response.JSON(w, http.StatusOK, template)
}

// getTemplateForm returns the form structure for a template
func (app *application) getTemplateForm(w http.ResponseWriter, r *http.Request) {
	templateID := r.PathValue("id")
	if templateID == "" {
		response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "template ID is required",
		})
		return
	}

	form, err := app.cvService.GenerateForm(templateID)
	if err != nil {
		response.JSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("failed to generate form: %v", err),
		})
		return
	}

	response.JSON(w, http.StatusOK, form)
}

// getTemplateSample returns sample data for a template
func (app *application) getTemplateSample(w http.ResponseWriter, r *http.Request) {
	templateID := r.PathValue("id")
	if templateID == "" {
		response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "template ID is required",
		})
		return
	}

	sampleData, err := app.cvService.GetSampleData(templateID)
	if err != nil {
		response.JSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("failed to generate sample data: %v", err),
		})
		return
	}

	response.JSON(w, http.StatusOK, sampleData)
}

// validateTemplateData validates data against a template schema
func (app *application) validateTemplateData(w http.ResponseWriter, r *http.Request) {
	templateID := r.PathValue("id")
	if templateID == "" {
		response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "template ID is required",
		})
		return
	}

	var templateData cv.TemplateData
	if err := json.NewDecoder(r.Body).Decode(&templateData); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("invalid JSON: %v", err),
		})
		return
	}

	// Ensure template ID matches
	templateData.TemplateID = templateID

	validation, err := app.cvService.ValidateData(templateID, templateData)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("validation failed: %v", err),
		})
		return
	}

	response.JSON(w, http.StatusOK, validation)
}

// generateCV generates a CV PDF
func (app *application) generateCV(w http.ResponseWriter, r *http.Request) {
	templateID := r.PathValue("id")
	if templateID == "" {
		response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "template ID is required",
		})
		return
	}

	var templateData cv.TemplateData
	if err := json.NewDecoder(r.Body).Decode(&templateData); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("invalid JSON: %v", err),
		})
		return
	}

	// Ensure template ID matches
	templateData.TemplateID = templateID

	// Get format from query parameter (default to PDF)
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "pdf"
	}

	request := cv.GenerationRequest{
		TemplateID: templateID,
		Data:       templateData,
		Format:     format,
	}

	result, err := app.cvService.GenerateCV(request)
	if err != nil {
		app.logger.Error("CV generation failed", "error", err, "template", templateID)
		response.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("CV generation failed: %v", err),
		})
		return
	}

	if !result.Success {
		response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": result.Message,
		})
		return
	}

	// Set appropriate headers for file download
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", result.Filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(result.Data)))

	// Write the PDF data
	w.WriteHeader(http.StatusOK)
	w.Write(result.Data)
}

// generateCVPreview generates a preview of the CV
func (app *application) generateCVPreview(w http.ResponseWriter, r *http.Request) {
	templateID := r.PathValue("id")
	if templateID == "" {
		response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "template ID is required",
		})
		return
	}

	var templateData cv.TemplateData
	if err := json.NewDecoder(r.Body).Decode(&templateData); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("invalid JSON: %v", err),
		})
		return
	}

	// Ensure template ID matches
	templateData.TemplateID = templateID

	result, err := app.cvService.GeneratePreview(templateID, templateData)
	if err != nil {
		app.logger.Error("CV preview generation failed", "error", err, "template", templateID)
		response.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("preview generation failed: %v", err),
		})
		return
	}

	if !result.Success {
		response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": result.Message,
		})
		return
	}

	// For now, return PDF as preview (in the future, could be PNG)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Length", strconv.Itoa(len(result.Data)))

	w.WriteHeader(http.StatusOK)
	w.Write(result.Data)
}

// getTemplateMetadata returns metadata for a template
func (app *application) getTemplateMetadata(w http.ResponseWriter, r *http.Request) {
	templateID := r.PathValue("id")
	if templateID == "" {
		response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "template ID is required",
		})
		return
	}

	metadata, err := app.cvService.GetTemplateMetadata(templateID)
	if err != nil {
		response.JSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("metadata not found: %v", err),
		})
		return
	}

	response.JSON(w, http.StatusOK, metadata)
}

// quickGenerate generates a CV using sample data (for testing)
func (app *application) quickGenerate(w http.ResponseWriter, r *http.Request) {
	templateID := r.PathValue("id")
	if templateID == "" {
		response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "template ID is required",
		})
		return
	}

	// Get sample data
	sampleData, err := app.cvService.GetSampleData(templateID)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("failed to get sample data: %v", err),
		})
		return
	}

	// Generate CV with sample data
	request := cv.GenerationRequest{
		TemplateID: templateID,
		Data:       *sampleData,
		Format:     "pdf",
	}

	result, err := app.cvService.GenerateCV(request)
	if err != nil {
		app.logger.Error("Quick CV generation failed", "error", err, "template", templateID)
		response.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("CV generation failed: %v", err),
		})
		return
	}

	if !result.Success {
		response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": result.Message,
		})
		return
	}

	// Set appropriate headers for file download
	filename := fmt.Sprintf("sample_%s_cv.pdf", templateID)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(result.Data)))

	w.WriteHeader(http.StatusOK)
	w.Write(result.Data)
}

// convertData converts data from one template format to another
func (app *application) convertData(w http.ResponseWriter, r *http.Request) {
	fromTemplate := r.URL.Query().Get("from")
	toTemplate := r.URL.Query().Get("to")

	if fromTemplate == "" || toTemplate == "" {
		response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "both 'from' and 'to' template parameters are required",
		})
		return
	}

	var templateData cv.TemplateData
	if err := json.NewDecoder(r.Body).Decode(&templateData); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("invalid JSON: %v", err),
		})
		return
	}

	// For now, return the data as-is with new template ID
	// In the future, implement smart conversion between template formats
	convertedData := app.cvService.CloneTemplateData(templateData)
	convertedData.TemplateID = toTemplate

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"message":        "Data converted successfully",
		"from_template":  fromTemplate,
		"to_template":    toTemplate,
		"converted_data": convertedData,
		"note":           "Basic conversion applied. Manual adjustments may be needed.",
	})
}

// validateTypstInstallation checks if Typst is available
func (app *application) validateTypstInstallation(w http.ResponseWriter, r *http.Request) {
	// Try to run typst --version

	cmd := exec.Command("typst", "--version")
	output, err := cmd.Output()

	if err != nil {
		response.JSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"typst_available": false,
			"error":           "Typst not found or not accessible",
			"message":         "Please install Typst CLI to use CV generation features",
			"install_url":     "https://github.com/typst/typst/releases",
		})
		return
	}

	version := strings.TrimSpace(string(output))
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"typst_available": true,
		"version":         version,
		"message":         "Typst is properly installed and accessible",
	})
}

// debugTemplate provides debugging information for a template
func (app *application) debugTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := r.PathValue("id")
	if templateID == "" {
		response.JSON(w, http.StatusBadRequest, map[string]string{
			"error": "template ID is required",
		})
		return
	}

	// Get template configuration
	config, err := app.cvService.GetTemplateConfig(templateID)
	if err != nil {
		response.JSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("template not found: %v", err),
		})
		return
	}

	// Get template files
	files, err := app.cvService.ListTemplateFiles(templateID)
	if err != nil {
		files = []string{"Error listing files: " + err.Error()}
	}

	// Get required fields
	requiredFields, err := app.cvService.GetRequiredFields(templateID)
	if err != nil {
		requiredFields = []string{"Error getting required fields: " + err.Error()}
	}

	// Get sample data
	sampleData, err := app.cvService.GetSampleData(templateID)
	var sampleDataPreview interface{} = "Error generating sample data: " + err.Error()
	if err == nil {
		sampleDataPreview = sampleData
	}

	debugInfo := map[string]interface{}{
		"template_id":        templateID,
		"config":             config,
		"files":              files,
		"required_fields":    requiredFields,
		"sample_data":        sampleDataPreview,
		"field_count":        len(config.Fields),
		"template_available": app.cvService.IsTemplateAvailable(templateID),
	}

	response.JSON(w, http.StatusOK, debugInfo)
}
