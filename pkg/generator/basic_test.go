package generator_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/AlexTLDR/mycv.quest/pkg/config"
	"github.com/AlexTLDR/mycv.quest/pkg/generator"
)

func TestGenerateBasicCV(t *testing.T) {
	t.Parallel()
	// Create test configuration
	cfg := &config.Config{
		Templates: map[string]config.Template{
			"basic": {
				Name:      "Basic Resume",
				Dir:       "../../templates/basic/template",
				InputFile: "main.typ",
			},
		},
		OutputDir: "test_output",
	}

	gen := generator.New(cfg)

	// Create test form data
	formData := url.Values{
		"name":                      {"John Doe"},
		"location":                  {"New York, NY"},
		"email":                     {"john.doe@example.com"},
		"github":                    {"johndoe"},
		"linkedin":                  {"johndoe"},
		"phone":                     {"+1-555-123-4567"},
		"personal_site":             {"https://johndoe.dev"},
		"accent_color":              {"#26428b"},
		"education[0][institution]": {"University of Technology"},
		"education[0][location]":    {"Boston, MA"},
		"education[0][start_date]":  {"2018"},
		"education[0][end_date]":    {"2022"},
		"education[0][degree]":      {"Bachelor of Science in Computer Science"},
		"education[0][details]":     {"Graduated Magna Cum Laude\nRelevant Coursework: Data Structures, Algorithms"},
		"work[0][title]":            {"Software Engineer"},
		"work[0][company]":          {"Tech Corp"},
		"work[0][location]":         {"San Francisco, CA"},
		"work[0][start_date]":       {"2022"},
		"work[0][end_date]":         {"Present"},
		"work[0][description]":      {"Developed web applications using Go and React\nImplemented REST APIs"},
		"projects[0][name]":         {"Portfolio Website"},
		"projects[0][role]":         {"Full Stack Developer"},
		"projects[0][start_date]":   {"2021"},
		"projects[0][end_date]":     {"2022"},
		"projects[0][url]":          {"https://github.com/johndoe/portfolio"},
		"projects[0][description]":  {"Built responsive portfolio website\nUsed modern web technologies"},
		"programming_languages":     {"Go, JavaScript, Python"},
		"technologies":              {"React, Node.js, Docker"},
	}

	// Create HTTP request with form data
	req := &http.Request{
		Method: http.MethodPost,
		Header: make(http.Header),
		Form:   formData,
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Generate CV
	pdfData, err := gen.GenerateBasicCV(cfg.Templates["basic"], req)
	if err != nil {
		t.Fatalf("Failed to generate basic CV: %v", err)
	}

	// Verify PDF data is not empty
	if len(pdfData) == 0 {
		t.Fatal("Generated PDF data is empty")
	}

	// Verify PDF header (PDF files start with %PDF-)
	if !strings.HasPrefix(string(pdfData[:4]), "%PDF") {
		t.Fatal("Generated data is not a valid PDF file")
	}
}

func TestGenerateBasicTypContent(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}
	gen := generator.New(cfg)

	// Create test form data
	formData := url.Values{
		"name":                      {"Jane Smith"},
		"location":                  {"Austin, TX"},
		"email":                     {"jane.smith@example.com"},
		"github":                    {"janesmith"},
		"linkedin":                  {"janesmith"},
		"phone":                     {"+1-555-987-6543"},
		"personal_site":             {"https://janesmith.dev"},
		"accent_color":              {"#ff6b6b"},
		"education[0][institution]": {"State University"},
		"education[0][location]":    {"Austin, TX"},
		"education[0][start_date]":  {"2019"},
		"education[0][end_date]":    {"2023"},
		"education[0][degree]":      {"Master of Science in Software Engineering"},
		"work[0][title]":            {"Senior Developer"},
		"work[0][company]":          {"Innovation Labs"},
		"work[0][location]":         {"Remote"},
		"work[0][start_date]":       {"2023"},
		"work[0][end_date]":         {"Present"},
		"work[0][description]":      {"Lead development of microservices architecture"},
		"programming_languages":     {"Rust, TypeScript, Go"},
		"technologies":              {"Kubernetes, PostgreSQL, Redis"},
	}

	req := &http.Request{
		Method: http.MethodPost,
		Header: make(http.Header),
		Form:   formData,
	}

	content := gen.GenerateBasicTypContent(req)

	// Test that content includes expected sections
	expectedSections := []string{
		"#import \"@preview/basic-resume:0.2.8\": *",
		"== Education",
		"== Work Experience",
		"== Skills",
		"Jane Smith",
		"jane.smith@example.com",
		"State University",
		"Senior Developer",
		"Innovation Labs",
		"Rust, TypeScript, Go",
		"Kubernetes, PostgreSQL, Redis",
	}

	for _, expected := range expectedSections {
		if !strings.Contains(content, expected) {
			t.Errorf("Generated content missing expected section: %s", expected)
		}
	}

	// Test accent color
	if !strings.Contains(content, "#ff6b6b") {
		t.Error("Generated content missing custom accent color")
	}
}

func TestGenerateBasicTypContentEmpty(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}
	gen := generator.New(cfg)

	// Create empty request
	req := &http.Request{
		Method: http.MethodPost,
		Header: make(http.Header),
		Form:   url.Values{},
	}

	content := gen.GenerateBasicTypContent(req)

	// Should still generate valid typst content with defaults
	expectedDefaults := []string{
		"#import \"@preview/basic-resume:0.2.8\": *",
		"== Education",
		"== Work Experience",
		"#26428b", // default accent color
	}

	for _, expected := range expectedDefaults {
		if !strings.Contains(content, expected) {
			t.Errorf("Generated content missing expected default: %s", expected)
		}
	}
}

func TestGenerateBasicTypContentSanitization(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}
	gen := generator.New(cfg)

	// Create form data with Typst-problematic content
	formData := url.Values{
		"name":     {"John \"The Great\" Doe"},
		"email":    {"john@example.com#test"},
		"github":   {"johndoe$repo"},
		"location": {"New York & Associates"},
	}

	req := &http.Request{
		Method: http.MethodPost,
		Header: make(http.Header),
		Form:   formData,
	}

	content := gen.GenerateBasicTypContent(req)

	// Check that Typst special characters are properly escaped
	expectedEscaped := []string{
		"\\\"", // Quotes should be escaped
		"\\#",  // Hash should be escaped (Typst function marker)
		"\\$",  // Dollar should be escaped (Typst math mode)
	}

	for _, expected := range expectedEscaped {
		if !strings.Contains(content, expected) {
			t.Errorf("Generated content missing expected escaped pattern: %s", expected)
		}
	}

	// Should still contain safe parts
	if !strings.Contains(content, "John") || !strings.Contains(content, "Doe") {
		t.Error("Sanitization removed legitimate content")
	}
}
