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

func (cv *CVGenerator) generateModernCV(template config.Template, r *http.Request) ([]byte, error) {
	// Ensure temp directory exists
	if err := os.MkdirAll("temp", 0o755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Create a unique directory for this generation
	timestamp := time.Now().Format("20060102_150405")
	workDir := filepath.Join("temp", "modern_"+timestamp)
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}
	defer os.RemoveAll(workDir) // Clean up

	// Handle photo upload if present
	var avatarFilename string
	photoUploaded := false
	if template.NeedsPhoto {
		filename, err := cv.handlePhotoUploadToWorkDir(r, workDir)
		if err != nil {
			return nil, fmt.Errorf("failed to handle photo upload: %w", err)
		}
		if filename != "" {
			avatarFilename = filename
			photoUploaded = true
		}
	}

	// Copy template files and avatar
	srcFiles := []string{"main.typ", "config.yaml"}
	for _, file := range srcFiles {
		src := filepath.Join(template.Dir, file)
		dst := filepath.Join(workDir, file)
		if err := utils.CopyFile(src, dst); err != nil {
			return nil, fmt.Errorf("failed to copy %s: %w", file, err)
		}
	}

	// Copy template's default avatar only if no photo was uploaded
	if !photoUploaded {
		avatarSrc := filepath.Join(template.Dir, "avatar.png")
		if _, err := os.Stat(avatarSrc); err == nil {
			avatarDst := filepath.Join(workDir, "avatar.png")
			if err := utils.CopyFile(avatarSrc, avatarDst); err != nil {
				return nil, fmt.Errorf("failed to copy avatar: %w", err)
			}
			avatarFilename = "avatar.png"
		}
	}

	// Generate main.typ with form data
	typContent := cv.generateModernTypContent(r, avatarFilename)
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

func (cv *CVGenerator) generateModernTypContent(r *http.Request, avatarFilename string) string {
	author := utils.SanitizeFormValue(r.FormValue("author"))
	jobTitle := utils.SanitizeFormValue(r.FormValue("job_title"))
	bio := utils.SanitizeFormValue(r.FormValue("bio"))
	email := utils.SanitizeFormValue(r.FormValue("email"))
	mobile := utils.SanitizeFormValue(r.FormValue("mobile"))
	location := utils.SanitizeFormValue(r.FormValue("location"))
	linkedin := utils.SanitizeFormValue(r.FormValue("linkedin"))
	github := utils.SanitizeFormValue(r.FormValue("github"))
	website := utils.SanitizeFormValue(r.FormValue("website"))

	content := fmt.Sprintf(`#import "@preview/modern-resume:0.1.0": modern-resume, experience-work, experience-edu, project, pill

#show: modern-resume.with(
  author: "%s",
  job-title: "%s",
  bio: [%s],
  avatar: image("%s"),
  contact-options: (
    email: link("mailto:%s")[%s],
`, author, jobTitle, bio, avatarFilename, email, strings.ReplaceAll(email, "@", "\\@"))

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
		title := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("education[%d][title]", i)))
		if title == "" {
			break
		}
		subtitle := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("education[%d][subtitle]", i)))
		dateFrom := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("education[%d][date_from]", i)))
		dateTo := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("education[%d][date_to]", i)))
		taskDescription := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("education[%d][task_description]", i)))

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
		title := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("work[%d][title]", i)))
		if title == "" {
			break
		}
		subtitle := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("work[%d][subtitle]", i)))
		facilityDescription := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("work[%d][facility_description]", i)))
		dateFrom := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("work[%d][date_from]", i)))
		dateTo := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("work[%d][date_to]", i)))
		taskDescription := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("work[%d][task_description]", i)))

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
	skills := utils.SanitizeFormValue(r.FormValue("skills"))
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
		title := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("projects[%d][title]", i)))
		if title == "" {
			break
		}
		subtitle := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("projects[%d][subtitle]", i)))
		dateFrom := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("projects[%d][date_from]", i)))
		dateTo := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("projects[%d][date_to]", i)))
		description := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("projects[%d][description]", i)))

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
		title := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("certificates[%d][title]", i)))
		if title == "" {
			break
		}
		subtitle := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("certificates[%d][subtitle]", i)))
		dateFrom := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("certificates[%d][date_from]", i)))
		dateTo := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("certificates[%d][date_to]", i)))

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
	languages := utils.SanitizeFormValue(r.FormValue("languages"))
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
	interests := utils.SanitizeFormValue(r.FormValue("interests"))
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
