package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type CVGenerator struct {
	TemplateDir string
	OutputDir   string
}

func NewCVGenerator() *CVGenerator {
	return &CVGenerator{
		TemplateDir: "templates/vantage",
		OutputDir:   "output",
	}
}

func (cv *CVGenerator) Generate() error {
	if err := cv.ensureOutputDir(); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	cmd := exec.Command("typst", "compile", "example.typ", filepath.Join("..", "..", cv.OutputDir, "cv.pdf"))
	cmd.Dir = cv.TemplateDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("typst compilation failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("CV generated successfully at %s/cv.pdf\n", cv.OutputDir)
	return nil
}

func (cv *CVGenerator) ensureOutputDir() error {
	return os.MkdirAll(cv.OutputDir, 0755)
}

func main() {
	generator := NewCVGenerator()

	if err := generator.Generate(); err != nil {
		log.Fatalf("Error generating CV: %v", err)
	}
}
