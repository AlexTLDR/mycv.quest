package cv

import (
	"fmt"
	"os"
)

// Service provides CV generation functionality
type Service struct {
	generator *Generator
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

	return &Service{
		generator: generator,
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

// GenerateCV generates a CV PDF
func (s *Service) GenerateCV(request GenerationRequest) (*GenerationResult, error) {
	return s.generator.GenerateCV(request)
}

// ValidateData validates CV data against template requirements
func (s *Service) ValidateData(templateID string, data CVData) error {
	template, err := s.GetTemplate(templateID)
	if err != nil {
		return err
	}

	// Basic validation - can be extended based on template.Fields
	if data.Contacts.Name == "" {
		return fmt.Errorf("name is required")
	}

	if data.Contacts.Email == "" {
		return fmt.Errorf("email is required")
	}

	if data.Position == "" {
		return fmt.Errorf("position is required")
	}

	// Validate technical expertise levels
	for _, skill := range data.TechnicalExpertise {
		if skill.Level < 1 || skill.Level > 5 {
			return fmt.Errorf("technical skill level must be between 1 and 5")
		}
	}

	_ = template // Use template for more sophisticated validation if needed

	return nil
}

// GetTemplateFields returns the field structure for a template
func (s *Service) GetTemplateFields(templateID string) (map[string]interface{}, error) {
	template, err := s.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	return template.Fields, nil
}
