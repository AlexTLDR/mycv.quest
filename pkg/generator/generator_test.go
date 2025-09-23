package generator_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexTLDR/mycv.quest/pkg/config"
	"github.com/AlexTLDR/mycv.quest/pkg/generator"
)

func TestNew(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{
		Templates: map[string]config.Template{
			"basic": {Name: "Basic Resume"},
		},
		OutputDir: "test_output",
	}

	gen := generator.New(cfg)

	if gen == nil {
		t.Fatal("New() returned nil")
	}

	// Test that the generator was created successfully
	// (Internal config is private, but we can test the public API works)
}

func TestListTemplates(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{
		Templates: map[string]config.Template{
			"basic":   {Name: "Basic Resume"},
			"modern":  {Name: "Modern Resume"},
			"vantage": {Name: "Vantage Resume"},
		},
	}

	gen := generator.New(cfg)

	// Capture stdout to test output
	// Note: This is a simple test - in a real scenario you might want to refactor
	// ListTemplates to return a string or write to an io.Writer for easier testing
	gen.ListTemplates()

	// Since ListTemplates prints to stdout, we can't easily test the output
	// without refactoring. This test mainly ensures it doesn't crash.
}

func TestGetTemplateData(t *testing.T) {
	t.Parallel()
	// Create temporary template directories for testing
	tempDir := t.TempDir()

	basicDir := filepath.Join(tempDir, "basic")
	modernDir := filepath.Join(tempDir, "modern")
	vantageDir := filepath.Join(tempDir, "vantage")

	err := os.MkdirAll(basicDir, 0o750)
	if err != nil {
		t.Fatalf("Failed to create basic dir: %v", err)
	}

	err = os.MkdirAll(modernDir, 0o750)
	if err != nil {
		t.Fatalf("Failed to create modern dir: %v", err)
	}

	err = os.MkdirAll(vantageDir, 0o750)
	if err != nil {
		t.Fatalf("Failed to create vantage dir: %v", err)
	}

	// Create thumbnail files
	err = os.WriteFile(filepath.Join(basicDir, "thumbnail.png"), []byte("fake png"), 0o600)
	if err != nil {
		t.Fatalf("Failed to create thumbnail: %v", err)
	}

	err = os.WriteFile(filepath.Join(vantageDir, "screenshot.png"), []byte("fake png"), 0o600)
	if err != nil {
		t.Fatalf("Failed to create screenshot: %v", err)
	}

	cfg := &config.Config{
		Templates: map[string]config.Template{
			"basic": {
				Name: "Basic Resume",
				Dir:  basicDir,
			},
			"modern": {
				Name: "Modern Resume",
				Dir:  modernDir,
			},
			"vantage": {
				Name: "Vantage Resume",
				Dir:  vantageDir,
			},
		},
	}

	gen := generator.New(cfg)
	templateData := gen.GetTemplateData()

	if len(templateData) != 3 {
		t.Fatalf("Expected 3 templates, got %d", len(templateData))
	}

	// Check that all templates are present
	templateKeys := make(map[string]bool)
	for _, tmpl := range templateData {
		templateKeys[tmpl.Key] = true

		// Test that basic fields are set
		if tmpl.Name == "" {
			t.Errorf("Template %s has empty name", tmpl.Key)
		}

		if tmpl.Description == "" {
			t.Errorf("Template %s has empty description", tmpl.Key)
		}

		if tmpl.PDFPath == "" {
			t.Errorf("Template %s has empty PDF path", tmpl.Key)
		}

		// Test specific thumbnail paths
		switch tmpl.Key {
		case "basic":
			expectedPath := "/static/" + basicDir + "/thumbnail.png"
			if tmpl.ThumbnailPath != expectedPath {
				t.Errorf("Basic template thumbnail path incorrect: expected %s, got %s", expectedPath, tmpl.ThumbnailPath)
			}
		case "vantage":
			expectedPath := "/static/" + vantageDir + "/screenshot.png"
			if tmpl.ThumbnailPath != expectedPath {
				t.Errorf("Vantage template thumbnail path incorrect: expected %s, got %s", expectedPath, tmpl.ThumbnailPath)
			}
		case "modern":
			// Modern should have empty thumbnail path since no file exists
			if tmpl.ThumbnailPath != "" {
				t.Errorf("Modern template should have empty thumbnail path, got %s", tmpl.ThumbnailPath)
			}
		}
	}

	expectedKeys := []string{"basic", "modern", "vantage"}
	for _, key := range expectedKeys {
		if !templateKeys[key] {
			t.Errorf("Template %s not found in results", key)
		}
	}
}

func TestGenerateFromFormInvalidTemplate(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{
		Templates: map[string]config.Template{
			"basic": {Name: "Basic Resume"},
		},
	}

	gen := generator.New(cfg)

	_, err := gen.GenerateFromForm("nonexistent", nil)
	if err == nil {
		t.Error("Expected error for nonexistent template")
	}

	if !strings.Contains(err.Error(), "template 'nonexistent' not found") {
		t.Errorf("Expected template not found error, got: %v", err)
	}
}

func TestHandlePhotoUploadToWorkDir(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}
	gen := generator.New(cfg)

	workDir := t.TempDir()

	// Test with no file upload
	emptyRequest := &http.Request{}
	filename, err := gen.HandlePhotoUploadToWorkDir(emptyRequest, workDir)
	if err != nil {
		t.Errorf("Expected no error with empty request, got: %v", err)
	}
	if filename != "" {
		t.Errorf("Expected empty filename with no upload, got: %s", filename)
	}

	// Note: Testing actual file upload would require creating a multipart request
	// which is complex. The main logic is tested in the integration tests.
}

func TestCopyPhoto(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}
	gen := generator.New(cfg)

	// Test with no photos directory
	err := gen.CopyPhoto("/nonexistent")
	if err == nil {
		t.Error("Expected error when cv-photos directory doesn't exist")
	}

	// Create test setup in project structure
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	// Create templates directory structure
	templatesDir := filepath.Join(tempDir, "templates")
	templateDir := filepath.Join(templatesDir, "test")
	photosDir := filepath.Join(tempDir, "cv-photos")

	err = os.MkdirAll(templateDir, 0o750)
	if err != nil {
		t.Fatalf("Failed to create template dir: %v", err)
	}

	err = os.MkdirAll(photosDir, 0o750)
	if err != nil {
		t.Fatalf("Failed to create photos dir: %v", err)
	}

	// Test with no photos in directory
	err = gen.CopyPhoto(templateDir)
	if err == nil {
		t.Error("Expected error when no photos found")
	}

	// Create a test photo
	photoPath := filepath.Join(photosDir, "test.jpg")
	err = os.WriteFile(photoPath, []byte("fake image data"), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test photo: %v", err)
	}

	// Test successful copy
	err = gen.CopyPhoto(templateDir)
	if err != nil {
		t.Errorf("Expected successful copy, got error: %v", err)
	}

	// Verify file was copied
	avatarPath := filepath.Join(templateDir, "avatar.png")
	if _, err := os.Stat(avatarPath); os.IsNotExist(err) {
		t.Error("Avatar file was not created")
	}

	// Verify content
	// #nosec G304 - avatarPath is constructed in test, safe
	content, err := os.ReadFile(avatarPath)
	if err != nil {
		t.Fatalf("Failed to read copied avatar: %v", err)
	}

	if string(content) != "fake image data" {
		t.Error("Avatar content doesn't match source")
	}
}

func TestGenerate(t *testing.T) {
	t.Parallel()
	// Create test template structure
	tempDir := t.TempDir()
	templatesDir := filepath.Join(tempDir, "templates")
	templateDir := filepath.Join(templatesDir, "basic")
	outputDir := filepath.Join(tempDir, "output")

	err := os.MkdirAll(templateDir, 0o750)
	if err != nil {
		t.Fatalf("Failed to create template dir: %v", err)
	}

	err = os.MkdirAll(outputDir, 0o750)
	if err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	// Create a simple typst file
	inputFile := filepath.Join(templateDir, "main.typ")
	err = os.WriteFile(inputFile, []byte("#set page(paper: \"a4\")\n= Test CV\nThis is a test."), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test typst file: %v", err)
	}

	cfg := &config.Config{
		Templates: map[string]config.Template{
			"basic": {
				Name:      "Basic Resume",
				Dir:       templateDir,
				InputFile: "main.typ", // Use relative path
			},
		},
		OutputDir: outputDir,
	}

	gen := generator.New(cfg)

	// Test invalid template
	err = gen.Generate(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent template")
	}

	// Test valid template (this will fail if typst is not installed)
	// In a real CI environment, you might want to skip this test or mock the typst command
	err = gen.Generate(context.Background(), "basic")
	if err != nil {
		// If typst is not installed, skip this test
		if strings.Contains(err.Error(), "executable file not found") {
			t.Skip("typst not installed, skipping compilation test")
		}
		t.Errorf("Generate failed: %v", err)
	} else {
		// Verify output file was created
		outputFile := filepath.Join(outputDir, "cv-basic.pdf")
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Error("Output PDF was not created")
		}
	}
}
