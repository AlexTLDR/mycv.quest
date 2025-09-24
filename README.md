<p align="center">
  <img src="./assets/logo.svg" alt="mycv.quest Logo" width="120" height="120">
</p>

# mycv.quest

A modern CV generator that transforms your data into beautiful, professional CVs using Typst templates.

## 🚀 Overview

This is a CV generator application that allows users to create professional CVs using various Typst templates. The application is built with modern Go web technologies and provides a fast, responsive experience for CV creation and generation.

## ✨ Features

- **Multiple CV Templates**: Choose from various professional Typst templates
- **Real-time Generation**: CVs are generated instantly in memory
- **No Data Persistence**: All data is processed in memory for privacy
- **Modern UI**: Clean, responsive interface for easy CV creation
- **PDF Export**: Generate high-quality PDF outputs

## 🛠️ Technology Stack

- **Backend**: Go
- **Templating**: [templ](https://templ.guide/) - Type-safe HTML templates for Go
- **CSS Framework**: [DaisyUI](https://daisyui.com/) - Semantic component classes for Tailwind CSS
- **CSS Build**: [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS framework
- **CV Engine**: [Typst](https://typst.app/) - Modern typesetting system
- **Frontend Build**: Node.js and npm for CSS compilation
- **Build Tool**: [Task](https://taskfile.dev/) - Task runner and build tool
- **Containerization**: Docker with multi-stage builds

## 📋 Prerequisites

- Go 1.21 or later
- [Task](https://taskfile.dev/installation/) (recommended) or standard Go toolchain
- Node.js 16+ and npm (for Tailwind CSS and DaisyUI compilation)
- Docker and Docker Compose (for containerized deployment)
- [templ CLI](https://templ.guide/quick-start/installation) for template generation
- [Typst](https://typst.app/) for CV compilation

## 🚀 Quick Start

### Local Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/AlexTLDR/mycv.quest.git
   cd mycv.quest
   ```

2. **Setup project and install dependencies**
   ```bash
   # Using Task (recommended)
   task setup

   # Or manually
   go mod download
   npm install
   ```

3. **Install templ CLI** (if not already installed)
   ```bash
   go install github.com/a-h/templ/cmd/templ@latest
   ```

4. **Build CSS and generate templates**
   ```bash
   # Using Task (recommended)
   task css
   task templ

   # Or manually
   npm run build-css-prod
   templ generate
   ```

5. **Build and run**
   ```bash
   # Using Task (recommended)
   task run

   # Or using Go directly
   go build -o bin/mycv-quest ./cmd/server
   ./bin/mycv-quest
   ```

6. **Visit the site**
   Open your browser to `http://localhost:8080`

## 📁 Project Structure

```
├── bin/                    # Compiled binaries (gitignored)
├── templates/             # templ template files
├── static/               # Static assets (CSS, JS, images)
│   └── css/              # Generated CSS files
├── src/                  # Source CSS files
│   └── input.css         # Tailwind CSS input file
├── cv-templates/         # Typst CV templates
├── handlers/             # HTTP handlers
├── models/               # Data models
├── services/             # Business logic
├── cmd/                  # Application entry points
│   └── server/           # Main server application
├── internal/             # Private application code
├── node_modules/         # Node.js dependencies (gitignored)
├── package.json          # Node.js dependencies
├── tailwind.config.js    # Tailwind CSS configuration
├── docker-compose.yml    # Docker compose configuration
├── Dockerfile           # Multi-stage Docker build
├── Taskfile.yml         # Task automation
├── go.mod               # Go module definition
└── README.md            # This file
```

## 🎨 Available CV Templates

Currently available Typst CV templates:

- **Basic Resume**: Clean and simple design ([Source](https://github.com/stuxf/basic-typst-resume-template))
- **Vantage**: Modern professional template ([Source](https://github.com/sardorml/vantage-typst))
- **Modern Resume**: Contemporary design ([Source](https://github.com/peterpf/modern-typst-resume))

### 🆕 Propose New Templates

Want to see more CV templates? Email me at **alex@alextldr.com** to propose new CV models from the [Typst Universe](https://typst.app/universe/). I'm always looking to expand our template collection!

## 🔧 Available Tasks

This project uses [Task](https://taskfile.dev/) for build automation:

```bash
# Setup project and install dependencies
task setup

# Install dependencies
task install

# Build CSS with Tailwind and DaisyUI
task css

# Build CSS in watch mode
task css-watch

# Generate templ templates
task templ

# Generate templ templates in watch mode
task templ-watch

# Build the application
task build

# Build and run the application
task run

# Run in development mode with hot reload
task dev

# Docker commands
task docker-build
task docker-run
task docker-compose-up
task docker-compose-dev
task docker-compose-down

# Build for production
task prod

# Clean build artifacts
task clean

# Run linters
task lint

# Run tests
task test

# Show all available tasks
task help
```

## 🌐 Deployment

### Docker Deployment

The application is containerized using Docker with a multi-stage build:

```bash
# Build the Docker image
docker build -t mycv-quest .

# Run the container
docker run -p 8080:8080 mycv-quest
```

## 🛣️ Roadmap

- **User Authentication System**: Allow users to create accounts and save their CVs
- **Template Customization**: Enable users to modify template colors, fonts, and layouts
- **Template Builder**: Add new fields or modify existing templates (e.g., add a hobby field to any template)
- **Multiple Export Formats**: Support for Word, LaTeX, and HTML exports

## 🔍 Development Workflow

1. **Make changes** to your Go code, templ templates, or CSS
2. **Rebuild CSS** if you modified CSS or templates:
   ```bash
   task css
   ```
3. **Generate templates** if you modified .templ files:
   ```bash
   task templ
   ```
4. **Build and run** the application:
   ```bash
   task run
   ```
5. **Test your changes** at `http://localhost:8080`

### Hot Reload Development:
```bash
# Run with auto-rebuild (CSS + Templates + Go)
task dev
```

## 🤝 Contributing

We welcome contributions! Whether you're:
- Adding new Typst CV templates
- Improving the user interface
- Fixing bugs or adding features
- Improving documentation

Please feel free to:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

For template contributions, please ensure they follow Typst best practices and include proper documentation.

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- **Typst Universe**: [typst.app/universe](https://typst.app/universe/) - Source for CV templates
- **templ Documentation**: [templ.guide](https://templ.guide/)
- **DaisyUI Documentation**: [daisyui.com](https://daisyui.com/)
- **Typst Documentation**: [typst.app/docs](https://typst.app/docs/)

## 📧 Contact

Alex - **alex@alextldr.com** - [GitHub](https://github.com/AlexTLDR)

For template suggestions, feature requests, or general inquiries, don't hesitate to reach out!

---

**Note**: All CVs are generated in memory with no persistent data storage, ensuring your privacy and security.
