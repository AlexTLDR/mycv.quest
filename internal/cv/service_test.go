package cv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCVServiceIntegration(t *testing.T) {
	// Set up test directories
	templatesDir := filepath.Join("..", "..", "assets", "templates", "typst")
	outputDir := filepath.Join("..", "..", "tmp", "test_output")

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create test output directory: %v", err)
	}

	// Clean up after test
	defer func() {
		os.RemoveAll(outputDir)
	}()

	// Initialize CV service
	service, err := NewService(templatesDir, outputDir)
	if err != nil {
		t.Fatalf("Failed to initialize CV service: %v", err)
	}

	t.Run("ListTemplates", func(t *testing.T) {
		templates, err := service.ListTemplates()
		if err != nil {
			t.Fatalf("Failed to list templates: %v", err)
		}

		if len(templates) == 0 {
			t.Fatal("No templates found")
		}

		// Check that we have expected templates
		templateIDs := make(map[string]bool)
		for _, template := range templates {
			templateIDs[template.ID] = true
		}

		expectedTemplates := []string{"vantage", "grotesk"}
		for _, expectedID := range expectedTemplates {
			if !templateIDs[expectedID] {
				t.Errorf("Expected template %s not found", expectedID)
			}
		}
	})

	// Create comprehensive test data
	testData := TemplateData{
		TemplateID: "test",
		Data: map[string]interface{}{
			"contacts": map[string]interface{}{
				"name":     "Jane Smith",
				"title":    "Senior Software Engineer",
				"email":    "jane.smith@example.com",
				"address":  "New York, NY",
				"location": "United States",
				"linkedin": map[string]interface{}{
					"url":         "https://linkedin.com/in/janesmith",
					"displayText": "janesmith",
				},
				"github": map[string]interface{}{
					"url":         "https://github.com/janesmith",
					"displayText": "@janesmith",
				},
				"website": map[string]interface{}{
					"url":         "https://janesmith.dev",
					"displayText": "janesmith.dev",
				},
			},
			"position": "Senior Software Engineer",
			"tagline":  "Experienced software engineer with 8+ years developing scalable web applications and leading engineering teams. Passionate about clean code, system architecture, and mentoring junior developers.",
			"jobs": []interface{}{
				map[string]interface{}{
					"position": "Senior Software Engineer",
					"company": map[string]interface{}{
						"name": "TechCorp Inc.",
						"link": "https://techcorp.com",
					},
					"product": map[string]interface{}{
						"name": "CloudScale Platform",
						"link": "https://cloudscale.techcorp.com",
					},
					"description": []interface{}{
						"Led development of microservices architecture serving 1M+ users",
						"Reduced system latency by 40% through performance optimization",
						"Mentored team of 5 junior developers",
					},
					"from":     "2021 Jan.",
					"to":       "present",
					"tags":     []interface{}{"Go", "Kubernetes", "PostgreSQL", "React"},
					"location": "New York, NY",
				},
				map[string]interface{}{
					"position": "Software Engineer",
					"company": map[string]interface{}{
						"name": "StartupCo",
						"link": "https://startupco.com",
					},
					"description": []interface{}{
						"Built REST APIs using Node.js and Express",
						"Implemented CI/CD pipelines with GitHub Actions",
						"Collaborated with product team on feature development",
					},
					"from":     "2019 Jun.",
					"to":       "2020 Dec.",
					"tags":     []interface{}{"Node.js", "Docker", "AWS", "MongoDB"},
					"location": "Remote",
				},
			},
			"education": []interface{}{
				map[string]interface{}{
					"place": map[string]interface{}{
						"name": "Massachusetts Institute of Technology",
						"link": "https://mit.edu",
					},
					"degree":   "B.S.",
					"major":    "Computer Science",
					"track":    "Software Engineering",
					"from":     "2015",
					"to":       "2019",
					"location": "Cambridge, MA",
				},
			},
			"technical_expertise": []interface{}{
				map[string]interface{}{"name": "Go", "level": 5},
				map[string]interface{}{"name": "JavaScript", "level": 4},
				map[string]interface{}{"name": "Python", "level": 4},
				map[string]interface{}{"name": "Kubernetes", "level": 4},
				map[string]interface{}{"name": "PostgreSQL", "level": 3},
			},
			"skills": []interface{}{
				"Go", "JavaScript", "Python", "React", "Node.js",
				"Kubernetes", "Docker", "PostgreSQL", "MongoDB",
				"AWS", "GitHub Actions", "REST APIs",
			},
			"methodology": []interface{}{
				"Agile", "Scrum", "Test-Driven Development",
				"Code Review", "Pair Programming",
			},
			"tools": []interface{}{
				"VS Code", "Git", "Docker", "Kubernetes",
				"Postman", "Figma", "Jira", "Slack",
			},
			"achievements": []interface{}{
				map[string]interface{}{
					"name":        "AWS Certified Solutions Architect",
					"description": "Professional certification demonstrating expertise in designing distributed systems on AWS",
				},
				map[string]interface{}{
					"name":        "Best Innovation Award 2023",
					"description": "Recognized for developing automated deployment system that reduced deployment time by 60%",
				},
			},
		},
	}

	// Test each template
	templates := []string{"vantage", "grotesk"}
	for _, templateID := range templates {
		t.Run("Template_"+templateID, func(t *testing.T) {
			// Test template availability
			if !service.IsTemplateAvailable(templateID) {
				t.Fatalf("Template %s is not available", templateID)
			}

			// Test getting template info
			template, err := service.GetTemplate(templateID)
			if err != nil {
				t.Fatalf("Failed to get template %s: %v", templateID, err)
			}

			if template.ID != templateID {
				t.Errorf("Expected template ID %s, got %s", templateID, template.ID)
			}

			// Test data validation
			validation, err := service.ValidateData(templateID, testData)
			if err != nil {
				t.Fatalf("Failed to validate data for template %s: %v", templateID, err)
			}

			if !validation.Valid {
				t.Errorf("Data validation failed for template %s:", templateID)
				for _, valErr := range validation.Errors {
					t.Errorf("  - %s: %s", valErr.Field, valErr.Message)
				}
			}

			// Test form generation
			form, err := service.GenerateForm(templateID)
			if err != nil {
				t.Errorf("Failed to generate form for template %s: %v", templateID, err)
			} else {
				if form.TemplateID != templateID {
					t.Errorf("Expected form template ID %s, got %s", templateID, form.TemplateID)
				}
				if len(form.Sections) == 0 {
					t.Errorf("Form for template %s has no sections", templateID)
				}
			}

			// Test CV generation (only for vantage since grotesk has font issues in CI)
			if templateID == "vantage" {
				request := GenerationRequest{
					TemplateID: templateID,
					Data:       testData,
					Format:     "pdf",
				}

				result, err := service.GenerateCV(request)
				if err != nil {
					t.Fatalf("Failed to generate CV with template %s: %v", templateID, err)
				}

				if !result.Success {
					t.Fatalf("CV generation failed for template %s: %s", templateID, result.Message)
				}

				if len(result.Data) == 0 {
					t.Errorf("Generated CV data is empty for template %s", templateID)
				}

				if result.Filename == "" {
					t.Errorf("Generated CV filename is empty for template %s", templateID)
				}

				// Verify PDF starts with PDF header
				if len(result.Data) < 4 || string(result.Data[:4]) != "%PDF" {
					t.Errorf("Generated file doesn't appear to be a valid PDF for template %s", templateID)
				}

				// Test saving the file
				outputPath := filepath.Join(outputDir, result.Filename)
				if err := os.WriteFile(outputPath, result.Data, 0644); err != nil {
					t.Errorf("Failed to save generated CV: %v", err)
				}

				// Verify file was created and has content
				if stat, err := os.Stat(outputPath); err != nil {
					t.Errorf("Generated CV file not found: %v", err)
				} else if stat.Size() == 0 {
					t.Errorf("Generated CV file is empty")
				}
			}
		})
	}

	t.Run("TemplateMetadata", func(t *testing.T) {
		templateID := "vantage"
		metadata, err := service.GetTemplateMetadata(templateID)
		if err != nil {
			t.Fatalf("Failed to get template metadata: %v", err)
		}

		if metadata.TemplateID != templateID {
			t.Errorf("Expected metadata template ID %s, got %s", templateID, metadata.TemplateID)
		}
	})

	t.Run("SampleDataGeneration", func(t *testing.T) {
		templateID := "vantage"
		sampleData, err := service.GetSampleData(templateID)
		if err != nil {
			t.Fatalf("Failed to generate sample data: %v", err)
		}

		if sampleData.TemplateID != templateID {
			t.Errorf("Expected sample data template ID %s, got %s", templateID, sampleData.TemplateID)
		}

		if len(sampleData.Data) == 0 {
			t.Errorf("Sample data is empty")
		}
	})

	t.Run("RequiredFields", func(t *testing.T) {
		templateID := "vantage"
		requiredFields, err := service.GetRequiredFields(templateID)
		if err != nil {
			t.Fatalf("Failed to get required fields: %v", err)
		}

		// Test that the method works - required fields may or may not exist depending on template config
		t.Logf("Required fields for %s: %v", templateID, requiredFields)

		// If there are required fields, they should be valid field names
		for _, field := range requiredFields {
			if field == "" {
				t.Error("Found empty required field name")
			}
		}

		// Test that the method returns consistent results
		requiredFields2, err := service.GetRequiredFields(templateID)
		if err != nil {
			t.Fatalf("Failed to get required fields on second call: %v", err)
		}

		if len(requiredFields) != len(requiredFields2) {
			t.Error("Required fields method returned inconsistent results")
		}
	})

	t.Run("DataExtraction", func(t *testing.T) {
		templateID := "vantage"
		displayName, err := service.ExtractDisplayName(templateID, testData)
		if err != nil {
			t.Fatalf("Failed to extract display name: %v", err)
		}

		expectedName := "Jane Smith"
		if displayName != expectedName {
			t.Errorf("Expected display name %s, got %s", expectedName, displayName)
		}
	})
}

func TestTemplateAdapter(t *testing.T) {
	registry := NewAdapterRegistry()

	// Test vantage adapter
	vantageAdapter, exists := registry.GetAdapter("vantage")
	if !exists {
		t.Fatal("Vantage adapter not found")
	}

	if vantageAdapter.GetTemplateID() != "vantage" {
		t.Errorf("Expected vantage adapter template ID 'vantage', got %s", vantageAdapter.GetTemplateID())
	}

	// Test grotesk adapter
	groteskAdapter, exists := registry.GetAdapter("grotesk")
	if !exists {
		t.Fatal("Grotesk adapter not found")
	}

	if groteskAdapter.GetTemplateID() != "grotesk" {
		t.Errorf("Expected grotesk adapter template ID 'grotesk', got %s", groteskAdapter.GetTemplateID())
	}

	// Test data conversion
	testData := TemplateData{
		TemplateID: "vantage",
		Data: map[string]interface{}{
			"contacts": map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
			},
			"position": "Developer",
		},
	}

	// Convert from vantage to grotesk (should be pass-through now)
	convertedData, err := registry.ConvertData("vantage", "grotesk", testData)
	if err != nil {
		t.Fatalf("Failed to convert data: %v", err)
	}

	if convertedData.TemplateID != "grotesk" {
		t.Errorf("Expected converted template ID 'grotesk', got %s", convertedData.TemplateID)
	}

	// Since both use same structure now, data should be identical
	if len(convertedData.Data) == 0 {
		t.Error("Converted data is empty")
	}
}

func TestServiceErrors(t *testing.T) {
	// Test with invalid templates directory (use a file as directory path)
	tempFile := filepath.Join(os.TempDir(), "not_a_directory.txt")
	os.WriteFile(tempFile, []byte("test"), 0644)
	defer os.Remove(tempFile)

	service, err := NewService(tempFile, os.TempDir())
	if err == nil {
		t.Error("Expected error when initializing service with invalid templates directory")
	}

	// Test with valid service but invalid template
	templatesDir := filepath.Join("..", "..", "assets", "templates", "typst")
	outputDir := filepath.Join("..", "..", "tmp", "test_output")

	service, err = NewService(templatesDir, outputDir)
	if err != nil {
		t.Fatalf("Failed to initialize CV service: %v", err)
	}

	// Test invalid template ID
	_, err = service.GetTemplate("nonexistent_template")
	if err == nil {
		t.Error("Expected error when getting nonexistent template")
	}

	// Test validation with invalid template
	testData := TemplateData{
		TemplateID: "nonexistent",
		Data:       map[string]interface{}{},
	}

	_, err = service.ValidateData("nonexistent_template", testData)
	if err == nil {
		t.Error("Expected error when validating data for nonexistent template")
	}
}
