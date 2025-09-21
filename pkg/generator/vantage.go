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

func (cv *CVGenerator) generateVantageCV(template config.Template, r *http.Request) error {
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
		if err := utils.CopyFile(src, dst); err != nil {
			return fmt.Errorf("failed to copy %s: %w", file, err)
		}
	}

	// Copy icons directory
	iconsDir := filepath.Join(template.Dir, "icons")
	if _, err := os.Stat(iconsDir); err == nil {
		destIconsDir := filepath.Join(workDir, "icons")
		if err := utils.CopyDir(iconsDir, destIconsDir); err != nil {
			return fmt.Errorf("failed to copy icons directory: %w", err)
		}
	}

	// Generate configuration.yaml with form data
	yamlContent := cv.generateVantageYAMLContent(r)
	if err := os.WriteFile(filepath.Join(workDir, "configuration.yaml"), yamlContent, 0o644); err != nil {
		return fmt.Errorf("failed to write configuration.yaml: %w", err)
	}

	// Compile PDF
	outputFile := filepath.Join(cv.config.OutputDir, "cv-vantage.pdf")
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
