package generator

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlexTLDR/mycv.quest/pkg/config"
	"github.com/AlexTLDR/mycv.quest/pkg/utils"
	"github.com/AlexTLDR/mycv.quest/templates"
)

type CVGenerator struct {
	config *config.Config
}

func New(cfg *config.Config) *CVGenerator {
	return &CVGenerator{
		config: cfg,
	}
}

func (cv *CVGenerator) ListTemplates() {
	fmt.Println("Available templates:")
	for key, template := range cv.config.Templates {
		fmt.Printf("  %s: %s\n", key, template.Name)
	}
}

func (cv *CVGenerator) Generate(templateKey string) error {
	template, exists := cv.config.GetTemplate(templateKey)
	if !exists {
		return fmt.Errorf("template '%s' not found", templateKey)
	}

	if err := utils.EnsureDir(cv.config.OutputDir); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if template.NeedsPhoto {
		if err := cv.copyPhoto(template.Dir); err != nil {
			return fmt.Errorf("failed to copy photo for %s template: %w", template.Name, err)
		}
	}

	outputFile := filepath.Join(cv.config.OutputDir, fmt.Sprintf("cv-%s.pdf", templateKey))
	absOutputFile, _ := filepath.Abs(outputFile)

	if err := config.ValidateTemplateArgs(template, absOutputFile); err != nil {
		return fmt.Errorf("invalid template arguments: %w", err)
	}

	cmd := exec.Command("typst", "compile", template.InputFile, absOutputFile)
	cmd.Dir = template.Dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("typst compilation failed for %s: %w\nOutput: %s", template.Name, err, string(output))
	}

	fmt.Printf("CV generated successfully using %s template at %s/cv-%s.pdf\n", template.Name, cv.config.OutputDir, templateKey)
	return nil
}

func (cv *CVGenerator) copyPhoto(templateDir string) error {
	photoFiles, err := filepath.Glob("cv-photos/*")
	if err != nil {
		return fmt.Errorf("failed to find photos: %w", err)
	}

	if len(photoFiles) == 0 {
		return fmt.Errorf("no photos found in cv-photos/ directory")
	}

	sourcePhoto := photoFiles[0]

	if err := config.ValidatePath(sourcePhoto, "cv-photos"); err != nil {
		return fmt.Errorf("invalid source photo path: %w", err)
	}

	destPhoto := filepath.Join(templateDir, "avatar.png")

	if err := config.ValidatePath(templateDir, "templates"); err != nil {
		return fmt.Errorf("invalid template directory: %w", err)
	}

	source, err := os.Open(sourcePhoto)
	if err != nil {
		return fmt.Errorf("failed to open source photo: %w", err)
	}
	defer source.Close()

	dest, err := os.Create(destPhoto)
	if err != nil {
		return fmt.Errorf("failed to create destination photo: %w", err)
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	if err != nil {
		return fmt.Errorf("failed to copy photo: %w", err)
	}

	fmt.Printf("Copied photo %s to %s\n", sourcePhoto, destPhoto)
	return nil
}

func (cv *CVGenerator) GetTemplateData() []templates.CVTemplate {
	var templateData []templates.CVTemplate

	descriptions := map[string]string{
		"vantage": "Clean and professional design with modern typography",
		"basic":   "Simple and elegant layout perfect for any industry",
		"modern":  "Contemporary design with visual elements and photo support",
	}

	for key, template := range cv.config.Templates {
		pdfPath := fmt.Sprintf("/static/output/cv-%s.pdf", key)
		thumbnailPath := ""

		// Check for thumbnail images
		if _, err := os.Stat(filepath.Join(template.Dir, "thumbnail.png")); err == nil {
			thumbnailPath = fmt.Sprintf("/static/%s/thumbnail.png", template.Dir)
		} else if _, err := os.Stat(filepath.Join(template.Dir, "screenshot.png")); err == nil {
			thumbnailPath = fmt.Sprintf("/static/%s/screenshot.png", template.Dir)
		}

		templateData = append(templateData, templates.CVTemplate{
			Key:           key,
			Name:          template.Name,
			Description:   descriptions[key],
			PDFPath:       pdfPath,
			ThumbnailPath: thumbnailPath,
		})
	}

	return templateData
}

func (cv *CVGenerator) GenerateFromForm(templateKey string, r *http.Request) error {
	template, exists := cv.config.GetTemplate(templateKey)
	if !exists {
		return fmt.Errorf("template '%s' not found", templateKey)
	}

	if err := utils.EnsureDir(cv.config.OutputDir); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Parse form data - handle both multipart and regular forms
	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
			return fmt.Errorf("failed to parse multipart form: %w", err)
		}
	} else {
		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("failed to parse form: %w", err)
		}
	}

	// Generate template-specific files
	switch templateKey {
	case "basic":
		return cv.generateBasicCV(template, r)
	case "modern":
		return cv.generateModernCV(template, r)
	case "vantage":
		return cv.generateVantageCV(template, r)
	default:
		return fmt.Errorf("unsupported template: %s", templateKey)
	}
}

func (cv *CVGenerator) handlePhotoUpload(r *http.Request, templateDir string) error {
	file, _, err := r.FormFile("avatar")
	if err != nil {
		// No file uploaded, use existing avatar if available
		return nil
	}
	defer file.Close()

	// Save uploaded file as avatar.png in template directory
	avatarPath := filepath.Join(templateDir, "avatar.png")
	dest, err := os.Create(avatarPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, file)
	return err
}
