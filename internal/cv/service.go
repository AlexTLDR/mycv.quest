package cv

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// Template represents a CV template
type Template struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Version      string `json:"version"`
	Author       string `json:"author"`
	PreviewImage string `json:"preview_image"`
}

// GenerationRequest represents a CV generation request
type GenerationRequest struct {
	TemplateID string                 `json:"template_id"`
	Data       map[string]interface{} `json:"data"`
	Format     string                 `json:"format"`
}

// GenerationResult represents the result of CV generation
type GenerationResult struct {
	Success   bool      `json:"success"`
	Filename  string    `json:"filename"`
	Data      []byte    `json:"data"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// Service handles CV generation by running typst commands
type Service struct {
	templatesDir string
}

// NewService creates a new CV service
func NewService(templatesDir string) *Service {
	return &Service{
		templatesDir: templatesDir,
	}
}

// GetAvailableTemplates returns all available CV templates
func (s *Service) GetAvailableTemplates() ([]Template, error) {
	var templates []Template

	entries, err := os.ReadDir(s.templatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			templateID := entry.Name()

			// Check if this directory contains a valid template
			templateDir := filepath.Join(s.templatesDir, templateID)
			if s.isValidTemplate(templateDir) {
				template := Template{
					ID:           templateID,
					Name:         s.getTemplateName(templateID),
					Description:  s.getTemplateDescription(templateID),
					Version:      "1.0.0",
					Author:       "Unknown",
					PreviewImage: fmt.Sprintf("/static/templates/%s/preview.png", templateID),
				}
				templates = append(templates, template)
			}
		}
	}

	return templates, nil
}

// GenerateCV generates a CV using typst
func (s *Service) GenerateCV(request GenerationRequest) (*GenerationResult, error) {
	result := &GenerationResult{
		CreatedAt: time.Now(),
	}

	// Create temporary directory for generation
	tempDir, err := os.MkdirTemp("", "cv-generation-*")
	if err != nil {
		result.Message = "Failed to create temporary directory"
		return result, err
	}
	defer os.RemoveAll(tempDir)

	// Copy template files to temp directory
	templateDir := filepath.Join(s.templatesDir, request.TemplateID)
	if err := s.copyTemplateFiles(templateDir, tempDir); err != nil {
		result.Message = "Failed to copy template files"
		return result, err
	}

	// Create data file based on template type
	if err := s.createDataFile(tempDir, request.TemplateID, request.Data); err != nil {
		result.Message = "Failed to create data file"
		return result, err
	}

	// Find the main typst file
	mainFile, err := s.findMainTypstFile(tempDir)
	if err != nil {
		result.Message = "Failed to find main typst file"
		return result, err
	}

	// Run typst compilation
	outputPath := filepath.Join(tempDir, "output.pdf")
	if err := s.runTypstCompilation(mainFile, outputPath); err != nil {
		result.Message = fmt.Sprintf("Typst compilation failed: %v", err)
		return result, err
	}

	// Read generated file
	data, err := os.ReadFile(outputPath)
	if err != nil {
		result.Message = "Failed to read generated file"
		return result, err
	}

	// Generate filename
	filename := fmt.Sprintf("cv_%s_%d.pdf", request.TemplateID, time.Now().Unix())

	result.Success = true
	result.Filename = filename
	result.Data = data

	return result, nil
}

// isValidTemplate checks if a directory contains a valid typst template
func (s *Service) isValidTemplate(templateDir string) bool {
	// Look for main typst files
	possibleMainFiles := []string{"cv.typ", "template.typ", "main.typ"}

	for _, mainFile := range possibleMainFiles {
		if _, err := os.Stat(filepath.Join(templateDir, mainFile)); err == nil {
			return true
		}
	}

	return false
}

// getTemplateName returns a human-readable name for the template
func (s *Service) getTemplateName(templateID string) string {
	switch templateID {
	case "vantage":
		return "Vantage CV"
	case "grotesk":
		return "Grotesk CV"
	default:
		// Convert snake_case or kebab-case to Title Case
		words := strings.FieldsFunc(templateID, func(c rune) bool {
			return c == '_' || c == '-'
		})
		for i, word := range words {
			words[i] = strings.Title(strings.ToLower(word))
		}
		return strings.Join(words, " ") + " CV"
	}
}

// getTemplateDescription returns a description for the template
func (s *Service) getTemplateDescription(templateID string) string {
	switch templateID {
	case "vantage":
		return "ATS friendly simple Typst CV template"
	case "grotesk":
		return "Modern two-column CV template with photo support"
	default:
		return "Professional CV template"
	}
}

// copyTemplateFiles copies all template files to destination
func (s *Service) copyTemplateFiles(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return s.copyFile(path, dstPath)
	})
}

// copyFile copies a single file
func (s *Service) copyFile(src, dst string) error {
	srcData, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, srcData, 0644)
}

// createTOMLDataFile creates a TOML data file
func (s *Service) createTOMLDataFile(tempDir string, data map[string]interface{}) error {
	dataPath := filepath.Join(tempDir, "info.toml")

	// Convert form data to grotesk format
	groteskData := s.convertToGroteskFormat(data)

	file, err := os.Create(dataPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return toml.NewEncoder(file).Encode(groteskData)
}

// createDataFile creates the appropriate data file for the template
func (s *Service) createDataFile(tempDir, templateID string, data map[string]interface{}) error {
	switch templateID {
	case "grotesk":
		// Grotesk uses TOML
		return s.createTOMLDataFile(tempDir, data)
	case "vantage":
		// Vantage uses YAML
		return s.createYAMLDataFile(tempDir, data)
	default:
		// Default to YAML
		return s.createYAMLDataFile(tempDir, data)
	}
}



// createYAMLDataFile creates a YAML data file
func (s *Service) createYAMLDataFile(tempDir string, data map[string]interface{}) error {
	dataPath := filepath.Join(tempDir, "data.yaml")

	file, err := os.Create(dataPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	return encoder.Encode(data)
}

// convertToGroteskFormat converts form data to grotesk template format
func (s *Service) convertToGroteskFormat(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Personal information
	personal := map[string]interface{}{
		"first_name":     "John",
		"last_name":      "Doe",
		"profile_image":  "portrait.png",
		"language":       "en",
		"include_icons":  false,
		"use_photo":      true,
	}

	// Extract name
	if contacts, ok := data["contacts"].(map[string]interface{}); ok {
		if name, ok := contacts["name"].(string); ok && name != "" {
			parts := strings.Fields(strings.TrimSpace(name))
			if len(parts) >= 2 {
				personal["first_name"] = parts[0]
				personal["last_name"] = strings.Join(parts[1:], " ")
			} else if len(parts) == 1 {
				personal["first_name"] = parts[0]
			}
		}
	}

	// Handle use_photo setting
	if usePhoto, ok := data["use_photo"].(bool); ok {
		personal["use_photo"] = usePhoto
	} else if usePhotoStr, ok := data["use_photo"].(string); ok {
		personal["use_photo"] = usePhotoStr == "true" || usePhotoStr == "on"
	}

	// Add contact info structure
	info := make(map[string]interface{})

	// Extract contact info from form data
	if contacts, ok := data["contacts"].(map[string]interface{}); ok {
		if email, ok := contacts["email"].(string); ok && email != "" {
			info["email"] = map[string]interface{}{
				"link":  "mailto:" + email,
				"label": email,
			}
		}

		if address, ok := contacts["address"].(string); ok && address != "" {
			info["address"] = address
		} else if location, ok := contacts["location"].(string); ok && location != "" {
			info["address"] = location
		}

		if phone, ok := contacts["phone"].(string); ok && phone != "" {
			info["telephone"] = phone
		}

		if linkedin, ok := contacts["linkedin"].(map[string]interface{}); ok {
			if url, ok := linkedin["url"].(string); ok && url != "" {
				displayText := "linkedin"
				if dt, ok := linkedin["displayText"]; ok {
					displayText = fmt.Sprintf("%v", dt)
				}
				info["linkedin"] = map[string]interface{}{
					"link":  url,
					"label": displayText,
				}
			}
		}

		if github, ok := contacts["github"].(map[string]interface{}); ok {
			if url, ok := github["url"].(string); ok && url != "" {
				displayText := "@username"
				if dt, ok := github["displayText"]; ok {
					displayText = fmt.Sprintf("%v", dt)
				}
				info["github"] = map[string]interface{}{
					"link":  url,
					"label": displayText,
				}
			}
		}
	}

	personal["info"] = info
	personal["icon"] = map[string]interface{}{
		"address":   "house",
		"telephone": "phone",
		"email":     "envelope",
		"linkedin":  "linkedin",
		"github":    "github",
		"homepage":  "globe",
	}

	// Add IA (AI) settings - required by grotesk template but we'll disable them
	personal["ia"] = map[string]interface{}{
		"inject_ai_prompt": false,
		"inject_keywords":  false,
		"keywords_list":    []interface{}{},
	}

	result["personal"] = personal

	// Extract summary/tagline
	if tagline, ok := data["tagline"].(string); ok && tagline != "" {
		result["summary"] = tagline
	} else if summary, ok := data["summary"].(string); ok && summary != "" {
		result["summary"] = summary
	}

	// Copy experience data (jobs -> experience for grotesk)
	if jobs, ok := data["jobs"]; ok {
		result["experience"] = jobs
	}

	// Copy education data
	if education, ok := data["education"]; ok {
		result["education"] = education
	}

	// Copy skills data
	if skills, ok := data["skills"]; ok {
		result["skills"] = skills
	}

	// Copy languages data
	if languages, ok := data["languages"]; ok {
		result["languages"] = languages
	}

	// Add required import settings
	result["import"] = map[string]interface{}{
		"fontawesome": "@preview/fontawesome:0.5.0",
	}

	// Add section icon settings
	result["section"] = map[string]interface{}{
		"icon": map[string]interface{}{
			"profile":          "id-card",
			"experience":       "briefcase",
			"education":        "graduation-cap",
			"skills":           "cogs",
			"languages":        "language",
			"other_experience": "wrench",
			"references":       "users",
			"personal":         "brain",
		},
	}

	// Add layout settings with all required nested structures
	result["layout"] = map[string]interface{}{
		"accent_color":      "#d4d2cc",
		"fill_color":        "#f4f1eb",
		"left_pane_width":   "71%",
		"paper_size":        "a4",
		"text": map[string]interface{}{
			"font":              "Arial", // Use Arial instead of HK Grotesk to avoid font issues
			"size":              "10pt",
			"cover_letter_size": "11pt",
			"color": map[string]interface{}{
				"light":  "#ededef",
				"medium": "#78787e",
				"dark":   "#3c3c42",
			},
		},
	}

	// Add language settings required by grotesk template
	result["language"] = map[string]interface{}{
		"en": map[string]interface{}{
			"subtitle":                   "Software Engineer",
			"ai_prompt":                  "", // Empty AI prompt to disable the feature
			"cv_document_name":           "Resume",
			"cover_letter_document_name": "Cover letter",
		},
		"es": map[string]interface{}{
			"subtitle":                   "Ingeniero de Software",
			"ai_prompt":                  "", // Empty AI prompt to disable the feature
			"cv_document_name":           "CV",
			"cover_letter_document_name": "Carta de motivación",
		},
	}

	return result
}

// findMainTypstFile finds the main typst file to compile
func (s *Service) findMainTypstFile(tempDir string) (string, error) {
	possibleMainFiles := []string{"main.typ", "cv.typ", "template.typ"}

	for _, mainFile := range possibleMainFiles {
		path := filepath.Join(tempDir, mainFile)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no main typst file found")
}

// runTypstCompilation runs typst to compile the CV
func (s *Service) runTypstCompilation(inputPath, outputPath string) error {
	cmd := exec.Command("typst", "compile", inputPath, outputPath)
	cmd.Dir = filepath.Dir(inputPath)

	var stderr bytes.Buffer
	var stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("typst compilation failed: %v\nStderr: %s\nStdout: %s",
			err, stderr.String(), stdout.String())
	}

	return nil
}
