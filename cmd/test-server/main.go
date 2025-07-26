package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AlexTLDR/mycv.quest/internal/cv"
)

var cvService *cv.Service

func main() {
	// Initialize CV service
	templatesDir := filepath.Join("assets", "templates", "typst")
	outputDir := filepath.Join("tmp", "cv_output")

	var err error
	cvService, err = cv.NewService(templatesDir, outputDir)
	if err != nil {
		log.Fatal("Failed to initialize CV service:", err)
	}

	// Routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/templates", templatesHandler)
	http.HandleFunc("/api/templates/", templateHandler)
	http.HandleFunc("/api/generate/", generateHandler)
	http.HandleFunc("/api/sample/", sampleHandler)
	http.HandleFunc("/api/form/", formHandler)

	fmt.Println("🚀 CV Generation Test Server")
	fmt.Println("============================")
	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("")
	fmt.Println("Available endpoints:")
	fmt.Println("  GET  /                     - This help page")
	fmt.Println("  GET  /api/templates        - List all templates")
	fmt.Println("  GET  /api/templates/{id}   - Get template details")
	fmt.Println("  GET  /api/form/{id}        - Get template form")
	fmt.Println("  GET  /api/sample/{id}      - Generate CV with sample data")
	fmt.Println("  POST /api/generate/{id}    - Generate CV with custom data")
	fmt.Println("")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <title>CV Generation Test Server</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        .endpoint { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 5px; }
        .method { color: #666; font-weight: bold; }
        button { background: #007cba; color: white; border: none; padding: 8px 16px; margin: 5px; border-radius: 4px; cursor: pointer; }
        button:hover { background: #005a8b; }
        .result { background: #f9f9f9; border: 1px solid #ddd; padding: 10px; margin: 10px 0; border-radius: 4px; }
        pre { overflow-x: auto; }
    </style>
</head>
<body>
    <h1>🚀 CV Generation Test Server</h1>

    <h2>Quick Actions</h2>
    <button onclick="listTemplates()">List Templates</button>
    <button onclick="generateSample('vantage')">Generate Vantage Sample</button>
    <button onclick="generateSample('grotesk')">Generate Grotesk Sample</button>
    <button onclick="getForm('vantage')">Get Vantage Form</button>

    <div id="result" class="result" style="display:none;">
        <h3>Result:</h3>
        <pre id="resultContent"></pre>
    </div>

    <h2>Available Endpoints</h2>

    <div class="endpoint">
        <span class="method">GET</span> <code>/api/templates</code><br>
        Lists all available CV templates
    </div>

    <div class="endpoint">
        <span class="method">GET</span> <code>/api/templates/{id}</code><br>
        Get details for a specific template (e.g., /api/templates/vantage)
    </div>

    <div class="endpoint">
        <span class="method">GET</span> <code>/api/form/{id}</code><br>
        Get the form structure for a template (e.g., /api/form/vantage)
    </div>

    <div class="endpoint">
        <span class="method">GET</span> <code>/api/sample/{id}</code><br>
        Generate a CV using sample data (e.g., /api/sample/vantage) - Downloads PDF
    </div>

    <div class="endpoint">
        <span class="method">POST</span> <code>/api/generate/{id}</code><br>
        Generate a CV with custom data. Send JSON in request body.
    </div>

    <h2>Example Usage</h2>
    <p>To generate a CV with custom data, POST JSON to <code>/api/generate/vantage</code>:</p>
    <pre>{
  "template_id": "vantage",
  "data": {
    "contacts": {
      "name": "Your Name",
      "email": "your.email@example.com",
      "title": "Your Title"
    },
    "position": "Target Position",
    "tagline": "Your professional summary..."
  }
}</pre>

    <script>
        function showResult(content) {
            document.getElementById('result').style.display = 'block';
            document.getElementById('resultContent').textContent = content;
        }

        function listTemplates() {
            fetch('/api/templates')
                .then(response => response.json())
                .then(data => showResult(JSON.stringify(data, null, 2)))
                .catch(error => showResult('Error: ' + error));
        }

        function generateSample(templateId) {
            window.open('/api/sample/' + templateId, '_blank');
        }

        function getForm(templateId) {
            fetch('/api/form/' + templateId)
                .then(response => response.json())
                .then(data => showResult(JSON.stringify(data, null, 2)))
                .catch(error => showResult('Error: ' + error));
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func templatesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	templates, err := cvService.ListTemplates()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list templates: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
	})
}

func templateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	templateID := strings.TrimPrefix(r.URL.Path, "/api/templates/")
	if templateID == "" {
		http.Error(w, "Template ID required", http.StatusBadRequest)
		return
	}

	template, err := cvService.GetTemplate(templateID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template not found: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	templateID := strings.TrimPrefix(r.URL.Path, "/api/form/")
	if templateID == "" {
		http.Error(w, "Template ID required", http.StatusBadRequest)
		return
	}

	form, err := cvService.GenerateForm(templateID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate form: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(form)
}

func sampleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	templateID := strings.TrimPrefix(r.URL.Path, "/api/sample/")
	if templateID == "" {
		http.Error(w, "Template ID required", http.StatusBadRequest)
		return
	}

	// Get sample data
	sampleData, err := cvService.GetSampleData(templateID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get sample data: %v", err), http.StatusInternalServerError)
		return
	}

	// Create realistic sample data
	switch templateID {
	case "vantage":
		sampleData.Data = createVantageSampleData()
	case "grotesk":
		sampleData.Data = createGroteskSampleData()
	}

	// Generate CV
	request := cv.GenerationRequest{
		TemplateID: templateID,
		Data:       *sampleData,
		Format:     "pdf",
	}

	result, err := cvService.GenerateCV(request)
	if err != nil {
		http.Error(w, fmt.Sprintf("CV generation failed: %v", err), http.StatusInternalServerError)
		return
	}

	if !result.Success {
		http.Error(w, result.Message, http.StatusBadRequest)
		return
	}

	// Return PDF
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=sample_%s.pdf", templateID))
	w.Header().Set("Content-Length", strconv.Itoa(len(result.Data)))
	w.Write(result.Data)
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	templateID := strings.TrimPrefix(r.URL.Path, "/api/generate/")
	if templateID == "" {
		http.Error(w, "Template ID required", http.StatusBadRequest)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var templateData cv.TemplateData
	if err := json.NewDecoder(r.Body).Decode(&templateData); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	templateData.TemplateID = templateID

	request := cv.GenerationRequest{
		TemplateID: templateID,
		Data:       templateData,
		Format:     "pdf",
	}

	result, err := cvService.GenerateCV(request)
	if err != nil {
		http.Error(w, fmt.Sprintf("CV generation failed: %v", err), http.StatusInternalServerError)
		return
	}

	if !result.Success {
		http.Error(w, result.Message, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", result.Filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(result.Data)))
	w.Write(result.Data)
}

func createVantageSampleData() map[string]interface{} {
	return map[string]interface{}{
		"contacts": map[string]interface{}{
			"name":     "John Doe",
			"title":    "Senior Software Engineer",
			"email":    "john.doe@example.com",
			"address":  "San Francisco, CA",
			"location": "United States",
			"linkedin": map[string]interface{}{
				"url":         "https://linkedin.com/in/johndoe",
				"displayText": "johndoe",
			},
			"github": map[string]interface{}{
				"url":         "https://github.com/johndoe",
				"displayText": "@johndoe",
			},
			"website": map[string]interface{}{
				"url":         "https://johndoe.dev",
				"displayText": "johndoe.dev",
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
				"location": "San Francisco, CA",
			},
		},
		"education": []interface{}{
			map[string]interface{}{
				"place": map[string]interface{}{
					"name": "University of California, Berkeley",
					"link": "https://berkeley.edu",
				},
				"degree":   "B.S.",
				"major":    "Computer Science",
				"from":     "2015",
				"to":       "2019",
				"location": "Berkeley, CA",
			},
		},
		"technical_expertise": []interface{}{
			map[string]interface{}{"name": "Go", "level": 5},
			map[string]interface{}{"name": "JavaScript", "level": 4},
			map[string]interface{}{"name": "Python", "level": 4},
		},
		"skills": []interface{}{
			"Go", "JavaScript", "Python", "React", "Node.js", "Kubernetes", "Docker", "PostgreSQL",
		},
		"methodology": []interface{}{
			"Agile", "Scrum", "Test-Driven Development",
		},
		"tools": []interface{}{
			"VS Code", "Git", "Docker", "Kubernetes",
		},
		"achievements": []interface{}{
			map[string]interface{}{
				"name":        "AWS Certified Solutions Architect",
				"description": "Professional certification demonstrating expertise in designing distributed systems on AWS",
			},
		},
	}
}

func createGroteskSampleData() map[string]interface{} {
	return map[string]interface{}{
		"personal": map[string]interface{}{
			"info": map[string]interface{}{
				"name":     "Jane Smith",
				"email":    "jane.smith@example.com",
				"phone":    "+1 (555) 123-4567",
				"address":  "New York, NY",
				"linkedin": "https://linkedin.com/in/janesmith",
				"github":   "https://github.com/janesmith",
				"website":  "https://janesmith.dev",
			},
		},
		"settings": map[string]interface{}{
			"language":        "en",
			"include_icons":   true,
			"left_pane_width": "60%",
			"font":            "HK Grotesk",
		},
		"sections": map[string]interface{}{
			"profile": map[string]interface{}{
				"icon": "user",
				"title": map[string]interface{}{
					"en": "Summary",
				},
				"content": map[string]interface{}{
					"en": "Creative and detail-oriented UX/UI designer with 6+ years of experience creating intuitive digital experiences.",
				},
			},
		},
		"layout": map[string]interface{}{
			"left_pane":  []interface{}{"profile", "experience"},
			"right_pane": []interface{}{"education", "skills"},
		},
	}
}
