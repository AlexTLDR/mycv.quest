package generator

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlexTLDR/mycv.quest/pkg/config"
	"github.com/AlexTLDR/mycv.quest/pkg/utils"
)

func (cv *CVGenerator) generateBasicCV(template config.Template, r *http.Request) ([]byte, error) {
	// Ensure temp directory exists
	if err := os.MkdirAll("temp", 0o755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Create a unique directory for this generation
	timestamp := time.Now().Format("20060102_150405")
	workDir := filepath.Join("temp", "basic_"+timestamp)
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}
	defer os.RemoveAll(workDir) // Clean up

	// Copy template files
	srcFiles := []string{"main.typ"}
	for _, file := range srcFiles {
		src := filepath.Join(template.Dir, file)
		dst := filepath.Join(workDir, file)
		if err := utils.CopyFile(src, dst); err != nil {
			return nil, fmt.Errorf("failed to copy %s: %w", file, err)
		}
	}

	// Generate main.typ with form data
	typContent := cv.generateBasicTypContent(r)
	if err := os.WriteFile(filepath.Join(workDir, "main.typ"), []byte(typContent), 0o644); err != nil {
		return nil, fmt.Errorf("failed to write main.typ: %w", err)
	}

	// Compile PDF to temporary file
	outputFile := filepath.Join(workDir, "output.pdf")
	absOutputFile, err := filepath.Abs(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for output file: %w", err)
	}

	cmd := exec.Command("typst", "compile", "main.typ", absOutputFile)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("typst compilation failed: %w\nOutput: %s", err, string(output))
	}

	// Read the generated PDF into memory
	pdfData, err := os.ReadFile(absOutputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read generated PDF: %w", err)
	}

	fmt.Printf("CV generated successfully in memory\n")
	return pdfData, nil
}

func (cv *CVGenerator) generateBasicTypContent(r *http.Request) string {
	name := utils.SanitizeFormValue(r.FormValue("name"))
	location := utils.SanitizeFormValue(r.FormValue("location"))
	email := utils.SanitizeFormValue(r.FormValue("email"))
	github := utils.NormalizeURL(utils.SanitizeFormValue(r.FormValue("github")))
	linkedin := utils.NormalizeURL(utils.SanitizeFormValue(r.FormValue("linkedin")))
	phone := utils.SanitizeFormValue(r.FormValue("phone"))
	personalSite := utils.NormalizeURL(utils.SanitizeFormValue(r.FormValue("personal_site")))
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
		institution := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("education[%d][institution]", i)))
		if institution == "" {
			break
		}
		location := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("education[%d][location]", i)))
		startDate := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("education[%d][start_date]", i)))
		endDate := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("education[%d][end_date]", i)))
		degree := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("education[%d][degree]", i)))
		details := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("education[%d][details]", i)))

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
		title := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("work[%d][title]", i)))
		if title == "" {
			break
		}
		company := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("work[%d][company]", i)))
		location := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("work[%d][location]", i)))
		startDate := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("work[%d][start_date]", i)))
		endDate := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("work[%d][end_date]", i)))
		description := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("work[%d][description]", i)))

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
		name := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("projects[%d][name]", i)))
		if name == "" {
			break
		}
		role := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("projects[%d][role]", i)))
		startDate := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("projects[%d][start_date]", i)))
		endDate := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("projects[%d][end_date]", i)))
		url := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("projects[%d][url]", i)))
		description := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("projects[%d][description]", i)))

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
	programmingLanguages := utils.SanitizeFormValue(r.FormValue("programming_languages"))
	technologies := utils.SanitizeFormValue(r.FormValue("technologies"))

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
