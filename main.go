package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	return os.MkdirAll(cv.OutputDir, 0755)
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
	destPhoto := filepath.Join(templateDir, "avatar.png")

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
	var templateFlag = flag.String("template", "vantage", "Template to use (vantage, basic, modern)")
	var listFlag = flag.Bool("list", false, "List available templates")
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
