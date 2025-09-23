package config

import (
	"fmt"
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

type Config struct {
	Templates map[string]Template
	OutputDir string
}

func NewConfig() *Config {
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

	return &Config{
		Templates: templates,
		OutputDir: "output",
	}
}

func (c *Config) GetTemplate(key string) (Template, bool) {
	template, exists := c.Templates[key]
	return template, exists
}

func (c *Config) GetTemplateKeys() []string {
	keys := make([]string, 0, len(c.Templates))
	for key := range c.Templates {
		keys = append(keys, key)
	}
	return keys
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

func ValidateTemplateArgs(template Template, outputFile string) error {
	allowedInputFiles := []string{"example.typ", "main.typ"}
	if !slices.Contains(allowedInputFiles, template.InputFile) {
		return fmt.Errorf("invalid input file: %s", template.InputFile)
	}

	// Check for command injection attempts
	if strings.ContainsAny(template.InputFile, ";|&$`") || strings.ContainsAny(outputFile, ";|&$`") {
		return fmt.Errorf("invalid characters in file paths")
	}

	if !strings.HasSuffix(outputFile, ".pdf") {
		return fmt.Errorf("output file must have .pdf extension")
	}

	if err := validatePath(outputFile, ""); err != nil {
		return fmt.Errorf("invalid output file path: %w", err)
	}

	return nil
}

func ValidatePath(path, expectedPrefix string) error {
	return validatePath(path, expectedPrefix)
}
