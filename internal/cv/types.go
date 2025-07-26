package cv

import (
	"time"
)

// TemplateData represents CV data in a generic, flexible format
type TemplateData struct {
	TemplateID string                 `json:"template_id" yaml:"template_id"`
	Data       map[string]interface{} `json:"data" yaml:"data"`
}

// Template represents a CV template with dynamic field definitions
type Template struct {
	ID            string                     `json:"id"`
	Name          string                     `json:"name"`
	Description   string                     `json:"description"`
	Version       string                     `json:"version"`
	Author        string                     `json:"author"`
	Features      []string                   `json:"features,omitempty"`
	Fields        map[string]FieldDefinition `json:"fields"`
	TemplateFiles []string                   `json:"template_files,omitempty"`
}

// FieldDefinition defines the structure and validation rules for a field
type FieldDefinition struct {
	Type        string                     `json:"type" yaml:"type"` // string, text, integer, boolean, array, object
	Required    bool                       `json:"required,omitempty" yaml:"required,omitempty"`
	Label       string                     `json:"label,omitempty" yaml:"label,omitempty"`
	Description string                     `json:"description,omitempty" yaml:"description,omitempty"`
	Default     interface{}                `json:"default,omitempty" yaml:"default,omitempty"`
	Options     []string                   `json:"options,omitempty" yaml:"options,omitempty"` // for enum-like fields
	Min         *int                       `json:"min,omitempty" yaml:"min,omitempty"`         // for integer validation
	Max         *int                       `json:"max,omitempty" yaml:"max,omitempty"`         // for integer validation
	Pattern     string                     `json:"pattern,omitempty" yaml:"pattern,omitempty"` // regex pattern for string validation
	Fields      map[string]FieldDefinition `json:"fields,omitempty" yaml:"fields,omitempty"`   // for nested objects
	Items       *FieldDefinition           `json:"items,omitempty" yaml:"items,omitempty"`     // for array items
}

// GenerationRequest represents a CV generation request
type GenerationRequest struct {
	TemplateID string       `json:"template_id"`
	Data       TemplateData `json:"data"`
	Format     string       `json:"format"` // "pdf", "png", etc.
}

// GenerationResult represents the result of CV generation
type GenerationResult struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message,omitempty"`
	Filename  string    `json:"filename,omitempty"`
	Data      []byte    `json:"data,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationResult represents the result of data validation
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// FieldValue represents a field value with metadata
type FieldValue struct {
	Value  interface{}     `json:"value"`
	Field  FieldDefinition `json:"field"`
	Path   string          `json:"path"`
	Errors []string        `json:"errors,omitempty"`
}

// FormSection represents a section in a dynamic form
type FormSection struct {
	ID          string                     `json:"id"`
	Title       string                     `json:"title"`
	Description string                     `json:"description,omitempty"`
	Fields      map[string]FieldDefinition `json:"fields"`
	Order       []string                   `json:"order,omitempty"` // field display order
}

// TemplateForm represents the complete form structure for a template
type TemplateForm struct {
	TemplateID string                 `json:"template_id"`
	Title      string                 `json:"title"`
	Sections   map[string]FormSection `json:"sections"`
	Order      []string               `json:"order"` // section display order
}

// DataExtractor defines how to extract a name or other identifier from template data
type DataExtractor struct {
	NamePath     []string `json:"name_path" yaml:"name_path"`         // path to name field (e.g., ["personal", "info", "name"])
	EmailPath    []string `json:"email_path" yaml:"email_path"`       // path to email field
	FallbackName string   `json:"fallback_name" yaml:"fallback_name"` // fallback if name not found
}

// TemplateConfig represents the complete template configuration
type TemplateConfig struct {
	Template
	DataExtractor DataExtractor `json:"data_extractor,omitempty" yaml:"data_extractor,omitempty"`
	MainFunction  string        `json:"main_function,omitempty" yaml:"main_function,omitempty"` // e.g., "vantage-cv", "grotesk-cv"
}

// PresetData represents preset/example data for a template
type PresetData struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Data        TemplateData `json:"data"`
}

// TemplateMetadata contains additional template information
type TemplateMetadata struct {
	TemplateID     string       `json:"template_id"`
	LastModified   time.Time    `json:"last_modified"`
	SampleImageURL string       `json:"sample_image_url,omitempty"`
	Tags           []string     `json:"tags,omitempty"`
	Presets        []PresetData `json:"presets,omitempty"`
}

// ConversionRule defines how to convert data between template formats
type ConversionRule struct {
	SourceTemplate string                 `json:"source_template"`
	TargetTemplate string                 `json:"target_template"`
	FieldMappings  map[string]string      `json:"field_mappings"`       // source_path -> target_path
	Transforms     map[string]interface{} `json:"transforms,omitempty"` // custom transformation rules
}

// TemplateCompatibility represents compatibility information between templates
type TemplateCompatibility struct {
	FromTemplate    string           `json:"from_template"`
	ToTemplate      string           `json:"to_template"`
	Compatibility   float64          `json:"compatibility"` // 0.0 to 1.0
	ConversionRules []ConversionRule `json:"conversion_rules,omitempty"`
}
