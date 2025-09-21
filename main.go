package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/AlexTLDR/mycv.quest/templates"
)

type Template struct {
	Name       string
	Dir        string
	InputFile  string
	NeedsPhoto bool
}

type CVGenerator struct {
	Templates map[string]Template
	OutputDir string
}

func validatePath(path, expectedPrefix string) error {
	cleanPath := filepath.Clean(path)

	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path contains directory traversal: %s", path)
	}

	if expectedPrefix != "" {
		absPath, err := filepath.Abs(cleanPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}

		absPrefix, err := filepath.Abs(expectedPrefix)
		if err != nil {
			return fmt.Errorf("failed to get absolute prefix: %w", err)
		}

		if !strings.HasPrefix(absPath, absPrefix) {
			return fmt.Errorf("path %s is outside expected directory %s", path, expectedPrefix)
		}
	}

	return nil
}

func validateTemplateArgs(template Template, outputFile string) error {
	allowedInputFiles := []string{"example.typ", "main.typ"}
	if !slices.Contains(allowedInputFiles, template.InputFile) {
		return fmt.Errorf("invalid input file: %s", template.InputFile)
	}

	if !strings.HasSuffix(outputFile, ".pdf") {
		return fmt.Errorf("output file must have .pdf extension")
	}

	if err := validatePath(outputFile, ""); err != nil {
		return fmt.Errorf("invalid output file path: %w", err)
	}

	return nil
}

func NewCVGenerator() *CVGenerator {
	templates := map[string]Template{
		"vantage": {
			Name:       "Vantage",
			Dir:        "templates/vantage",
			InputFile:  "example.typ",
			NeedsPhoto: false,
		},
		"basic": {
			Name:       "Basic Resume",
			Dir:        "templates/basic/template",
			InputFile:  "main.typ",
			NeedsPhoto: false,
		},
		"modern": {
			Name:       "Modern Resume",
			Dir:        "templates/modern/template",
			InputFile:  "main.typ",
			NeedsPhoto: true,
		},
	}

	return &CVGenerator{
		Templates: templates,
		OutputDir: "output",
	}
}

func (cv *CVGenerator) ListTemplates() {
	fmt.Println("Available templates:")
	for key, template := range cv.Templates {
		fmt.Printf("  %s: %s\n", key, template.Name)
	}
}

func (cv *CVGenerator) Generate(templateKey string) error {
	template, exists := cv.Templates[templateKey]
	if !exists {
		return fmt.Errorf("template '%s' not found", templateKey)
	}

	if err := cv.ensureOutputDir(); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if template.NeedsPhoto {
		if err := cv.copyPhoto(template.Dir); err != nil {
			return fmt.Errorf("failed to copy photo for %s template: %w", template.Name, err)
		}
	}

	outputFile := filepath.Join(cv.OutputDir, fmt.Sprintf("cv-%s.pdf", templateKey))
	absOutputFile, _ := filepath.Abs(outputFile)

	if err := validateTemplateArgs(template, absOutputFile); err != nil {
		return fmt.Errorf("invalid template arguments: %w", err)
	}

	cmd := exec.Command("typst", "compile", template.InputFile, absOutputFile)
	cmd.Dir = template.Dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("typst compilation failed for %s: %w\nOutput: %s", template.Name, err, string(output))
	}

	fmt.Printf("CV generated successfully using %s template at %s/cv-%s.pdf\n", template.Name, cv.OutputDir, templateKey)
	return nil
}

func (cv *CVGenerator) ensureOutputDir() error {
	return os.MkdirAll(cv.OutputDir, 0o750)
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

	if err := validatePath(sourcePhoto, "cv-photos"); err != nil {
		return fmt.Errorf("invalid source photo path: %w", err)
	}

	destPhoto := filepath.Join(templateDir, "avatar.png")

	if err := validatePath(templateDir, "templates"); err != nil {
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

func (cv *CVGenerator) getTemplateData() []templates.CVTemplate {
	var templateData []templates.CVTemplate

	descriptions := map[string]string{
		"vantage": "Clean and professional design with modern typography",
		"basic":   "Simple and elegant layout perfect for any industry",
		"modern":  "Contemporary design with visual elements and photo support",
	}

	for key, template := range cv.Templates {
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

func (cv *CVGenerator) setupRoutes() {
	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("."))))

	// Home page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		templateData := cv.getTemplateData()
		templates.Index(templateData).Render(r.Context(), w)
	})

	// Generate CV endpoint
	http.HandleFunc("/generate/", func(w http.ResponseWriter, r *http.Request) {
		templateKey := strings.TrimPrefix(r.URL.Path, "/generate/")

		if err := cv.Generate(templateKey); err != nil {
			http.Error(w, fmt.Sprintf("Error generating CV: %v", err), http.StatusInternalServerError)
			return
		}

		// Redirect to the generated PDF
		http.Redirect(w, r, fmt.Sprintf("/static/output/cv-%s.pdf", templateKey), http.StatusSeeOther)
	})
}

func main() {
	templateFlag := flag.String("template", "vantage", "Template to use (vantage, basic, modern)")
	listFlag := flag.Bool("list", false, "List available templates")
	serveFlag := flag.Bool("serve", false, "Start web server")
	portFlag := flag.String("port", "8080", "Port to serve on")
	flag.Parse()

	generator := NewCVGenerator()

	if *serveFlag {
		generator.setupRoutes()
		fmt.Printf("Starting server on http://localhost:%s\n", *portFlag)
		log.Fatal(http.ListenAndServe(":"+*portFlag, nil))
		return
	}

	if *listFlag {
		generator.ListTemplates()
		return
	}

	if err := generator.Generate(*templateFlag); err != nil {
		log.Fatalf("Error generating CV: %v", err)
	}
}
