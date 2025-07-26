package cv

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Generator handles CV generation using Typst with dynamic template support
type Generator struct {
	templatesDir string
	outputDir    string
	parser       *TemplateParser
	adapters     *AdapterRegistry
}

// NewGenerator creates a new CV generator
func NewGenerator(templatesDir, outputDir string) *Generator {
	return &Generator{
		templatesDir: templatesDir,
		outputDir:    outputDir,
		parser:       NewTemplateParser(templatesDir),
		adapters:     NewAdapterRegistry(),
	}
}

// GetAvailableTemplates returns all available CV templates
func (g *Generator) GetAvailableTemplates() ([]Template, error) {
	var templates []Template

	entries, err := os.ReadDir(g.templatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			config, err := g.parser.ParseTemplate(entry.Name())
			if err != nil {
				continue // Skip templates with invalid config
			}
			templates = append(templates, config.Template)
		}
	}

	return templates, nil
}

// GetTemplate returns a specific template by ID
func (g *Generator) GetTemplate(templateID string) (*Template, error) {
	config, err := g.parser.ParseTemplate(templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to load template %s: %w", templateID, err)
	}

	return &config.Template, nil
}

// GetTemplateConfig returns the full template configuration
func (g *Generator) GetTemplateConfig(templateID string) (*TemplateConfig, error) {
	return g.parser.ParseTemplate(templateID)
}

// GenerateForm creates a form structure for a template
func (g *Generator) GenerateForm(templateID string) (*TemplateForm, error) {
	config, err := g.parser.ParseTemplate(templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to load template config: %w", err)
	}

	return g.parser.GenerateForm(config)
}

// ValidateData validates template data against its configuration
func (g *Generator) ValidateData(templateID string, data TemplateData) (ValidationResult, error) {
	config, err := g.parser.ParseTemplate(templateID)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("failed to load template config: %w", err)
	}

	return g.parser.ValidateData(config, data), nil
}

// GenerateCV generates a CV using the specified template and data
func (g *Generator) GenerateCV(request GenerationRequest) (*GenerationResult, error) {
	result := &GenerationResult{
		CreatedAt: time.Now(),
	}

	// Load template configuration
	config, err := g.parser.ParseTemplate(request.TemplateID)
	if err != nil {
		result.Message = err.Error()
		return result, err
	}

	// Convert data to template format if adapter exists
	adaptedData := request.Data
	if adapter, exists := g.adapters.GetAdapter(request.TemplateID); exists {
		convertedData, err := adapter.ConvertToTemplate(request.Data)
		if err != nil {
			result.Message = fmt.Sprintf("Data conversion failed: %v", err)
			return result, err
		}
		adaptedData = convertedData
	}

	// Validate adapted data
	validation := g.parser.ValidateData(config, adaptedData)
	if !validation.Valid {
		var errorMessages []string
		for _, err := range validation.Errors {
			errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", err.Field, err.Message))
		}
		result.Message = "Validation failed: " + strings.Join(errorMessages, "; ")
		return result, fmt.Errorf("data validation failed")
	}

	// Create temporary directory for generation
	tempDir, err := os.MkdirTemp("", "cv-generation-*")
	if err != nil {
		result.Message = "Failed to create temporary directory"
		return result, err
	}
	defer os.RemoveAll(tempDir)

	// Generate data file (YAML)
	dataPath := filepath.Join(tempDir, "data.yaml")
	if err := g.writeDataFile(dataPath, adaptedData.Data); err != nil {
		result.Message = "Failed to write data file"
		return result, err
	}

	// Copy template files to temp directory
	templateDir := filepath.Join(g.templatesDir, request.TemplateID)
	if err := g.copyTemplateFiles(templateDir, tempDir); err != nil {
		result.Message = "Failed to copy template files"
		return result, err
	}

	// Generate main Typst file
	mainTypPath := filepath.Join(tempDir, "main.typ")
	if err := g.generateMainTypstFile(mainTypPath, config); err != nil {
		result.Message = "Failed to generate main Typst file"
		return result, err
	}

	// Run Typst compilation
	outputPath := filepath.Join(tempDir, "output.pdf")
	if err := g.runTypstCompilation(mainTypPath, outputPath); err != nil {
		result.Message = fmt.Sprintf("Typst compilation failed: %v", err)
		return result, err
	}

	// Read generated file
	data, err := os.ReadFile(outputPath)
	if err != nil {
		result.Message = "Failed to read generated file"
		return result, err
	}

	// Generate unique filename
	displayName := g.parser.GetDisplayName(config, adaptedData)
	filename := fmt.Sprintf("cv_%s_%s_%d.pdf",
		request.TemplateID,
		strings.ReplaceAll(strings.ToLower(displayName), " ", "_"),
		time.Now().Unix())

	result.Success = true
	result.Filename = filename
	result.Data = data

	return result, nil
}

// writeDataFile writes template data to a YAML file
func (g *Generator) writeDataFile(path string, data map[string]interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	return encoder.Encode(data)
}

// copyTemplateFiles copies template files to destination
func (g *Generator) copyTemplateFiles(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip configuration files as they're not needed for compilation
		filename := info.Name()
		if filename == "config.yaml" || filename == "config.yml" ||
			filename == "config.toml" || filename == "info.toml" {
			return nil
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return g.copyFile(path, dstPath)
	})
}

// copyFile copies a single file
func (g *Generator) copyFile(src, dst string) error {
	srcData, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, srcData, 0644)
}

// generateMainTypstFile generates the main Typst file that imports template and uses data
func (g *Generator) generateMainTypstFile(path string, config *TemplateConfig) error {
	// Determine the main function name
	mainFunction := config.MainFunction
	if mainFunction == "" {
		mainFunction = config.ID + "-cv"
	}

	// Generate import statement - try both template.typ and the template name
	importPath := "template.typ"
	templateTypPath := filepath.Join(g.templatesDir, config.ID, config.ID+".typ")
	if _, err := os.Stat(templateTypPath); err == nil {
		importPath = config.ID + ".typ"
	}

	// Check if it's a package import (for Typst Universe packages)
	if strings.Contains(config.Version, "@preview/") ||
		strings.Contains(config.Name, "@preview/") {
		// This is a package, use package import
		packageName := config.ID
		if strings.Contains(config.Name, "@preview/") {
			packageName = strings.Split(config.Name, "/")[1]
		}

		mainTemplate := fmt.Sprintf(`#import "@preview/%s:%s": %s

#let data = yaml("data.yaml")

#%s(data)
`, packageName, config.Version, mainFunction, mainFunction)

		return os.WriteFile(path, []byte(mainTemplate), 0644)
	}

	// Standard template import
	mainTemplate := fmt.Sprintf(`#import "%s": %s

#let data = yaml("data.yaml")

#%s(data)
`, importPath, mainFunction, mainFunction)

	return os.WriteFile(path, []byte(mainTemplate), 0644)
}

// runTypstCompilation runs Typst to compile the CV
func (g *Generator) runTypstCompilation(inputPath, outputPath string) error {
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

// GeneratePreview generates a preview image of the CV
func (g *Generator) GeneratePreview(templateID string, data TemplateData) (*GenerationResult, error) {
	// Similar to GenerateCV but compile to PNG instead
	request := GenerationRequest{
		TemplateID: templateID,
		Data:       data,
		Format:     "png",
	}

	// For now, just generate PDF and return it
	// TODO: Add PNG generation support when Typst supports it
	return g.GenerateCV(request)
}

// GetTemplateMetadata returns metadata for a template
func (g *Generator) GetTemplateMetadata(templateID string) (*TemplateMetadata, error) {
	_, err := g.parser.ParseTemplate(templateID)
	if err != nil {
		return nil, err
	}

	templateDir := filepath.Join(g.templatesDir, templateID)
	stat, err := os.Stat(templateDir)
	if err != nil {
		return nil, err
	}

	metadata := &TemplateMetadata{
		TemplateID:   templateID,
		LastModified: stat.ModTime(),
		Tags:         []string{}, // Could be extracted from config if available
	}

	// Look for sample image
	samplePaths := []string{"sample.png", "preview.png", "example.png", "sample.jpg", "preview.jpg"}
	for _, samplePath := range samplePaths {
		fullPath := filepath.Join(templateDir, samplePath)
		if _, err := os.Stat(fullPath); err == nil {
			metadata.SampleImageURL = fmt.Sprintf("/static/templates/%s/%s", templateID, samplePath)
			break
		}
	}

	return metadata, nil
}

// ListTemplateFiles returns a list of files in a template directory
func (g *Generator) ListTemplateFiles(templateID string) ([]string, error) {
	templateDir := filepath.Join(g.templatesDir, templateID)
	var files []string

	err := filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relPath, err := filepath.Rel(templateDir, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}

		return nil
	})

	return files, err
}

// GetSampleData generates sample data for a template
func (g *Generator) GetSampleData(templateID string) (*TemplateData, error) {
	config, err := g.parser.ParseTemplate(templateID)
	if err != nil {
		return nil, err
	}

	// Generate sample data based on field definitions
	sampleData := make(map[string]interface{})
	for fieldName, fieldDef := range config.Fields {
		sampleValue := g.generateSampleValue(fieldDef)
		if sampleValue != nil {
			sampleData[fieldName] = sampleValue
		}
	}

	return &TemplateData{
		TemplateID: templateID,
		Data:       sampleData,
	}, nil
}

// generateSampleValue creates a sample value for a field definition
func (g *Generator) generateSampleValue(fieldDef FieldDefinition) interface{} {
	if fieldDef.Default != nil {
		return fieldDef.Default
	}

	switch fieldDef.Type {
	case "string":
		if len(fieldDef.Options) > 0 {
			return fieldDef.Options[0]
		}
		return "Sample " + fieldDef.Label
	case "text":
		return "This is a sample " + strings.ToLower(fieldDef.Label) + " text."
	case "integer":
		if fieldDef.Min != nil {
			return *fieldDef.Min
		}
		return 1
	case "boolean":
		return true
	case "array":
		if fieldDef.Items != nil {
			sample := g.generateSampleValue(*fieldDef.Items)
			return []interface{}{sample}
		}
		return []string{"Sample Item"}
	case "object":
		if fieldDef.Fields != nil {
			obj := make(map[string]interface{})
			for subFieldName, subFieldDef := range fieldDef.Fields {
				obj[subFieldName] = g.generateSampleValue(subFieldDef)
			}
			return obj
		}
		return map[string]interface{}{}
	default:
		return nil
	}
}
