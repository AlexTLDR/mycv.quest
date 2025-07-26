package cv

import (
	"fmt"
	"os"
	"path/filepath"
)

// Service provides CV generation functionality with dynamic template support
type Service struct {
	generator *Generator
	parser    *TemplateParser
}

// NewService creates a new CV service
func NewService(templatesDir, outputDir string) (*Service, error) {
	// Ensure directories exist
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create templates directory: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	generator := NewGenerator(templatesDir, outputDir)
	parser := NewTemplateParser(templatesDir)

	return &Service{
		generator: generator,
		parser:    parser,
	}, nil
}

// ListTemplates returns all available templates
func (s *Service) ListTemplates() ([]Template, error) {
	return s.generator.GetAvailableTemplates()
}

// GetTemplate returns a specific template
func (s *Service) GetTemplate(templateID string) (*Template, error) {
	return s.generator.GetTemplate(templateID)
}

// GetTemplateConfig returns the full template configuration
func (s *Service) GetTemplateConfig(templateID string) (*TemplateConfig, error) {
	return s.generator.GetTemplateConfig(templateID)
}

// GenerateForm creates a dynamic form structure for a template
func (s *Service) GenerateForm(templateID string) (*TemplateForm, error) {
	return s.generator.GenerateForm(templateID)
}

// ValidateData validates template data against its configuration
func (s *Service) ValidateData(templateID string, data TemplateData) (ValidationResult, error) {
	return s.generator.ValidateData(templateID, data)
}

// GenerateCV generates a CV PDF
func (s *Service) GenerateCV(request GenerationRequest) (*GenerationResult, error) {
	return s.generator.GenerateCV(request)
}

// GeneratePreview generates a preview of the CV
func (s *Service) GeneratePreview(templateID string, data TemplateData) (*GenerationResult, error) {
	return s.generator.GeneratePreview(templateID, data)
}

// GetTemplateMetadata returns metadata for a template
func (s *Service) GetTemplateMetadata(templateID string) (*TemplateMetadata, error) {
	return s.generator.GetTemplateMetadata(templateID)
}

// GetSampleData generates sample data for a template
func (s *Service) GetSampleData(templateID string) (*TemplateData, error) {
	return s.generator.GetSampleData(templateID)
}

// ListTemplateFiles returns files in a template directory
func (s *Service) ListTemplateFiles(templateID string) ([]string, error) {
	return s.generator.ListTemplateFiles(templateID)
}

// ExtractDisplayName extracts a display name from template data
func (s *Service) ExtractDisplayName(templateID string, data TemplateData) (string, error) {
	config, err := s.parser.ParseTemplate(templateID)
	if err != nil {
		return "", err
	}

	return s.parser.GetDisplayName(config, data), nil
}

// ConvertFieldValue converts a string value to the appropriate type
func (s *Service) ConvertFieldValue(templateID, fieldName, value string) (interface{}, error) {
	config, err := s.parser.ParseTemplate(templateID)
	if err != nil {
		return nil, err
	}

	fieldDef, exists := config.Fields[fieldName]
	if !exists {
		return value, nil // Return as-is if field not found
	}

	return s.parser.ConvertValue(value, fieldDef)
}

// GetFieldDefinition returns the definition for a specific field
func (s *Service) GetFieldDefinition(templateID, fieldName string) (*FieldDefinition, error) {
	config, err := s.parser.ParseTemplate(templateID)
	if err != nil {
		return nil, err
	}

	if fieldDef, exists := config.Fields[fieldName]; exists {
		return &fieldDef, nil
	}

	return nil, fmt.Errorf("field %s not found in template %s", fieldName, templateID)
}

// ValidateField validates a single field value
func (s *Service) ValidateField(templateID, fieldName string, value interface{}) []string {
	config, err := s.parser.ParseTemplate(templateID)
	if err != nil {
		return []string{fmt.Sprintf("Template error: %v", err)}
	}

	fieldDef, exists := config.Fields[fieldName]
	if !exists {
		return []string{"Field not found in template"}
	}

	data := map[string]interface{}{fieldName: value}
	return s.parser.validateFieldValue(data, fieldName, fieldDef, fieldName)
}

// CreateTemplateData creates a TemplateData structure from raw data
func (s *Service) CreateTemplateData(templateID string, data map[string]interface{}) TemplateData {
	return TemplateData{
		TemplateID: templateID,
		Data:       data,
	}
}

// ExtractFieldValue extracts a nested field value from template data
func (s *Service) ExtractFieldValue(data TemplateData, path []string) (interface{}, bool) {
	return s.parser.ExtractValue(data.Data, path)
}

// SetFieldValue sets a nested field value in template data
func (s *Service) SetFieldValue(data *TemplateData, path []string, value interface{}) error {
	if len(path) == 0 {
		return fmt.Errorf("empty path")
	}

	current := data.Data

	// Navigate to the parent of the target field
	for i := 0; i < len(path)-1; i++ {
		key := path[i]

		if next, exists := current[key]; exists {
			if nextMap, ok := next.(map[string]interface{}); ok {
				current = nextMap
			} else {
				// Need to replace with map
				newMap := make(map[string]interface{})
				current[key] = newMap
				current = newMap
			}
		} else {
			// Create new map
			newMap := make(map[string]interface{})
			current[key] = newMap
			current = newMap
		}
	}

	// Set the final value
	current[path[len(path)-1]] = value
	return nil
}

// GetTemplateVersion returns the version of a template
func (s *Service) GetTemplateVersion(templateID string) (string, error) {
	config, err := s.parser.ParseTemplate(templateID)
	if err != nil {
		return "", err
	}

	return config.Version, nil
}

// IsTemplateAvailable checks if a template is available
func (s *Service) IsTemplateAvailable(templateID string) bool {
	templateDir := filepath.Join(s.generator.templatesDir, templateID)
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		return false
	}

	// Check if we can parse the template
	_, err := s.parser.ParseTemplate(templateID)
	return err == nil
}

// GetCompatibleTemplates returns templates compatible with the given template
func (s *Service) GetCompatibleTemplates(templateID string) ([]Template, error) {
	// For now, return all templates as potentially compatible
	// In the future, this could implement actual compatibility checking
	// based on field structure similarity
	return s.ListTemplates()
}

// CloneTemplateData creates a deep copy of template data
func (s *Service) CloneTemplateData(data TemplateData) TemplateData {
	cloned := TemplateData{
		TemplateID: data.TemplateID,
		Data:       make(map[string]interface{}),
	}

	// Deep copy the data map
	for key, value := range data.Data {
		cloned.Data[key] = s.cloneValue(value)
	}

	return cloned
}

// cloneValue recursively clones a value
func (s *Service) cloneValue(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		cloned := make(map[string]interface{})
		for key, val := range v {
			cloned[key] = s.cloneValue(val)
		}
		return cloned
	case []interface{}:
		cloned := make([]interface{}, len(v))
		for i, val := range v {
			cloned[i] = s.cloneValue(val)
		}
		return cloned
	case []string:
		cloned := make([]string, len(v))
		copy(cloned, v)
		return cloned
	default:
		return value
	}
}

// GetRequiredFields returns all required fields for a template
func (s *Service) GetRequiredFields(templateID string) ([]string, error) {
	config, err := s.parser.ParseTemplate(templateID)
	if err != nil {
		return nil, err
	}

	var required []string
	for fieldName, fieldDef := range config.Fields {
		if fieldDef.Required {
			required = append(required, fieldName)
		}

		// Check nested required fields
		if fieldDef.Type == "object" && fieldDef.Fields != nil {
			for nestedName, nestedDef := range fieldDef.Fields {
				if nestedDef.Required {
					required = append(required, fieldName+"."+nestedName)
				}
			}
		}
	}

	return required, nil
}

// GetFieldsOfType returns all fields of a specific type
func (s *Service) GetFieldsOfType(templateID, fieldType string) ([]string, error) {
	config, err := s.parser.ParseTemplate(templateID)
	if err != nil {
		return nil, err
	}

	var fields []string
	for fieldName, fieldDef := range config.Fields {
		if fieldDef.Type == fieldType {
			fields = append(fields, fieldName)
		}
	}

	return fields, nil
}
