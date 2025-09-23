package generator

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/AlexTLDR/mycv.quest/pkg/config"
	"github.com/AlexTLDR/mycv.quest/pkg/utils"
	"gopkg.in/yaml.v2"
)

func (cv *CVGenerator) GenerateVantageCV(template config.Template, r *http.Request) ([]byte, error) {
	// Ensure temp directory exists
	if err := os.MkdirAll("temp", 0o755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Create a unique directory for this generation
	timestamp := time.Now().Format("20060102_150405")
	workDir := filepath.Join("temp", "vantage_"+timestamp)
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}
	defer os.RemoveAll(workDir) // Clean up

	// Copy template files
	srcFiles := []string{"example.typ", "vantage-typst.typ"}
	for _, file := range srcFiles {
		src := filepath.Join(template.Dir, file)
		dst := filepath.Join(workDir, file)
		if err := utils.CopyFile(src, dst); err != nil {
			return nil, fmt.Errorf("failed to copy %s: %w", file, err)
		}
	}

	// Copy icons directory
	iconsDir := filepath.Join(template.Dir, "icons")
	if _, err := os.Stat(iconsDir); err == nil {
		destIconsDir := filepath.Join(workDir, "icons")
		if err := utils.CopyDir(iconsDir, destIconsDir); err != nil {
			return nil, fmt.Errorf("failed to copy icons directory: %w", err)
		}
	}

	// Generate configuration.yaml with form data
	yamlContent := cv.GenerateVantageYAMLContent(r)
	if err := os.WriteFile(filepath.Join(workDir, "configuration.yaml"), yamlContent, 0o644); err != nil {
		return nil, fmt.Errorf("failed to write configuration.yaml: %w", err)
	}

	// Compile PDF to temporary file
	outputFile := filepath.Join(workDir, "output.pdf")
	absOutputFile, err := filepath.Abs(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for output file: %w", err)
	}

	cmd := exec.Command("typst", "compile", "example.typ", absOutputFile)
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

func (cv *CVGenerator) GenerateVantageYAMLContent(r *http.Request) []byte {
	data := map[string]interface{}{
		"contacts": map[string]interface{}{
			"name":     utils.SanitizeFormValue(r.FormValue("name")),
			"title":    utils.SanitizeFormValue(r.FormValue("title")),
			"email":    utils.SanitizeFormValue(r.FormValue("email")),
			"address":  utils.SanitizeFormValue(r.FormValue("address")),
			"location": utils.SanitizeFormValue(r.FormValue("location")),
			"linkedin": map[string]string{
				"url":         utils.NormalizeURL(utils.SanitizeFormValue(r.FormValue("linkedin_url"))),
				"displayText": utils.SanitizeFormValue(r.FormValue("linkedin_display_text")),
			},
			"github": map[string]string{
				"url":         utils.NormalizeURL(utils.SanitizeFormValue(r.FormValue("github_url"))),
				"displayText": utils.SanitizeFormValue(r.FormValue("github_display_text")),
			},
			"website": map[string]string{
				"url":         utils.NormalizeURL(utils.SanitizeFormValue(r.FormValue("website_url"))),
				"displayText": utils.SanitizeFormValue(r.FormValue("website_display_text")),
			},
		},
		"position":  utils.SanitizeFormValue(r.FormValue("position")),
		"tagline":   utils.SanitizeFormValue(r.FormValue("tagline")),
		"objective": utils.SanitizeFormValue(r.FormValue("objective")),
	}

	// Parse jobs
	var jobs []map[string]interface{}
	for i := 0; ; i++ {
		position := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("jobs[%d][position]", i)))
		if position == "" {
			break
		}

		job := map[string]interface{}{
			"position": position,
			"company": map[string]string{
				"name": utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("jobs[%d][company_name]", i))),
				"link": utils.NormalizeURL(utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("jobs[%d][company_link]", i)))),
			},
			"product": map[string]string{
				"name": utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("jobs[%d][product_name]", i))),
				"link": utils.NormalizeURL(utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("jobs[%d][product_link]", i)))),
			},
			"from":     utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("jobs[%d][from]", i))),
			"to":       utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("jobs[%d][to]", i))),
			"location": utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("jobs[%d][location]", i))),
		}

		// Parse description
		description := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("jobs[%d][description]", i)))
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
		tags := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("jobs[%d][tags]", i)))
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
		placeName := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("education[%d][place_name]", i)))
		if placeName == "" {
			break
		}

		edu := map[string]interface{}{
			"place": map[string]string{
				"name": placeName,
				"link": utils.NormalizeURL(utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("education[%d][place_link]", i)))),
			},
			"degree":   utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("education[%d][degree]", i))),
			"major":    utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("education[%d][major]", i))),
			"track":    utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("education[%d][track]", i))),
			"from":     utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("education[%d][from]", i))),
			"to":       utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("education[%d][to]", i))),
			"location": utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("education[%d][location]", i))),
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
		name := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("technical_expertise[%d][name]", i)))
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
		name := utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("achievements[%d][name]", i)))
		if name == "" {
			break
		}

		achievements = append(achievements, map[string]interface{}{
			"name":        name,
			"description": utils.SanitizeFormValue(r.FormValue(fmt.Sprintf("achievements[%d][description]", i))),
		})
	}
	if achievements == nil {
		achievements = []map[string]interface{}{}
	}
	data["achievements"] = achievements

	// Parse comma-separated lists with defaults
	skills := utils.SanitizeFormValue(r.FormValue("skills"))
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

	methodology := utils.SanitizeFormValue(r.FormValue("methodology"))
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

	tools := utils.SanitizeFormValue(r.FormValue("tools"))
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
