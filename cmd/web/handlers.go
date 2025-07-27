package main

import (
	"fmt"
	"net/http"

	templates "github.com/AlexTLDR/mycv.quest/assets/templates/templ"
	"github.com/AlexTLDR/mycv.quest/internal/response"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	err := response.Component(w, http.StatusOK, templates.HomePage(data))
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) protected(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is a protected handler"))
}

func (app *application) cvBuilder(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	// Get available templates dynamically
	availableTemplates, err := app.getAvailableTemplates()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Convert to template struct for the template
	var templateData []templates.Template
	for _, t := range availableTemplates {
		templateData = append(templateData, templates.Template{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			Version:     t.Version,
			Author:      t.Author,
		})
	}

	err = response.Component(w, http.StatusOK, templates.CVBuilderPage(data, templateData))
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) templates(w http.ResponseWriter, r *http.Request) {
	// Get available templates dynamically
	availableTemplates, err := app.getAvailableTemplates()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Simple HTML response without using Base template to avoid errors
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	html := `<!DOCTYPE html>
<html lang="en" class="dark">
<head>
	<meta charset="utf-8">
	<title>Templates - MyCV.Quest</title>
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<link rel="stylesheet" href="/static/css/main.css">
	<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body class="min-h-screen bg-background text-foreground">
	<header class="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
		<div class="container mx-auto flex h-14 items-center justify-between px-4">
			<div class="flex items-center">
				<h1 class="mr-6 text-lg font-semibold">
					<a href="/" class="hover:text-primary transition-colors">MyCV.Quest</a>
				</h1>
				<nav class="flex items-center space-x-6 text-sm font-medium">
					<a href="/" class="transition-colors hover:text-foreground/80 text-foreground/60">Home</a>
					<a href="/cv-builder" class="transition-colors hover:text-foreground/80 text-foreground/60">CV Builder</a>
					<a href="/templates" class="transition-colors hover:text-foreground/80 text-primary">Templates</a>
				</nav>
			</div>
			<button
				onclick="document.documentElement.classList.toggle('dark')"
				class="p-2 rounded-md hover:bg-accent transition-colors"
			>
				<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"></path>
				</svg>
			</button>
		</div>
	</header>
	<main class="flex-1">
		<div class="container mx-auto p-6 max-w-6xl">
			<div class="mb-12 text-center">
				<h1 class="text-4xl font-bold bg-gradient-to-r from-primary to-violet-600 bg-clip-text text-transparent mb-4">CV Templates</h1>
				<p class="text-xl text-muted-foreground max-w-2xl mx-auto">Choose from our collection of professional CV templates designed to make you stand out</p>
			</div>
			<div class="grid grid-cols-1 md:grid-cols-2 gap-8 mb-12">`

	for _, template := range availableTemplates {
		html += fmt.Sprintf(`
			<div class="group hover:shadow-xl transition-all duration-300 border-2 hover:border-primary/20 rounded-lg border-border bg-card text-card-foreground shadow-sm">
				<div class="p-6">
					<div class="bg-gradient-to-br from-muted/50 to-muted/30 rounded-xl p-8 mb-6 min-h-[280px] flex items-center justify-center border border-border/50">
						<div class="text-center">
							<div class="w-20 h-24 mx-auto mb-4 bg-primary/10 rounded-lg border-2 border-dashed border-primary/30 flex items-center justify-center">
								<svg class="w-8 h-8 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
								</svg>
							</div>
							<p class="text-sm font-medium text-foreground">Preview Available Soon</p>
							<p class="text-xs text-muted-foreground mt-1">%s.typ</p>
						</div>
					</div>
					<div class="mb-6">
						<h3 class="text-2xl font-bold text-primary mb-2">%s</h3>
						<p class="text-muted-foreground mb-4 leading-relaxed">%s</p>
						<div class="flex items-center justify-between text-sm">
							<div class="space-y-1">
								<div class="flex items-center gap-2">
									<span class="text-muted-foreground">Version:</span>
									<span class="font-semibold text-foreground">%s</span>
								</div>
								<div class="flex items-center gap-2">
									<span class="text-muted-foreground">Author:</span>
									<span class="font-semibold text-foreground">%s</span>
								</div>
							</div>
						</div>
					</div>
					<div class="flex gap-3">
						<a href="/cv-builder?template=%s" class="flex-1">
							<button class="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-10 px-4 py-2 w-full text-base py-3">
								<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
								</svg>
								Use This Template
							</button>
						</a>
						<button class="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground px-4 py-3">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
							</svg>
						</button>
					</div>
				</div>
			</div>`,
			template.ID,
			template.Name,
			template.Description,
			template.Version,
			template.Author,
			template.ID)
	}

	if len(availableTemplates) == 0 {
		html += `
			<div class="col-span-full text-center py-16">
				<div class="max-w-md mx-auto rounded-lg border bg-card text-card-foreground shadow-sm">
					<div class="p-8 text-center">
						<div class="w-16 h-16 mx-auto mb-4 bg-muted rounded-full flex items-center justify-center">
							<svg class="w-8 h-8 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
							</svg>
						</div>
						<h3 class="text-xl font-semibold mb-2">No Templates Found</h3>
						<p class="text-muted-foreground mb-4">Add templates to the assets/templates/typst/ directory</p>
						<p class="text-sm text-muted-foreground">Each template should have a config.yaml and template.typ file</p>
					</div>
				</div>
			</div>`
	}

	html += `
			</div>
			<div class="text-center pt-8 border-t border-border">
				<a href="/cv-builder" class="inline-flex items-center text-lg text-primary hover:text-primary/80 transition-colors font-medium">
					<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"></path>
					</svg>
					Back to CV Builder
				</a>
			</div>
		</div>
	</main>
	<footer class="border-t py-6 md:py-0">
		<div class="container mx-auto flex flex-col items-center justify-between gap-4 md:h-24 md:flex-row px-4">
			<div class="text-center text-sm leading-loose text-muted-foreground md:text-left">
				© 2025 MyCV.Quest. Built with Go and templ.
			</div>
		</div>
	</footer>
</body>
</html>`

	w.Write([]byte(html))
}
