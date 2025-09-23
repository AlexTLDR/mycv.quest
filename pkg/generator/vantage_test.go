package generator_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/AlexTLDR/mycv.quest/pkg/config"
	"github.com/AlexTLDR/mycv.quest/pkg/generator"
	"gopkg.in/yaml.v2"
)

func TestGenerateVantageCV(t *testing.T) {
	t.Parallel()
	// Create test configuration
	cfg := &config.Config{
		Templates: map[string]config.Template{
			"vantage": {
				Name:      "Vantage Resume",
				Dir:       "../../templates/vantage",
				InputFile: "example.typ",
			},
		},
		OutputDir: "test_output",
	}

	gen := generator.New(cfg)

	// Create test form data
	formData := url.Values{
		"name":                          {"Charlie Brown"},
		"title":                         {"Full Stack Developer"},
		"email":                         {"charlie.brown@example.com"},
		"address":                       {"123 Main St"},
		"location":                      {"Portland, OR"},
		"linkedin_url":                  {"https://linkedin.com/in/charliebrown"},
		"linkedin_display_text":         {"linkedin.com/in/charliebrown"},
		"github_url":                    {"https://github.com/charliebrown"},
		"github_display_text":           {"github.com/charliebrown"},
		"website_url":                   {"https://charliebrown.dev"},
		"website_display_text":          {"charliebrown.dev"},
		"position":                      {"Senior Developer"},
		"tagline":                       {"Building scalable web applications"},
		"objective":                     {"Seeking challenging opportunities in full-stack development"},
		"jobs[0][position]":             {"Software Engineer"},
		"jobs[0][company_name]":         {"Acme Corp"},
		"jobs[0][company_link]":         {"https://acme.com"},
		"jobs[0][product_name]":         {"E-commerce Platform"},
		"jobs[0][product_link]":         {"https://acme.com/platform"},
		"jobs[0][from]":                 {"2022"},
		"jobs[0][to]":                   {"Present"},
		"jobs[0][location]":             {"Remote"},
		"jobs[0][description]":          {"Developed microservices architecture\nImplemented REST APIs\nOptimized database performance"},
		"jobs[0][tags]":                 {"Go, PostgreSQL, Docker, Kubernetes"},
		"education[0][place_name]":      {"Oregon State University"},
		"education[0][place_link]":      {"https://oregonstate.edu"},
		"education[0][degree]":          {"Bachelor of Science"},
		"education[0][major]":           {"Computer Science"},
		"education[0][track]":           {"Software Engineering"},
		"education[0][from]":            {"2018"},
		"education[0][to]":              {"2022"},
		"education[0][location]":        {"Corvallis, OR"},
		"technical_expertise[0][name]":  {"Go"},
		"technical_expertise[0][level]": {"5"},
		"technical_expertise[1][name]":  {"React"},
		"technical_expertise[1][level]": {"4"},
		"achievements[0][name]":         {"Employee of the Year"},
		"achievements[0][description]":  {"Recognized for outstanding contribution to platform development"},
		"skills":                        {"Go, React, PostgreSQL, Docker"},
		"methodology":                   {"Agile, Scrum, TDD"},
		"tools":                         {"Git, Docker, Kubernetes, AWS"},
	}

	// Create HTTP request with form data
	req := &http.Request{
		Method: http.MethodPost,
		Header: make(http.Header),
		Form:   formData,
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Generate CV
	pdfData, err := gen.GenerateVantageCV(cfg.Templates["vantage"], req)
	if err != nil {
		t.Fatalf("Failed to generate vantage CV: %v", err)
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

func TestGenerateVantageYAMLContent(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}
	gen := generator.New(cfg)

	// Create comprehensive test form data
	formData := url.Values{
		"name":                          {"Diana Prince"},
		"title":                         {"Senior Software Architect"},
		"email":                         {"diana.prince@example.com"},
		"address":                       {"456 Hero Lane"},
		"location":                      {"Metropolis, NY"},
		"linkedin_url":                  {"https://linkedin.com/in/dianaprince"},
		"linkedin_display_text":         {"linkedin.com/in/dianaprince"},
		"github_url":                    {"github.com/dianaprince"},
		"github_display_text":           {"github.com/dianaprince"},
		"website_url":                   {"dianaprince.dev"},
		"website_display_text":          {"dianaprince.dev"},
		"position":                      {"Software Architect"},
		"tagline":                       {"Designing robust software systems"},
		"objective":                     {"Lead architectural decisions for enterprise applications"},
		"jobs[0][position]":             {"Senior Software Engineer"},
		"jobs[0][company_name]":         {"TechCorp"},
		"jobs[0][company_link]":         {"https://techcorp.com"},
		"jobs[0][product_name]":         {"Cloud Platform"},
		"jobs[0][product_link]":         {"https://techcorp.com/platform"},
		"jobs[0][from]":                 {"2021"},
		"jobs[0][to]":                   {"Present"},
		"jobs[0][location]":             {"New York, NY"},
		"jobs[0][description]":          {"Lead architecture design\n- Implemented microservices\n- Mentored junior developers"},
		"jobs[0][tags]":                 {"Go, Kubernetes, AWS, PostgreSQL"},
		"jobs[1][position]":             {"Software Developer"},
		"jobs[1][company_name]":         {"StartupXYZ"},
		"jobs[1][company_link]":         {"https://startupxyz.com"},
		"jobs[1][product_name]":         {"Mobile App"},
		"jobs[1][product_link]":         {"https://startupxyz.com/app"},
		"jobs[1][from]":                 {"2019"},
		"jobs[1][to]":                   {"2021"},
		"jobs[1][location]":             {"San Francisco, CA"},
		"jobs[1][description]":          {"Full-stack development"},
		"jobs[1][tags]":                 {"React, Node.js, MongoDB"},
		"education[0][place_name]":      {"MIT"},
		"education[0][place_link]":      {"https://mit.edu"},
		"education[0][degree]":          {"Master of Science"},
		"education[0][major]":           {"Computer Science"},
		"education[0][track]":           {"Distributed Systems"},
		"education[0][from]":            {"2017"},
		"education[0][to]":              {"2019"},
		"education[0][location]":        {"Cambridge, MA"},
		"technical_expertise[0][name]":  {"Go"},
		"technical_expertise[0][level]": {"5"},
		"technical_expertise[1][name]":  {"Kubernetes"},
		"technical_expertise[1][level]": {"4"},
		"technical_expertise[2][name]":  {"React"},
		"technical_expertise[2][level]": {"3"},
		"achievements[0][name]":         {"Best Architecture Award"},
		"achievements[0][description]":  {"Awarded for innovative microservices design"},
		"achievements[1][name]":         {"Tech Conference Speaker"},
		"achievements[1][description]":  {"Presented at major technology conferences"},
		"skills":                        {"Go, Rust, Python, TypeScript"},
		"methodology":                   {"Agile, DevOps, Domain-Driven Design"},
		"tools":                         {"Docker, Kubernetes, Terraform, GitLab CI"},
	}

	req := &http.Request{
		Method: http.MethodPost,
		Header: make(http.Header),
		Form:   formData,
	}

	yamlContent := gen.GenerateVantageYAMLContent(req)

	// Parse the generated YAML to verify structure
	var data map[string]interface{}
	err := yaml.Unmarshal(yamlContent, &data)
	if err != nil {
		t.Fatalf("Generated YAML is not valid: %v", err)
	}

	// Test contacts section
	contacts, ok := data["contacts"].(map[interface{}]interface{})
	if !ok {
		t.Fatal("Contacts section missing or invalid")
	}

	expectedContacts := map[string]string{
		"name":     "Diana Prince",
		"title":    "Senior Software Architect",
		"email":    "diana.prince@example.com",
		"address":  "456 Hero Lane",
		"location": "Metropolis, NY",
	}

	for key, expected := range expectedContacts {
		if actual, exists := contacts[key]; !exists || actual != expected {
			t.Errorf("Contact %s: expected %s, got %v", key, expected, actual)
		}
	}

	// Test nested contact links
	linkedin, ok := contacts["linkedin"].(map[interface{}]interface{})
	if !ok {
		t.Fatal("LinkedIn contact missing or invalid")
	}
	if linkedin["url"] != "https://linkedin.com/in/dianaprince" {
		t.Errorf("LinkedIn URL incorrect: %v", linkedin["url"])
	}

	// Test jobs section
	jobs, ok := data["jobs"].([]interface{})
	if !ok {
		t.Fatal("Jobs section missing or invalid")
	}

	if len(jobs) != 2 {
		t.Fatalf("Expected 2 jobs, got %d", len(jobs))
	}

	// Test first job
	job1, ok := jobs[0].(map[interface{}]interface{})
	if !ok {
		t.Fatal("First job invalid")
	}

	if job1["position"] != "Senior Software Engineer" {
		t.Errorf("Job position incorrect: %v", job1["position"])
	}

	// Test job description parsing
	description, ok := job1["description"].([]interface{})
	if !ok {
		t.Fatal("Job description should be a list")
	}

	expectedDescItems := []string{
		"Lead architecture design",
		"Implemented microservices",
		"Mentored junior developers",
	}

	if len(description) != len(expectedDescItems) {
		t.Fatalf("Expected %d description items, got %d", len(expectedDescItems), len(description))
	}

	// Test technical expertise section
	techExpertise, ok := data["technical_expertise"].([]interface{})
	if !ok {
		t.Fatal("Technical expertise section missing or invalid")
	}

	if len(techExpertise) != 3 {
		t.Fatalf("Expected 3 technical expertise items, got %d", len(techExpertise))
	}

	// Test expertise levels
	expertise1, ok := techExpertise[0].(map[interface{}]interface{})
	if !ok {
		t.Fatal("First technical expertise item invalid")
	}

	if expertise1["name"] != "Go" || expertise1["level"] != 5 {
		t.Errorf("Technical expertise incorrect: name=%v, level=%v", expertise1["name"], expertise1["level"])
	}

	// Test achievements section
	achievements, ok := data["achievements"].([]interface{})
	if !ok {
		t.Fatal("Achievements section missing or invalid")
	}

	if len(achievements) != 2 {
		t.Fatalf("Expected 2 achievements, got %d", len(achievements))
	}

	// Test comma-separated lists
	skills, ok := data["skills"].([]interface{})
	if !ok {
		t.Fatal("Skills section missing or invalid")
	}

	expectedSkills := []string{"Go", "Rust", "Python", "TypeScript"}
	if len(skills) != len(expectedSkills) {
		t.Fatalf("Expected %d skills, got %d", len(expectedSkills), len(skills))
	}

	for i, expected := range expectedSkills {
		if skills[i] != expected {
			t.Errorf("Skill %d: expected %s, got %v", i, expected, skills[i])
		}
	}
}

func TestGenerateVantageYAMLContentMinimal(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}
	gen := generator.New(cfg)

	// Create minimal form data
	formData := url.Values{
		"name":     {"John Minimal"},
		"title":    {"Developer"},
		"email":    {"john@example.com"},
		"position": {"Software Developer"},
	}

	req := &http.Request{
		Method: http.MethodPost,
		Header: make(http.Header),
		Form:   formData,
	}

	yamlContent := gen.GenerateVantageYAMLContent(req)

	// Parse the generated YAML
	var data map[string]interface{}
	err := yaml.Unmarshal(yamlContent, &data)
	if err != nil {
		t.Fatalf("Generated YAML is not valid: %v", err)
	}

	// Should have empty arrays for sections with no data
	jobs, ok := data["jobs"].([]interface{})
	if !ok || jobs == nil {
		t.Fatal("Jobs should be empty array, not nil")
	}

	education, ok := data["education"].([]interface{})
	if !ok || education == nil {
		t.Fatal("Education should be empty array, not nil")
	}

	// Should still have basic contact info
	contacts, ok := data["contacts"].(map[interface{}]interface{})
	if !ok {
		t.Fatal("Contacts section missing")
	}

	if contacts["name"] != "John Minimal" {
		t.Errorf("Name incorrect: %v", contacts["name"])
	}
}

func TestGenerateVantageYAMLContentSanitization(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}
	gen := generator.New(cfg)

	// Create form data with Typst-problematic content
	formData := url.Values{
		"name":                  {"John \"The Developer\" Smith"},
		"title":                 {"Developer#Lead"},
		"email":                 {"john@example.com#test"},
		"jobs[0][position]":     {"Engineer$Senior"},
		"jobs[0][company_name]": {"Tech Corp & Associates"},
		"jobs[0][description]":  {"Did good things\nUsed $tech and #functions\nMore good things"},
		"skills":                {"Go, React, Node#js"},
		"website_url":           {"john.dev"},
	}

	req := &http.Request{
		Method: http.MethodPost,
		Header: make(http.Header),
		Form:   formData,
	}

	yamlContent := gen.GenerateVantageYAMLContent(req)

	// Parse the YAML - it should still be valid
	var data map[string]interface{}
	err := yaml.Unmarshal(yamlContent, &data)
	if err != nil {
		t.Fatalf("Generated YAML is not valid: %v", err)
	}

	// Convert back to string to check content
	yamlString := string(yamlContent)

	// Check that Typst special characters are properly escaped in YAML values
	expectedEscaped := []string{
		"\\\"", // Quotes should be escaped
		"\\#",  // Hash should be escaped (Typst function marker)
		"\\$",  // Dollar should be escaped (Typst math mode)
	}

	for _, expected := range expectedEscaped {
		if !strings.Contains(yamlString, expected) {
			t.Errorf("Generated YAML missing expected escaped pattern: %s", expected)
		}
	}

	// Check URL normalization
	if !strings.Contains(yamlString, "https://john.dev") {
		t.Error("URL should be normalized to https://")
	}

	// Should still contain legitimate content
	if !strings.Contains(yamlString, "John") || !strings.Contains(yamlString, "Smith") {
		t.Error("Sanitization removed legitimate content")
	}
}

func TestGenerateVantageYAMLContentTechnicalExpertiseLevels(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}
	gen := generator.New(cfg)

	// Test different level values including edge cases
	formData := url.Values{
		"name":                          {"Test User"},
		"email":                         {"test@example.com"},
		"technical_expertise[0][name]":  {"Go"},
		"technical_expertise[0][level]": {"5"},
		"technical_expertise[1][name]":  {"JavaScript"},
		"technical_expertise[1][level]": {"0"}, // Should default to 4
		"technical_expertise[2][name]":  {"Python"},
		"technical_expertise[2][level]": {"invalid"}, // Should default to 4
		"technical_expertise[3][name]":  {"Rust"},
		"technical_expertise[3][level]": {"3"},
	}

	req := &http.Request{
		Method: http.MethodPost,
		Header: make(http.Header),
		Form:   formData,
	}

	yamlContent := gen.GenerateVantageYAMLContent(req)

	var data map[string]interface{}
	err := yaml.Unmarshal(yamlContent, &data)
	if err != nil {
		t.Fatalf("Generated YAML is not valid: %v", err)
	}

	techExpertise, ok := data["technical_expertise"].([]interface{})
	if !ok {
		t.Fatal("Technical expertise section missing")
	}

	// Test level handling
	testCases := []struct {
		index         int
		expectedName  string
		expectedLevel int
	}{
		{0, "Go", 5},
		{1, "JavaScript", 4}, // 0 should default to 4
		{2, "Python", 4},     // invalid should default to 4
		{3, "Rust", 3},
	}

	for _, tc := range testCases {
		if tc.index >= len(techExpertise) {
			t.Fatalf("Expected at least %d technical expertise items", tc.index+1)
		}

		item, ok := techExpertise[tc.index].(map[interface{}]interface{})
		if !ok {
			t.Fatalf("Technical expertise item %d invalid", tc.index)
		}

		if item["name"] != tc.expectedName {
			t.Errorf("Item %d name: expected %s, got %v", tc.index, tc.expectedName, item["name"])
		}

		level, ok := item["level"].(int)
		if !ok {
			t.Fatalf("Item %d level should be int, got %T", tc.index, item["level"])
		}

		if level != tc.expectedLevel {
			t.Errorf("Item %d level: expected %d, got %d", tc.index, tc.expectedLevel, level)
		}
	}
}
