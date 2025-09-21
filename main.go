package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/AlexTLDR/mycv.quest/templates"
	"gopkg.in/yaml.v2"
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

func (cv *CVGenerator) getTemplateData() []templates.CVTemplate {
	var templateData []templates.CVTemplate

	descriptions := map[string]string{
		"vantage": "Clean and professional design with modern typography",
		"basic":   "Simple and elegant layout perfect for any industry",
		"modern":  "Contemporary design with visual elements and photo support",
	}

	for key, template := range cv.Templates {
		pdfPath := fmt.Sprintf("/static/output/cv-%s.pdf", key)
		thumbnailPath := ""

		// Check for thumbnail images
		if _, err := os.Stat(filepath.Join(template.Dir, "thumbnail.png")); err == nil {
			thumbnailPath = fmt.Sprintf("/static/%s/thumbnail.png", template.Dir)
		} else if _, err := os.Stat(filepath.Join(template.Dir, "screenshot.png")); err == nil {
			thumbnailPath = fmt.Sprintf("/static/%s/screenshot.png", template.Dir)
		}

		templateData = append(templateData, templates.CVTemplate{
			Key:           key,
			Name:          template.Name,
			Description:   descriptions[key],
			PDFPath:       pdfPath,
			ThumbnailPath: thumbnailPath,
		})
	}

	return templateData
}

func (cv *CVGenerator) setupRoutes() {
	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("."))))

	// Home page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		templateData := cv.getTemplateData()
		templates.Index(templateData).Render(r.Context(), w)
	})

	// Form endpoints
	http.HandleFunc("/form/", func(w http.ResponseWriter, r *http.Request) {
		templateKey := strings.TrimPrefix(r.URL.Path, "/form/")

		switch templateKey {
		case "basic":
			templates.BasicForm().Render(r.Context(), w)
		case "modern":
			templates.ModernForm().Render(r.Context(), w)
		case "vantage":
			templates.VantageForm().Render(r.Context(), w)
		default:
			http.NotFound(w, r)
		}
	})

	// Generate CV endpoint
	http.HandleFunc("/generate/", func(w http.ResponseWriter, r *http.Request) {
		templateKey := strings.TrimPrefix(r.URL.Path, "/generate/")

		if r.Method == "GET" {
			// Redirect to form page
			http.Redirect(w, r, fmt.Sprintf("/form/%s", templateKey), http.StatusSeeOther)
			return
		}

		if r.Method == "POST" {
			if err := cv.GenerateFromForm(templateKey, r); err != nil {
				http.Error(w, fmt.Sprintf("Error generating CV: %v", err), http.StatusInternalServerError)
				return
			}

			// Redirect to the generated PDF
			http.Redirect(w, r, fmt.Sprintf("/static/output/cv-%s.pdf", templateKey), http.StatusSeeOther)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
}

func (cv *CVGenerator) GenerateFromForm(templateKey string, r *http.Request) error {
	template, exists := cv.Templates[templateKey]
	if !exists {
		return fmt.Errorf("template '%s' not found", templateKey)
	}

	if err := cv.ensureOutputDir(); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Parse form data - handle both multipart and regular forms
	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
			return fmt.Errorf("failed to parse multipart form: %w", err)
		}
	} else {
		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("failed to parse form: %w", err)
		}
	}

	// Generate template-specific files
	switch templateKey {
	case "basic":
		return cv.generateBasicCV(template, r)
	case "modern":
		return cv.generateModernCV(template, r)
	case "vantage":
		return cv.generateVantageCV(template, r)
	default:
		return fmt.Errorf("unsupported template: %s", templateKey)
	}
}

func (cv *CVGenerator) generateBasicCV(template Template, r *http.Request) error {
	// Create a unique directory for this generation
	timestamp := time.Now().Format("20060102_150405")
	workDir := filepath.Join("temp", "basic_"+timestamp)
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}
	defer os.RemoveAll(workDir) // Clean up

	// Copy template files
	srcFiles := []string{"main.typ"}
	for _, file := range srcFiles {
		src := filepath.Join(template.Dir, file)
		dst := filepath.Join(workDir, file)
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("failed to copy %s: %w", file, err)
		}
	}

	// Generate main.typ with form data
	typContent := cv.generateBasicTypContent(r)
	if err := os.WriteFile(filepath.Join(workDir, "main.typ"), []byte(typContent), 0o644); err != nil {
		return fmt.Errorf("failed to write main.typ: %w", err)
	}

	// Compile PDF
	outputFile := filepath.Join(cv.OutputDir, "cv-basic.pdf")
	absOutputFile, _ := filepath.Abs(outputFile)

	cmd := exec.Command("typst", "compile", "main.typ", absOutputFile)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("typst compilation failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("CV generated successfully at %s\n", outputFile)
	return nil
}

func (cv *CVGenerator) generateModernCV(template Template, r *http.Request) error {
	// Handle photo upload if present
	if template.NeedsPhoto {
		if err := cv.handlePhotoUpload(r, template.Dir); err != nil {
			return fmt.Errorf("failed to handle photo upload: %w", err)
		}
	}

	// Create a unique directory for this generation
	timestamp := time.Now().Format("20060102_150405")
	workDir := filepath.Join("temp", "modern_"+timestamp)
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}
	defer os.RemoveAll(workDir) // Clean up

	// Copy template files and avatar
	srcFiles := []string{"main.typ", "config.yaml"}
	for _, file := range srcFiles {
		src := filepath.Join(template.Dir, file)
		dst := filepath.Join(workDir, file)
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("failed to copy %s: %w", file, err)
		}
	}

	// Copy avatar if it exists
	avatarSrc := filepath.Join(template.Dir, "avatar.png")
	if _, err := os.Stat(avatarSrc); err == nil {
		avatarDst := filepath.Join(workDir, "avatar.png")
		if err := copyFile(avatarSrc, avatarDst); err != nil {
			return fmt.Errorf("failed to copy avatar: %w", err)
		}
	}

	// Generate main.typ with form data
	typContent := cv.generateModernTypContent(r)
	if err := os.WriteFile(filepath.Join(workDir, "main.typ"), []byte(typContent), 0o644); err != nil {
		return fmt.Errorf("failed to write main.typ: %w", err)
	}

	// Compile PDF
	outputFile := filepath.Join(cv.OutputDir, "cv-modern.pdf")
	absOutputFile, _ := filepath.Abs(outputFile)

	cmd := exec.Command("typst", "compile", "main.typ", absOutputFile)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("typst compilation failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("CV generated successfully at %s\n", outputFile)
	return nil
}

func (cv *CVGenerator) generateVantageCV(template Template, r *http.Request) error {
	// Create a unique directory for this generation
	timestamp := time.Now().Format("20060102_150405")
	workDir := filepath.Join("temp", "vantage_"+timestamp)
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}
	defer os.RemoveAll(workDir) // Clean up

	// Copy template files
	srcFiles := []string{"example.typ", "vantage-typst.typ"}
	for _, file := range srcFiles {
		src := filepath.Join(template.Dir, file)
		dst := filepath.Join(workDir, file)
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("failed to copy %s: %w", file, err)
		}
	}

	// Copy icons directory
	iconsDir := filepath.Join(template.Dir, "icons")
	if _, err := os.Stat(iconsDir); err == nil {
		destIconsDir := filepath.Join(workDir, "icons")
		if err := copyDir(iconsDir, destIconsDir); err != nil {
			return fmt.Errorf("failed to copy icons directory: %w", err)
		}
	}

	// Generate configuration.yaml with form data
	yamlContent := cv.generateVantageYAMLContent(r)
	if err := os.WriteFile(filepath.Join(workDir, "configuration.yaml"), yamlContent, 0o644); err != nil {
		return fmt.Errorf("failed to write configuration.yaml: %w", err)
	}

	// Compile PDF
	outputFile := filepath.Join(cv.OutputDir, "cv-vantage.pdf")
	absOutputFile, _ := filepath.Abs(outputFile)

	cmd := exec.Command("typst", "compile", "example.typ", absOutputFile)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("typst compilation failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("CV generated successfully at %s\n", outputFile)
	return nil
}

// Helper functions for file operations
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

func (cv *CVGenerator) handlePhotoUpload(r *http.Request, templateDir string) error {
	file, _, err := r.FormFile("avatar")
	if err != nil {
		// No file uploaded, use existing avatar if available
		return nil
	}
	defer file.Close()

	// Save uploaded file as avatar.png in template directory
	avatarPath := filepath.Join(templateDir, "avatar.png")
	dest, err := os.Create(avatarPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, file)
	return err
}

// Content generation functions
func (cv *CVGenerator) generateBasicTypContent(r *http.Request) string {
	name := r.FormValue("name")
	location := r.FormValue("location")
	email := r.FormValue("email")
	github := r.FormValue("github")
	linkedin := r.FormValue("linkedin")
	phone := r.FormValue("phone")
	personalSite := r.FormValue("personal_site")
	accentColor := r.FormValue("accent_color")
	if accentColor == "" {
		accentColor = "#26428b"
	}

	content := fmt.Sprintf(`#import "@preview/basic-resume:0.2.8": *

// Personal information
#let name = "%s"
#let location = "%s"
#let email = "%s"
#let github = "%s"
#let linkedin = "%s"
#let phone = "%s"
#let personal-site = "%s"

#show: resume.with(
  author: name,
  location: location,
  email: email,
  github: github,
  linkedin: linkedin,
  phone: phone,
  personal-site: personal-site,
  accent-color: "%s",
  font: "New Computer Modern",
  paper: "us-letter",
  author-position: left,
  personal-info-position: left,
)

`, name, location, email, github, linkedin, phone, personalSite, accentColor)

	// Add education section
	content += "== Education\n\n"
	for i := 0; ; i++ {
		institution := r.FormValue(fmt.Sprintf("education[%d][institution]", i))
		if institution == "" {
			break
		}
		location := r.FormValue(fmt.Sprintf("education[%d][location]", i))
		startDate := r.FormValue(fmt.Sprintf("education[%d][start_date]", i))
		endDate := r.FormValue(fmt.Sprintf("education[%d][end_date]", i))
		degree := r.FormValue(fmt.Sprintf("education[%d][degree]", i))
		details := r.FormValue(fmt.Sprintf("education[%d][details]", i))

		content += fmt.Sprintf(`#edu(
  institution: "%s",
  location: "%s",
  dates: dates-helper(start-date: "%s", end-date: "%s"),
  degree: "%s",
)
`, institution, location, startDate, endDate, degree)

		if details != "" {
			lines := strings.Split(details, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					content += fmt.Sprintf("- %s\n", strings.TrimSpace(line))
				}
			}
		}
		content += "\n"
	}

	// Add work experience section
	content += "== Work Experience\n\n"
	for i := 0; ; i++ {
		title := r.FormValue(fmt.Sprintf("work[%d][title]", i))
		if title == "" {
			break
		}
		company := r.FormValue(fmt.Sprintf("work[%d][company]", i))
		location := r.FormValue(fmt.Sprintf("work[%d][location]", i))
		startDate := r.FormValue(fmt.Sprintf("work[%d][start_date]", i))
		endDate := r.FormValue(fmt.Sprintf("work[%d][end_date]", i))
		description := r.FormValue(fmt.Sprintf("work[%d][description]", i))

		content += fmt.Sprintf(`#work(
  title: "%s",
  location: "%s",
  company: "%s",
  dates: dates-helper(start-date: "%s", end-date: "%s"),
)
`, title, location, company, startDate, endDate)

		if description != "" {
			lines := strings.Split(description, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					content += fmt.Sprintf("%s\n", strings.TrimSpace(line))
				}
			}
		}
		content += "\n"
	}

	// Add projects section
	content += "== Projects\n\n"
	for i := 0; ; i++ {
		name := r.FormValue(fmt.Sprintf("projects[%d][name]", i))
		if name == "" {
			break
		}
		role := r.FormValue(fmt.Sprintf("projects[%d][role]", i))
		startDate := r.FormValue(fmt.Sprintf("projects[%d][start_date]", i))
		endDate := r.FormValue(fmt.Sprintf("projects[%d][end_date]", i))
		url := r.FormValue(fmt.Sprintf("projects[%d][url]", i))
		description := r.FormValue(fmt.Sprintf("projects[%d][description]", i))

		projectCall := "#project(\n  name: \"" + name + "\","
		if role != "" {
			projectCall += "\n  role: \"" + role + "\","
		}
		if startDate != "" {
			if endDate != "" {
				projectCall += fmt.Sprintf("\n  dates: dates-helper(start-date: \"%s\", end-date: \"%s\"),", startDate, endDate)
			} else {
				projectCall += fmt.Sprintf("\n  dates: dates-helper(start-date: \"%s\"),", startDate)
			}
		}
		if url != "" {
			projectCall += "\n  url: \"" + url + "\","
		}
		projectCall += "\n)\n"

		content += projectCall

		if description != "" {
			lines := strings.Split(description, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					content += fmt.Sprintf("%s\n", strings.TrimSpace(line))
				}
			}
		}
		content += "\n"
	}

	// Add skills section
	programmingLanguages := r.FormValue("programming_languages")
	technologies := r.FormValue("technologies")

	if programmingLanguages != "" || technologies != "" {
		content += "== Skills\n"
		if programmingLanguages != "" {
			content += fmt.Sprintf("- *Programming Languages*: %s\n", programmingLanguages)
		}
		if technologies != "" {
			content += fmt.Sprintf("- *Technologies*: %s\n", technologies)
		}
	}

	return content
}

func (cv *CVGenerator) generateModernTypContent(r *http.Request) string {
	author := r.FormValue("author")
	jobTitle := r.FormValue("job_title")
	bio := r.FormValue("bio")
	email := r.FormValue("email")
	mobile := r.FormValue("mobile")
	location := r.FormValue("location")
	linkedin := r.FormValue("linkedin")
	github := r.FormValue("github")
	website := r.FormValue("website")

	content := fmt.Sprintf(`#import "@preview/modern-resume:0.1.0": modern-resume, experience-work, experience-edu, project, pill

#show: modern-resume.with(
  author: "%s",
  job-title: "%s",
  bio: [%s],
  avatar: image("avatar.png"),
  contact-options: (
    email: link("mailto:%s")[%s],
`, author, jobTitle, bio, email, strings.ReplaceAll(email, "@", "\\@"))

	if mobile != "" {
		content += fmt.Sprintf("    mobile: \"%s\",\n", mobile)
	}
	if location != "" {
		content += fmt.Sprintf("    location: \"%s\",\n", location)
	}
	if linkedin != "" {
		content += fmt.Sprintf("    linkedin: link(\"https://www.linkedin.com/in/%s\")[linkedin/%s],\n", linkedin, linkedin)
	}
	if github != "" {
		content += fmt.Sprintf("    github: link(\"%s\")[%s],\n", github, github)
	}
	if website != "" {
		content += fmt.Sprintf("    website: link(\"https://%s\")[%s],\n", website, website)
	}

	content += "  ),\n)\n\n"

	// Add education section
	content += "== Education\n\n"
	for i := 0; ; i++ {
		title := r.FormValue(fmt.Sprintf("education[%d][title]", i))
		if title == "" {
			break
		}
		subtitle := r.FormValue(fmt.Sprintf("education[%d][subtitle]", i))
		dateFrom := r.FormValue(fmt.Sprintf("education[%d][date_from]", i))
		dateTo := r.FormValue(fmt.Sprintf("education[%d][date_to]", i))
		taskDescription := r.FormValue(fmt.Sprintf("education[%d][task_description]", i))

		content += fmt.Sprintf(`#experience-edu(
  title: "%s",
  subtitle: "%s",
  task-description: [
`, title, subtitle)

		if taskDescription != "" {
			lines := strings.Split(taskDescription, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					content += fmt.Sprintf("    %s\n", strings.TrimSpace(line))
				}
			}
		}

		content += "  ],\n"
		if dateFrom != "" {
			content += fmt.Sprintf("  date-from: \"%s\",\n", dateFrom)
		}
		if dateTo != "" {
			content += fmt.Sprintf("  date-to: \"%s\",\n", dateTo)
		}
		content += ")\n\n"
	}

	// Add work experience section
	content += "== Work experience\n\n"
	for i := 0; ; i++ {
		title := r.FormValue(fmt.Sprintf("work[%d][title]", i))
		if title == "" {
			break
		}
		subtitle := r.FormValue(fmt.Sprintf("work[%d][subtitle]", i))
		facilityDescription := r.FormValue(fmt.Sprintf("work[%d][facility_description]", i))
		dateFrom := r.FormValue(fmt.Sprintf("work[%d][date_from]", i))
		dateTo := r.FormValue(fmt.Sprintf("work[%d][date_to]", i))
		taskDescription := r.FormValue(fmt.Sprintf("work[%d][task_description]", i))

		content += fmt.Sprintf(`#experience-work(
  title: "%s",
  subtitle: "%s",
  facility-description: "%s",
  task-description: [
`, title, subtitle, facilityDescription)

		if taskDescription != "" {
			lines := strings.Split(taskDescription, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					content += fmt.Sprintf("    %s\n", strings.TrimSpace(line))
				}
			}
		}

		content += "  ],\n"
		if dateFrom != "" {
			content += fmt.Sprintf("  date-from: \"%s\",\n", dateFrom)
		}
		if dateTo != "" {
			content += fmt.Sprintf("  date-to: \"%s\",\n", dateTo)
		}
		content += ")\n\n"
	}

	content += "#colbreak()\n\n"

	// Add skills section
	skills := r.FormValue("skills")
	if skills != "" {
		content += "== Skills\n\n"
		skillList := strings.Split(skills, ",")
		for _, skill := range skillList {
			skill = strings.TrimSpace(skill)
			if skill != "" {
				content += fmt.Sprintf("#pill(\"%s\", fill: true)\n", skill)
			}
		}
		content += "\n"
	}

	// Add projects section
	content += "== Projects\n\n"
	for i := 0; ; i++ {
		title := r.FormValue(fmt.Sprintf("projects[%d][title]", i))
		if title == "" {
			break
		}
		subtitle := r.FormValue(fmt.Sprintf("projects[%d][subtitle]", i))
		dateFrom := r.FormValue(fmt.Sprintf("projects[%d][date_from]", i))
		dateTo := r.FormValue(fmt.Sprintf("projects[%d][date_to]", i))
		description := r.FormValue(fmt.Sprintf("projects[%d][description]", i))

		content += fmt.Sprintf(`#project(
  title: "%s",
`, title)

		if subtitle != "" {
			content += fmt.Sprintf("  subtitle: \"%s\",\n", subtitle)
		}

		if description != "" {
			content += "  description: [\n"
			lines := strings.Split(description, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					content += fmt.Sprintf("    %s\n", strings.TrimSpace(line))
				}
			}
			content += "  ],\n"
		}

		if dateFrom != "" {
			content += fmt.Sprintf("  date-from: \"%s\",\n", dateFrom)
		}
		if dateTo != "" {
			content += fmt.Sprintf("  date-to: \"%s\",\n", dateTo)
		}
		content += ")\n\n"
	}

	// Add certificates section
	content += "== Certificates\n\n"
	for i := 0; ; i++ {
		title := r.FormValue(fmt.Sprintf("certificates[%d][title]", i))
		if title == "" {
			break
		}
		subtitle := r.FormValue(fmt.Sprintf("certificates[%d][subtitle]", i))
		dateFrom := r.FormValue(fmt.Sprintf("certificates[%d][date_from]", i))
		dateTo := r.FormValue(fmt.Sprintf("certificates[%d][date_to]", i))

		content += fmt.Sprintf(`#project(
  title: "%s",
`, title)

		if subtitle != "" {
			content += fmt.Sprintf("  subtitle: \"%s\",\n", subtitle)
		}
		if dateFrom != "" {
			content += fmt.Sprintf("  date-from: \"%s\",\n", dateFrom)
		}
		if dateTo != "" {
			content += fmt.Sprintf("  date-to: \"%s\",\n", dateTo)
		}
		content += ")\n\n"
	}

	// Add languages section
	languages := r.FormValue("languages")
	if languages != "" {
		content += "== Languages\n\n"
		langList := strings.Split(languages, ",")
		for _, lang := range langList {
			lang = strings.TrimSpace(lang)
			if lang != "" {
				content += fmt.Sprintf("#pill(\"%s\")\n", lang)
			}
		}
		content += "\n"
	}

	// Add interests section
	interests := r.FormValue("interests")
	if interests != "" {
		content += "== Interests\n\n"
		interestList := strings.Split(interests, ",")
		for _, interest := range interestList {
			interest = strings.TrimSpace(interest)
			if interest != "" {
				content += fmt.Sprintf("#pill(\"%s\")\n", interest)
			}
		}
	}

	return content
}

func (cv *CVGenerator) generateVantageYAMLContent(r *http.Request) []byte {
	data := map[string]interface{}{
		"contacts": map[string]interface{}{
			"name":     r.FormValue("name"),
			"title":    r.FormValue("title"),
			"email":    r.FormValue("email"),
			"address":  r.FormValue("address"),
			"location": r.FormValue("location"),
			"linkedin": map[string]string{
				"url":         r.FormValue("linkedin_url"),
				"displayText": r.FormValue("linkedin_display_text"),
			},
			"github": map[string]string{
				"url":         r.FormValue("github_url"),
				"displayText": r.FormValue("github_display_text"),
			},
			"website": map[string]string{
				"url":         r.FormValue("website_url"),
				"displayText": r.FormValue("website_display_text"),
			},
		},
		"position":  r.FormValue("position"),
		"tagline":   r.FormValue("tagline"),
		"objective": r.FormValue("objective"),
	}

	// Parse jobs
	var jobs []map[string]interface{}
	for i := 0; ; i++ {
		position := r.FormValue(fmt.Sprintf("jobs[%d][position]", i))
		if position == "" {
			break
		}

		job := map[string]interface{}{
			"position": position,
			"company": map[string]string{
				"name": r.FormValue(fmt.Sprintf("jobs[%d][company_name]", i)),
				"link": r.FormValue(fmt.Sprintf("jobs[%d][company_link]", i)),
			},
			"product": map[string]string{
				"name": r.FormValue(fmt.Sprintf("jobs[%d][product_name]", i)),
				"link": r.FormValue(fmt.Sprintf("jobs[%d][product_link]", i)),
			},
			"from":     r.FormValue(fmt.Sprintf("jobs[%d][from]", i)),
			"to":       r.FormValue(fmt.Sprintf("jobs[%d][to]", i)),
			"location": r.FormValue(fmt.Sprintf("jobs[%d][location]", i)),
		}

		// Parse description
		description := r.FormValue(fmt.Sprintf("jobs[%d][description]", i))
		if description != "" {
			lines := strings.Split(description, "\n")
			var descList []string
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					if !strings.HasPrefix(line, "-") {
						line = "- " + line
					}
					descList = append(descList, strings.TrimPrefix(line, "- "))
				}
			}
			job["description"] = descList
		}

		// Parse tags
		tags := r.FormValue(fmt.Sprintf("jobs[%d][tags]", i))
		if tags != "" {
			tagList := strings.Split(tags, ",")
			var cleanTags []string
			for _, tag := range tagList {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					cleanTags = append(cleanTags, tag)
				}
			}
			job["tags"] = cleanTags
		}

		jobs = append(jobs, job)
	}
	if jobs == nil {
		jobs = []map[string]interface{}{}
	}
	data["jobs"] = jobs

	// Parse education
	var education []map[string]interface{}
	for i := 0; ; i++ {
		placeName := r.FormValue(fmt.Sprintf("education[%d][place_name]", i))
		if placeName == "" {
			break
		}

		edu := map[string]interface{}{
			"place": map[string]string{
				"name": placeName,
				"link": r.FormValue(fmt.Sprintf("education[%d][place_link]", i)),
			},
			"degree":   r.FormValue(fmt.Sprintf("education[%d][degree]", i)),
			"major":    r.FormValue(fmt.Sprintf("education[%d][major]", i)),
			"track":    r.FormValue(fmt.Sprintf("education[%d][track]", i)),
			"from":     r.FormValue(fmt.Sprintf("education[%d][from]", i)),
			"to":       r.FormValue(fmt.Sprintf("education[%d][to]", i)),
			"location": r.FormValue(fmt.Sprintf("education[%d][location]", i)),
		}

		education = append(education, edu)
	}
	if education == nil {
		education = []map[string]interface{}{}
	}
	data["education"] = education

	// Parse technical expertise
	var technicalExpertise []map[string]interface{}
	for i := 0; ; i++ {
		name := r.FormValue(fmt.Sprintf("technical_expertise[%d][name]", i))
		if name == "" {
			break
		}

		levelStr := r.FormValue(fmt.Sprintf("technical_expertise[%d][level]", i))
		level, _ := strconv.Atoi(levelStr)
		if level == 0 {
			level = 4 // default
		}

		technicalExpertise = append(technicalExpertise, map[string]interface{}{
			"name":  name,
			"level": level,
		})
	}
	if technicalExpertise == nil {
		technicalExpertise = []map[string]interface{}{}
	}
	data["technical_expertise"] = technicalExpertise

	// Parse achievements
	var achievements []map[string]interface{}
	for i := 0; ; i++ {
		name := r.FormValue(fmt.Sprintf("achievements[%d][name]", i))
		if name == "" {
			break
		}

		achievements = append(achievements, map[string]interface{}{
			"name":        name,
			"description": r.FormValue(fmt.Sprintf("achievements[%d][description]", i)),
		})
	}
	if achievements == nil {
		achievements = []map[string]interface{}{}
	}
	data["achievements"] = achievements

	// Parse comma-separated lists with defaults
	skills := r.FormValue("skills")
	var cleanSkills []string
	if skills != "" {
		skillList := strings.Split(skills, ",")
		for _, skill := range skillList {
			skill = strings.TrimSpace(skill)
			if skill != "" {
				cleanSkills = append(cleanSkills, skill)
			}
		}
	}
	data["skills"] = cleanSkills

	methodology := r.FormValue("methodology")
	var cleanMethods []string
	if methodology != "" {
		methodList := strings.Split(methodology, ",")
		for _, method := range methodList {
			method = strings.TrimSpace(method)
			if method != "" {
				cleanMethods = append(cleanMethods, method)
			}
		}
	}
	data["methodology"] = cleanMethods

	tools := r.FormValue("tools")
	var cleanTools []string
	if tools != "" {
		toolList := strings.Split(tools, ",")
		for _, tool := range toolList {
			tool = strings.TrimSpace(tool)
			if tool != "" {
				cleanTools = append(cleanTools, tool)
			}
		}
	}
	data["tools"] = cleanTools

	yamlData, _ := yaml.Marshal(data)
	return yamlData
}

func main() {
	templateFlag := flag.String("template", "vantage", "Template to use (vantage, basic, modern)")
	listFlag := flag.Bool("list", false, "List available templates")
	serveFlag := flag.Bool("serve", false, "Start web server")
	portFlag := flag.String("port", "8080", "Port to serve on")
	flag.Parse()

	generator := NewCVGenerator()

	if *serveFlag {
		generator.setupRoutes()
		fmt.Printf("Starting server on http://localhost:%s\n", *portFlag)
		log.Fatal(http.ListenAndServe(":"+*portFlag, nil))
		return
	}

	if *listFlag {
		generator.ListTemplates()
		return
	}

	if err := generator.Generate(*templateFlag); err != nil {
		log.Fatalf("Error generating CV: %v", err)
	}
}
