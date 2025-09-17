package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
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

func main() {
	templateFlag := flag.String("template", "vantage", "Template to use (vantage, basic, modern)")
	listFlag := flag.Bool("list", false, "List available templates")
	flag.Parse()

	generator := NewCVGenerator()

	if *listFlag {
		generator.ListTemplates()
		return
	}

	if err := generator.Generate(*templateFlag); err != nil {
		log.Fatalf("Error generating CV: %v", err)
	}
}
