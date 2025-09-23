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

const (
	DefaultAvatarFilename = "avatar.png"
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
		if err := cv.CopyPhoto(template.Dir); err != nil {
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

func (cv *CVGenerator) GetTemplateData() []templates.CVTemplate {
	var templateData []templates.CVTemplate

	descriptions := map[string]string{
		"vantage": "Clean and professional design with modern typography",
		"basic":   "Simple and elegant layout perfect for any industry",
		"modern":  "Contemporary design with visual elements and photo support",
	}

	// Map template keys to their example PDF paths
	examplePDFs := map[string]string{
		"vantage": "/static/templates/vantage/example.pdf",
		"basic":   "/static/templates/basic/example-resume.pdf",
		"modern":  "/static/templates/modern/template/test.pdf",
	}

	for key, template := range cv.config.Templates {
		// Use example PDF for preview instead of generated CV
		pdfPath := examplePDFs[key]
		if pdfPath == "" {
			// Fallback if no example PDF is found
			pdfPath = fmt.Sprintf("/static/templates/%s/example.pdf", key)
		}

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

func (cv *CVGenerator) GenerateFromForm(templateKey string, r *http.Request) ([]byte, error) {
	template, exists := cv.config.GetTemplate(templateKey)
	if !exists {
		return nil, fmt.Errorf("template '%s' not found", templateKey)
	}

	// Parse form data - handle both multipart and regular forms
	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
			return nil, fmt.Errorf("failed to parse multipart form: %w", err)
		}
	} else {
		if err := r.ParseForm(); err != nil {
			return nil, fmt.Errorf("failed to parse form: %w", err)
		}
	}

	// Generate template-specific files and return PDF data
	switch templateKey {
	case "basic":
		return cv.GenerateBasicCV(template, r)
	case "modern":
		return cv.GenerateModernCV(template, r)
	case "vantage":
		return cv.GenerateVantageCV(template, r)
	default:
		return nil, fmt.Errorf("unsupported template: %s", templateKey)
	}
}

func (cv *CVGenerator) CopyPhoto(templateDir string) error {
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

func (cv *CVGenerator) HandlePhotoUploadToWorkDir(r *http.Request, workDir string) (string, error) {
	file, _, err := r.FormFile("avatar")
	if err != nil {
		// No file uploaded, that's okay
		return "", nil
	}
	defer file.Close()

	// Read first few bytes to detect file format
	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil {
		return "", err
	}

	// Reset file pointer to beginning
	if seeker, ok := file.(io.Seeker); ok {
		_, err = seeker.Seek(0, 0)
		if err != nil {
			return "", err
		}
	}

	// Detect file format from header
	var filename string
	switch {
	case n >= 8 && header[0] == 0x89 && header[1] == 0x50 && header[2] == 0x4E && header[3] == 0x47:
		// PNG format
		filename = DefaultAvatarFilename
	case n >= 3 && header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF:
		// JPEG format
		filename = "avatar.jpg"
	default:
		// Default to PNG if format not recognized
		filename = DefaultAvatarFilename
	}

	// Save uploaded file with detected extension
	avatarPath := filepath.Join(workDir, filename)
	dest, err := os.Create(avatarPath)
	if err != nil {
		return "", err
	}
	defer dest.Close()

	_, err = io.Copy(dest, file)
	return filename, err
}
