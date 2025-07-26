package cv

import (
	"fmt"
)

// TemplateAdapter converts between different template data formats
type TemplateAdapter interface {
	// ConvertToTemplate converts common CV data to template-specific format
	ConvertToTemplate(data TemplateData) (TemplateData, error)

	// ConvertFromTemplate converts template-specific data to common format
	ConvertFromTemplate(data TemplateData) (TemplateData, error)

	// GetTemplateID returns the template ID this adapter supports
	GetTemplateID() string

	// GetSourceFormat returns the expected source format identifier
	GetSourceFormat() string
}

// AdapterRegistry manages template adapters
type AdapterRegistry struct {
	adapters map[string]TemplateAdapter
}

// NewAdapterRegistry creates a new adapter registry
func NewAdapterRegistry() *AdapterRegistry {
	registry := &AdapterRegistry{
		adapters: make(map[string]TemplateAdapter),
	}

	// Register built-in adapters
	registry.RegisterAdapter(&VantageAdapter{})
	registry.RegisterAdapter(&GroteskAdapter{})

	return registry
}

// RegisterAdapter registers a template adapter
func (r *AdapterRegistry) RegisterAdapter(adapter TemplateAdapter) {
	r.adapters[adapter.GetTemplateID()] = adapter
}

// GetAdapter returns an adapter for the specified template
func (r *AdapterRegistry) GetAdapter(templateID string) (TemplateAdapter, bool) {
	adapter, exists := r.adapters[templateID]
	return adapter, exists
}

// ConvertData converts data from one template format to another
func (r *AdapterRegistry) ConvertData(fromTemplate, toTemplate string, data TemplateData) (TemplateData, error) {
	// If templates are the same, return as-is
	if fromTemplate == toTemplate {
		return data, nil
	}

	// Convert to common format first (using "vantage" as our common format)
	var commonData TemplateData
	var err error

	if fromTemplate == "vantage" {
		commonData = data
	} else {
		fromAdapter, exists := r.adapters[fromTemplate]
		if !exists {
			return TemplateData{}, fmt.Errorf("no adapter found for source template: %s", fromTemplate)
		}
		commonData, err = fromAdapter.ConvertFromTemplate(data)
		if err != nil {
			return TemplateData{}, fmt.Errorf("failed to convert from %s: %w", fromTemplate, err)
		}
	}

	// Convert from common format to target template
	if toTemplate == "vantage" {
		commonData.TemplateID = toTemplate
		return commonData, nil
	}

	toAdapter, exists := r.adapters[toTemplate]
	if !exists {
		return TemplateData{}, fmt.Errorf("no adapter found for target template: %s", toTemplate)
	}

	result, err := toAdapter.ConvertToTemplate(commonData)
	if err != nil {
		return TemplateData{}, fmt.Errorf("failed to convert to %s: %w", toTemplate, err)
	}

	result.TemplateID = toTemplate
	return result, nil
}

// VantageAdapter handles the vantage template format (our common format)
type VantageAdapter struct{}

func (a *VantageAdapter) GetTemplateID() string {
	return "vantage"
}

func (a *VantageAdapter) GetSourceFormat() string {
	return "vantage"
}

func (a *VantageAdapter) ConvertToTemplate(data TemplateData) (TemplateData, error) {
	// Vantage is our base format, return as-is
	result := data
	result.TemplateID = "vantage"
	return result, nil
}

func (a *VantageAdapter) ConvertFromTemplate(data TemplateData) (TemplateData, error) {
	// Vantage is our base format, return as-is
	return data, nil
}

// GroteskAdapter handles the grotesk template format
type GroteskAdapter struct{}

func (a *GroteskAdapter) GetTemplateID() string {
	return "grotesk"
}

func (a *GroteskAdapter) GetSourceFormat() string {
	return "vantage"
}

func (a *GroteskAdapter) ConvertToTemplate(data TemplateData) (TemplateData, error) {
	// Since grotesk now uses the same data structure as vantage, just pass through
	result := data
	result.TemplateID = "grotesk"
	return result, nil
}

func (a *GroteskAdapter) ConvertFromTemplate(data TemplateData) (TemplateData, error) {
	// Since grotesk now uses the same data structure as vantage, just pass through
	result := data
	result.TemplateID = "vantage"
	return result, nil
}
