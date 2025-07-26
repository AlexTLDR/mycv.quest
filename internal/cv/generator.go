package cv

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Generator handles CV generation using Typst
type Generator struct {
	templatesDir string
	outputDir    string
}

// NewGenerator creates a new CV generator
func NewGenerator(templatesDir, outputDir string) *Generator {
	return &Generator{
		templatesDir: templatesDir,
		outputDir:    outputDir,
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
			templatePath := filepath.Join(g.templatesDir, entry.Name())
			configPath := filepath.Join(templatePath, "config.yaml")

			if _, err := os.Stat(configPath); err == nil {
				template, err := g.loadTemplate(entry.Name(), configPath)
				if err != nil {
					continue // Skip templates with invalid config
				}
				templates = append(templates, template)
			}
		}
	}

	return templates, nil
}

// GetTemplate returns a specific template by ID
func (g *Generator) GetTemplate(templateID string) (*Template, error) {
	configPath := filepath.Join(g.templatesDir, templateID, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("template %s not found", templateID)
	}

	template, err := g.loadTemplate(templateID, configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load template %s: %w", templateID, err)
	}

	return &template, nil
}

// GenerateCV generates a CV using the specified template and data
func (g *Generator) GenerateCV(request GenerationRequest) (*GenerationResult, error) {
	result := &GenerationResult{
		CreatedAt: time.Now(),
	}

	// Validate template exists
	_, err := g.GetTemplate(request.TemplateID)
	if err != nil {
		result.Message = err.Error()
		return result, err
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
	if err := g.writeDataFile(dataPath, request.Data); err != nil {
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
	if err := g.generateMainTypstFile(mainTypPath, request.Data); err != nil {
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
	filename := fmt.Sprintf("cv_%s_%s_%d.pdf",
		request.TemplateID,
		request.Data.Contacts.Name,
		time.Now().Unix())

	result.Success = true
	result.Filename = filename
	result.Data = data

	return result, nil
}

// loadTemplate loads a template configuration from file
func (g *Generator) loadTemplate(id, configPath string) (Template, error) {
	var template Template

	data, err := os.ReadFile(configPath)
	if err != nil {
		return template, err
	}

	if err := yaml.Unmarshal(data, &template); err != nil {
		return template, err
	}

	template.ID = id
	return template, nil
}

// writeDataFile writes CV data to a YAML file
func (g *Generator) writeDataFile(path string, data CVData) error {
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

		// Skip config.yaml as it's not needed for compilation
		if info.Name() == "config.yaml" {
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
func (g *Generator) generateMainTypstFile(path string, data CVData) error {
	const mainTemplate = `#import "template.typ": vantage-cv

#let data = yaml("data.yaml")

#vantage-cv(data)
`

	return os.WriteFile(path, []byte(mainTemplate), 0644)
}

// runTypstCompilation runs Typst to compile the CV
func (g *Generator) runTypstCompilation(inputPath, outputPath string) error {
	cmd := exec.Command("typst", "compile", inputPath, outputPath)
	cmd.Dir = filepath.Dir(inputPath)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("typst compilation failed: %v\nStderr: %s", err, stderr.String())
	}

	return nil
}
