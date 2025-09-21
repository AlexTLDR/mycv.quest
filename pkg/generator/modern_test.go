package generator

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/AlexTLDR/mycv.quest/pkg/config"
)

func TestGenerateModernCV(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Templates: map[string]config.Template{
			"modern": {
				Name:       "Modern Resume",
				Dir:        "../../templates/modern/template",
				InputFile:  "main.typ",
				NeedsPhoto: true,
			},
		},
		OutputDir: "test_output",
	}

	generator := New(cfg)

	// Create test form data
	formData := url.Values{
		"author":                         {"Alice Johnson"},
		"job_title":                      {"Senior Software Engineer"},
		"bio":                            {"Passionate developer with 5+ years of experience"},
		"email":                          {"alice.johnson@example.com"},
		"mobile":                         {"+1-555-444-3333"},
		"location":                       {"Seattle, WA"},
		"linkedin":                       {"alicejohnson"},
		"github":                         {"alicejohnson"},
		"website":                        {"https://alicejohnson.dev"},
		"education[0][title]":            {"Master of Computer Science"},
		"education[0][subtitle]":         {"University of Washington"},
		"education[0][date_from]":        {"2018"},
		"education[0][date_to]":          {"2020"},
		"education[0][task_description]": {"Specialized in Machine Learning\nThesis on Neural Networks"},
		"work[0][title]":                 {"Senior Software Engineer"},
		"work[0][subtitle]":              {"Tech Innovations Inc."},
		"work[0][facility_description]":  {"Leading technology company"},
		"work[0][date_from]":             {"2020"},
		"work[0][date_to]":               {"Present"},
		"work[0][task_description]":      {"Lead development of cloud platforms\nManage team of 5 engineers"},
		"skills":                         {"Go, Rust, Kubernetes, AWS, Machine Learning"},
		"projects[0][title]":             {"Cloud Migration Platform"},
		"projects[0][subtitle]":          {"Internal Tool"},
		"projects[0][date_from]":         {"2023"},
		"projects[0][date_to]":           {"2024"},
		"projects[0][description]":       {"Automated cloud migration for legacy systems"},
		"certificates[0][title]":         {"AWS Solutions Architect"},
		"certificates[0][subtitle]":      {"Amazon Web Services"},
		"certificates[0][date_from]":     {"2022"},
		"certificates[0][date_to]":       {"2025"},
		"languages":                      {"English, Spanish, French"},
		"interests":                      {"Rock climbing, Photography, Open source"},
	}

	// Create HTTP request with form data
	req := &http.Request{
		Method: "POST",
		Header: make(http.Header),
		Form:   formData,
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Generate CV
	pdfData, err := generator.generateModernCV(cfg.Templates["modern"], req)
	if err != nil {
		t.Fatalf("Failed to generate modern CV: %v", err)
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

func TestGenerateModernTypContent(t *testing.T) {
	cfg := &config.Config{}
	generator := New(cfg)

	// Create test form data
	formData := url.Values{
		"author":                         {"Bob Wilson"},
		"job_title":                      {"DevOps Engineer"},
		"bio":                            {"Infrastructure automation specialist"},
		"email":                          {"bob.wilson@example.com"},
		"mobile":                         {"+1-555-111-2222"},
		"location":                       {"Denver, CO"},
		"linkedin":                       {"bobwilson"},
		"github":                         {"https://github.com/bobwilson"},
		"website":                        {"bobwilson.tech"},
		"education[0][title]":            {"Bachelor of Engineering"},
		"education[0][subtitle]":         {"Colorado State University"},
		"education[0][date_from]":        {"2016"},
		"education[0][date_to]":          {"2020"},
		"education[0][task_description]": {"Computer Engineering major\nMinor in Mathematics"},
		"work[0][title]":                 {"DevOps Engineer"},
		"work[0][subtitle]":              {"CloudTech Solutions"},
		"work[0][facility_description]":  {"Cloud infrastructure provider"},
		"work[0][date_from]":             {"2020"},
		"work[0][date_to]":               {"Present"},
		"work[0][task_description]":      {"Manage CI/CD pipelines\nAutomate infrastructure deployment"},
		"skills":                         {"Docker, Kubernetes, Terraform, Python"},
		"projects[0][title]":             {"Infrastructure as Code"},
		"projects[0][subtitle]":          {"Open Source"},
		"projects[0][date_from]":         {"2022"},
		"projects[0][date_to]":           {"2023"},
		"projects[0][description]":       {"Terraform modules for AWS infrastructure"},
		"languages":                      {"English, German"},
		"interests":                      {"Hiking, Music production"},
	}

	req := &http.Request{
		Method: "POST",
		Header: make(http.Header),
		Form:   formData,
	}

	avatarFilename := "avatar.png"
	content := generator.generateModernTypContent(req, avatarFilename)

	// Test that content includes expected sections and data
	expectedSections := []string{
		"#import \"@preview/modern-resume:0.1.0\": modern-resume",
		"== Education",
		"== Work experience",
		"== Skills",
		"== Projects",
		"== Certificates",
		"== Languages",
		"== Interests",
		"Bob Wilson",
		"DevOps Engineer",
		"bob.wilson@example.com",
		"Colorado State University",
		"CloudTech Solutions",
		"Infrastructure as Code",
		"image(\"avatar.png\")",
	}

	for _, expected := range expectedSections {
		if !strings.Contains(content, expected) {
			t.Errorf("Generated content missing expected section: %s", expected)
		}
	}

	// Test GitHub URL normalization
	if !strings.Contains(content, "https://github.com/bobwilson") {
		t.Error("GitHub URL not properly normalized in content")
	}

	// Test website URL normalization
	if !strings.Contains(content, "https://bobwilson.tech") {
		t.Error("Website URL not properly normalized in content")
	}
}

func TestGenerateModernTypContentMinimal(t *testing.T) {
	cfg := &config.Config{}
	generator := New(cfg)

	// Create minimal form data - only required fields
	formData := url.Values{
		"author":    {"Minimal User"},
		"job_title": {"Developer"},
		"bio":       {"Simple bio"},
		"email":     {"minimal@example.com"},
	}

	req := &http.Request{
		Method: "POST",
		Header: make(http.Header),
		Form:   formData,
	}

	avatarFilename := "default.png"
	content := generator.generateModernTypContent(req, avatarFilename)

	// Should still generate valid typst content with minimal data
	expectedElements := []string{
		"#import \"@preview/modern-resume:0.1.0\": modern-resume",
		"Minimal User",
		"Developer",
		"Simple bio",
		"minimal@example.com",
		"image(\"default.png\")",
		"== Education",
		"== Work experience",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(content, expected) {
			t.Errorf("Generated content missing expected element: %s", expected)
		}
	}
}

func TestGenerateModernTypContentSanitization(t *testing.T) {
	cfg := &config.Config{}
	generator := New(cfg)

	// Create form data with Typst-problematic content
	formData := url.Values{
		"author":    {"Alice \"The Engineer\" Doe"},
		"job_title": {"Senior Engineer#Lead"},
		"bio":       {"Bio with $math and #functions"},
		"email":     {"alice@example.com#test"},
		"github":    {"alice$github"},
		"website":   {"alice.dev"},
	}

	req := &http.Request{
		Method: "POST",
		Header: make(http.Header),
		Form:   formData,
	}

	avatarFilename := "avatar.png"
	content := generator.generateModernTypContent(req, avatarFilename)

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
	if !strings.Contains(content, "Alice") || !strings.Contains(content, "Doe") {
		t.Error("Sanitization removed legitimate content")
	}
}

func TestGenerateModernTypContentSkillsParsing(t *testing.T) {
	cfg := &config.Config{}
	generator := New(cfg)

	// Test skills parsing with various formats
	formData := url.Values{
		"author":    {"Test User"},
		"job_title": {"Developer"},
		"bio":       {"Test bio"},
		"email":     {"test@example.com"},
		"skills":    {"Go, React,   Node.js  , Python,Docker"},
		"languages": {"English,Spanish, French"},
		"interests": {"Coding,   Gaming,  Travel"},
	}

	req := &http.Request{
		Method: "POST",
		Header: make(http.Header),
		Form:   formData,
	}

	avatarFilename := "avatar.png"
	content := generator.generateModernTypContent(req, avatarFilename)

	// Check that skills are properly parsed and formatted as pills
	expectedSkills := []string{
		"Go",
		"React",
		"Node.js",
		"Python",
		"Docker",
	}

	for _, skill := range expectedSkills {
		skillPattern := "#pill(\"" + skill + "\", fill: true)"
		if !strings.Contains(content, skillPattern) {
			t.Errorf("Generated content missing expected skill pill for: %s", skill)
		}
	}

	// Check languages and interests
	expectedLanguages := []string{
		"English",
		"Spanish",
		"French",
	}

	for _, lang := range expectedLanguages {
		langPattern := "#pill(\"" + lang + "\")"
		if !strings.Contains(content, langPattern) {
			t.Errorf("Generated content missing expected language pill for: %s", lang)
		}
	}
}
