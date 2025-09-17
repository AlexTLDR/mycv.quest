package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type Template struct {
	Name      string
	Dir       string
	InputFile string
}

type CVGenerator struct {
	Templates map[string]Template
	OutputDir string
}

func NewCVGenerator() *CVGenerator {
	templates := map[string]Template{
		"vantage": {
			Name:      "Vantage",
			Dir:       "templates/vantage",
			InputFile: "example.typ",
		},
		"basic": {
			Name:      "Basic Resume",
			Dir:       "templates/basic/template",
			InputFile: "main.typ",
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

func main() {
	var templateFlag = flag.String("template", "vantage", "Template to use (vantage, basic)")
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
