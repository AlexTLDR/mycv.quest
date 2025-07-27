package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/AlexTLDR/mycv.quest/assets"
	templates "github.com/AlexTLDR/mycv.quest/assets/templates/templ"
	"github.com/AlexTLDR/mycv.quest/internal/version"
	"gopkg.in/yaml.v3"
)

type Template struct {
	ID          string `yaml:"-"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
	Author      string `yaml:"author"`
}

func (app *application) getAvailableTemplates() ([]Template, error) {
	var templates []Template

	// Walk through the templates/typst directory in embedded files
	err := fs.WalkDir(assets.EmbeddedFiles, "templates/typst", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Look for config.yaml files
		if d.Name() == "config.yaml" {
			// Extract template ID from path
			templateID := filepath.Base(filepath.Dir(path))

			// Read the config file
			configData, err := fs.ReadFile(assets.EmbeddedFiles, path)
			if err != nil {
				app.logger.Error("Failed to read template config", "path", path, "error", err)
				return nil // Continue walking, don't fail completely
			}

			// Parse the YAML
			var template Template
			if err := yaml.Unmarshal(configData, &template); err != nil {
				app.logger.Error("Failed to parse template config", "path", path, "error", err)
				return nil // Continue walking
			}

			template.ID = templateID
			templates = append(templates, template)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return templates, nil
}

func (app *application) newTemplateData(r *http.Request) templates.PageData {
	data := templates.PageData{
		Version: version.Get(),
	}

	return data
}

func (app *application) backgroundTask(r *http.Request, fn func() error) {
	app.wg.Add(1)

	go func() {
		defer app.wg.Done()

		defer func() {
			pv := recover()
			if pv != nil {
				app.reportServerError(r, fmt.Errorf("%v", pv))
			}
		}()

		err := fn()
		if err != nil {
			app.reportServerError(r, err)
		}
	}()
}
