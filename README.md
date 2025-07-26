# MyCV Quest

A dynamic CV generation system built with Go, Typst, and Templ that allows users to create professional CVs using multiple templates with a single data structure.

## 🎯 Overview

MyCV Quest is a modern CV builder that features a **universal template system** - users enter their data once and can generate CVs in different designs instantly. The key innovation is the template adapter system that enables any CV data to work with any template through automatic data conversion.

### Key Features

- 🔄 **Universal Data Format**: Same CV data works with all templates
- 📄 **Professional PDF Generation**: High-quality output via Typst compilation  
- 🎨 **Multiple Templates**: ATS-friendly and modern design options
- 🌐 **Complete Web API**: RESTful endpoints for CV generation
- 🧪 **Zero Template Modification**: Use official Typst templates as-is
- ⚡ **Dynamic Discovery**: Automatically detects and configures new templates

## 🏗️ Architecture

### Core Components

```
mycv.quest/
├── internal/cv/           # Core CV generation logic
│   ├── service.go         # Main service interface
│   ├── generator.go       # PDF generation with Typst
│   ├── parser.go          # Template configuration parsing
│   ├── adapter.go         # Template data conversion system
│   ├── types.go           # Data structures
│   └── service_test.go    # Comprehensive tests
├── assets/
│   ├── templates/typst/   # CV templates
│   │   ├── vantage/       # ATS-friendly template
│   │   └── grotesk/       # Modern design template
│   ├── templates/templ/   # Web templates
│   ├── static/            # Static assets
│   └── emails/templ/      # Email templates
├── cmd/
│   ├── web/               # Main web application
│   └── test-server/       # Development API server
└── internal/              # Supporting packages
    ├── database/          # Database operations
    ├── smtp/              # Email functionality
    └── response/          # HTTP helpers
```

### Template Adapter System

The core innovation enabling **universal CV data compatibility**:

1. **Common Data Format**: All templates work with the same data structure
2. **Automatic Conversion**: Adapters convert between template-specific formats
3. **No Template Modification**: Official templates work as-is
4. **Extensible**: Easy to add new templates without changing existing code

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- [Typst CLI](https://typst.app/docs/tutorial/installation/) (`typst` command available in PATH)
- [Task](https://taskfile.dev/installation/) (install as `go-task`)
- [Templ](https://templ.guide/quick-start/installation)
- PostgreSQL (for production) or Docker (for development)

### Basic CV Generation

```go
package main

import (
    "os"
    "path/filepath"
    "github.com/AlexTLDR/mycv.quest/internal/cv"
)

func main() {
    // Initialize service
    templatesDir := filepath.Join("assets", "templates", "typst")
    outputDir := filepath.Join("tmp", "cv_output")
    
    service, err := cv.NewService(templatesDir, outputDir)
    if err != nil {
        panic(err)
    }
    
    // Create CV data
    data := cv.TemplateData{
        TemplateID: "vantage",
        Data: map[string]interface{}{
            "contacts": map[string]interface{}{
                "name":  "John Doe",
                "email": "john@example.com",
                "title": "Software Engineer",
            },
            "position": "Senior Developer",
            "tagline":  "Experienced software engineer with 8+ years...",
            "jobs": []interface{}{
                map[string]interface{}{
                    "position": "Senior Developer",
                    "company": map[string]interface{}{
                        "name": "TechCorp",
                    },
                    "from": "2021 Jan.",
                    "to": "present",
                    "description": []interface{}{
                        "Led development of microservices",
                        "Reduced system latency by 40%",
                    },
                },
            },
            "skills": []interface{}{"Go", "JavaScript", "React"},
        },
    }
    
    // Generate CV
    request := cv.GenerationRequest{
        TemplateID: "vantage", // or "grotesk"
        Data:       data,
        Format:     "pdf",
    }
    
    result, err := service.GenerateCV(request)
    if err != nil {
        panic(err)
    }
    
    // Save PDF
    os.WriteFile("my-cv.pdf", result.Data, 0644)
}
```

### Development Setup

#### 1. Start PostgreSQL Test Instance

```bash
docker run --name mycv-postgres -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres:latest
```

#### 2. Set Environment Variables

```bash
export DB_DSN="postgres:password@localhost:5432/postgres?sslmode=disable"
```

#### 3. Available Commands

View all available tasks:
```bash
go-task --list
```

**Development:**
- `go-task build` - Build the application (automatically generates templ files)
- `go-task run` - Build and run the application
- `go-task dev` - Run with templ watching and live reload
- `go-task test` - Run all tests
- `go-task clean` - Clean build artifacts and generated templ files

**CV Generation Testing:**
```bash
# Run CV generation tests
go test ./internal/cv -v

# Start development API server
go run cmd/test-server/main.go
```

**Development Server API:**
```bash
# List templates
curl http://localhost:8080/api/templates

# Generate sample CV
curl -o sample.pdf http://localhost:8080/api/sample/vantage

# Get template form structure
curl http://localhost:8080/api/form/vantage
```

## 📋 API Reference

### Core Service Methods

#### `NewService(templatesDir, outputDir string) (*Service, error)`
Creates a new CV generation service.

#### `ListTemplates() ([]Template, error)`
Returns all available templates with metadata.

#### `GenerateCV(request GenerationRequest) (*GenerationResult, error)`
Generates a CV PDF from template and data.

#### `ValidateData(templateID string, data TemplateData) (ValidationResult, error)`
Validates data against template requirements.

#### `GenerateForm(templateID string) (*TemplateForm, error)`
Creates a web form structure for a template.

### REST API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/templates` | List all templates |
| GET | `/api/templates/{id}` | Get template details |
| GET | `/api/form/{id}` | Get template form structure |
| GET | `/api/sample/{id}` | Generate CV with sample data |
| POST | `/api/generate/{id}` | Generate CV with custom data |

### Universal Data Structure

All templates work with this standardized data format:

```yaml
contacts:
  name: "John Doe"
  title: "Software Engineer"
  email: "john@example.com"
  address: "San Francisco, CA"
  linkedin:
    url: "https://linkedin.com/in/johndoe"
    displayText: "johndoe"
  github:
    url: "https://github.com/johndoe"
    displayText: "@johndoe"
  website:
    url: "https://johndoe.dev"
    displayText: "johndoe.dev"

position: "Senior Developer"
tagline: "Professional summary highlighting key achievements..."

jobs:
  - position: "Senior Developer"
    company:
      name: "TechCorp"
      link: "https://techcorp.com"
    product:
      name: "CloudScale Platform"
      link: "https://cloudscale.com"
    from: "2021 Jan."
    to: "present"
    location: "San Francisco, CA"
    description:
      - "Led development of microservices architecture"
      - "Reduced system latency by 40%"
    tags: ["Go", "Kubernetes", "PostgreSQL"]

education:
  - place:
      name: "University of California, Berkeley"
      link: "https://berkeley.edu"
    degree: "B.S."
    major: "Computer Science"
    track: "Software Engineering"
    from: "2015"
    to: "2019"
    location: "Berkeley, CA"

technical_expertise:
  - name: "Go"
    level: 5
  - name: "JavaScript" 
    level: 4

skills: ["Go", "JavaScript", "React", "Node.js", "Docker"]
methodology: ["Agile", "TDD", "Code Review"]
tools: ["VS Code", "Git", "Docker", "Kubernetes"]

achievements:
  - name: "AWS Certified Solutions Architect"
    description: "Professional certification demonstrating cloud expertise"
```

## 📚 Available Templates

### Vantage Template
- **Style**: ATS-friendly, professional design
- **Layout**: Two-column with clean typography
- **Features**: Icon support, skill level indicators, project links
- **Best For**: Corporate environments, traditional applications

### Grotesk Template
- **Style**: Modern typography, clean aesthetics  
- **Layout**: Flexible two-column design with FontAwesome icons
- **Features**: Contemporary styling, responsive layout
- **Best For**: Creative industries, startups, modern companies

### Adding New Templates

1. Create template directory: `assets/templates/typst/template-name/`
2. Add Typst template file: `template.typ` 
3. Create configuration: `config.yaml`
4. Test with existing data structure
5. Add template adapter if data conversion needed

The system automatically discovers and validates new templates.

## 🧪 Testing

### Test Commands

```bash
# Run CV generation tests
go test ./internal/cv -v

# Run all tests  
go test ./...

# Test with coverage
go-task test:cover

# Start development server for manual testing
go run cmd/test-server/main.go
```

### Test Coverage

- ✅ Template discovery and parsing
- ✅ Data validation for all templates
- ✅ CV generation with sample data  
- ✅ Template adapter system
- ✅ Form generation and field extraction
- ✅ Error handling and edge cases
- ✅ Multi-template compatibility validation

## 🛠️ Development

### Key Design Principles

1. **Universal Compatibility**: Same data works with all templates
2. **Zero Template Modification**: Use official templates as-is
3. **Dynamic Discovery**: No hardcoded template lists
4. **Type Safety**: Strong typing with comprehensive validation
5. **Extensible Architecture**: Easy to add new features

### Database Migrations

- `go-task migrations:new NAME=migration_name` - Create new migration
- `go-task migrations:up` - Apply all migrations
- `go-task migrations:down` - Rollback all migrations
- `go-task migrations:version` - Show current migration version

### Quality Control

- `go-task audit` - Run quality control checks
- `go-task tidy` - Format code and tidy modules

## 🎉 Success Metrics

### ✅ Completed Features
- [x] Dynamic template system with automatic discovery
- [x] Universal data format working with multiple templates
- [x] Template adapter system for format conversion
- [x] Comprehensive validation and form generation
- [x] Professional PDF generation via Typst
- [x] Complete REST API with development server
- [x] Extensive testing infrastructure
- [x] Two working templates (vantage, grotesk)

### 🎯 Key Achievements
- **Zero Template Modification**: Official templates work as-is
- **Universal Data**: Same CV data generates different designs
- **Production Ready**: Professional quality PDF output
- **Developer Friendly**: Comprehensive APIs and testing
- **Extensible**: Easy to add new templates and features

## 🚀 Next Steps

1. **Web Interface**: Complete frontend for CV creation
2. **More Templates**: Add additional design options
3. **Template Preview**: Generate thumbnail previews
4. **User Management**: Authentication and CV storage
5. **Export Options**: Multiple format support (PNG, HTML)
6. **Template Marketplace**: Community template sharing

## 🤝 Contributing

1. Fork the repository
2. Create feature branch: `git checkout -b feature/new-template`
3. Add template in `assets/templates/typst/`
4. Test with existing data: `go test ./internal/cv -v`
5. Submit pull request with template documentation

## 📄 Technology Stack

- **Backend:** Go with standard library HTTP router
- **Templates:** [Templ](https://templ.guide/) for type-safe web templates
- **CV Generation:** [Typst](https://typst.app/) for professional PDF output
- **Database:** PostgreSQL with migrations
- **Email:** SMTP with templ-based email templates
- **Build:** Task runner for development workflow

## 📄 License

[License details here]

---

**🎉 The CV generation system is complete and production-ready!** 

This system successfully demonstrates that multiple CV templates can work with the same data structure through intelligent adaptation, making it easy for users to try different designs without recreating their CV data.