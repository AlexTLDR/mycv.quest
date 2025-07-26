package cv

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// TemplateParser handles parsing template configurations and field definitions
type TemplateParser struct {
	templatesDir string
}

// NewTemplateParser creates a new template parser
func NewTemplateParser(templatesDir string) *TemplateParser {
	return &TemplateParser{
		templatesDir: templatesDir,
	}
}

// ParseTemplate reads and parses a template configuration
func (p *TemplateParser) ParseTemplate(templateID string) (*TemplateConfig, error) {
	templateDir := filepath.Join(p.templatesDir, templateID)

	// Try to find configuration file (config.yaml, config.toml, info.toml)
	configFiles := []string{"config.yaml", "config.yml", "config.toml", "info.toml"}
	var configPath string
	var configType string

	for _, filename := range configFiles {
		path := filepath.Join(templateDir, filename)
		if _, err := os.Stat(path); err == nil {
			configPath = path
			if strings.HasSuffix(filename, ".toml") {
				configType = "toml"
			} else {
				configType = "yaml"
			}
			break
		}
	}

	if configPath == "" {
		return nil, fmt.Errorf("no configuration file found for template %s", templateID)
	}

	// Read configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config TemplateConfig

	// Parse based on file type
	switch configType {
	case "toml":
		if err := toml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse TOML config: %w", err)
		}
	case "yaml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	}

	config.ID = templateID

	// If no explicit field definitions found, try to infer from template files
	if len(config.Fields) == 0 {
		inferredFields, err := p.inferFieldsFromTemplate(templateDir)
		if err == nil {
			config.Fields = inferredFields
		}
	}

	// Set default data extractor if not specified
	if len(config.DataExtractor.NamePath) == 0 {
		config.DataExtractor = p.inferDataExtractor(config.Fields)
	}

	// Set default main function if not specified
	if config.MainFunction == "" {
		config.MainFunction = templateID + "-cv"
	}

	return &config, nil
}

// inferFieldsFromTemplate attempts to infer field structure from .typ files
func (p *TemplateParser) inferFieldsFromTemplate(templateDir string) (map[string]FieldDefinition, error) {
	fields := make(map[string]FieldDefinition)

	// Look for .typ files
	err := filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(info.Name(), ".typ") {
			typFields, err := p.parseTypstFile(path)
			if err == nil {
				for k, v := range typFields {
					fields[k] = v
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return fields, nil
}

// parseTypstFile extracts field references from Typst files
func (p *TemplateParser) parseTypstFile(filePath string) (map[string]FieldDefinition, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	content := string(data)
	fields := make(map[string]FieldDefinition)

	// Regex patterns to find field references in Typst code
	patterns := []string{
		`data\.([a-zA-Z_][a-zA-Z0-9_]*)`,         // data.fieldname
		`#data\.([a-zA-Z_][a-zA-Z0-9_]*)`,        // #data.fieldname
		`data\.at\("([^"]+)"\)`,                  // data.at("fieldname")
		`if "([^"]+)" in data`,                   // if "fieldname" in data
		`for ([a-zA-Z_][a-zA-Z0-9_]*) in data\.`, // for item in data.something
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content, -1)

		for _, match := range matches {
			if len(match) > 1 {
				fieldName := match[1]
				if fieldName != "" {
					// Infer field type based on usage context
					fieldType := p.inferFieldType(content, fieldName)
					fields[fieldName] = FieldDefinition{
						Type:  fieldType,
						Label: p.generateLabel(fieldName),
					}
				}
			}
		}
	}

	return fields, nil
}

// inferFieldType attempts to determine field type from usage context
func (p *TemplateParser) inferFieldType(content, fieldName string) string {
	// Look for patterns that indicate field type
	patterns := map[string]string{
		`for\s+\w+\s+in\s+data\.` + fieldName + `\[`:  "array",
		`data\.` + fieldName + `\.len\(\)`:            "array",
		`if\s+data\.` + fieldName + `\s*!=\s*""`:      "string",
		`data\.` + fieldName + `\.join\(`:             "array",
		`#text\([^)]*\)\[#data\.` + fieldName + `\]`:  "string",
		`data\.` + fieldName + `\s*==\s*(true|false)`: "boolean",
		`data\.` + fieldName + `\s*[<>=]\s*\d+`:       "integer",
	}

	for pattern, fieldType := range patterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			return fieldType
		}
	}

	// Default to string if cannot determine
	return "string"
}

// generateLabel creates a human-readable label from field name
func (p *TemplateParser) generateLabel(fieldName string) string {
	// Convert snake_case to Title Case
	words := strings.Split(fieldName, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

// inferDataExtractor creates a default data extractor based on field definitions
func (p *TemplateParser) inferDataExtractor(fields map[string]FieldDefinition) DataExtractor {
	extractor := DataExtractor{
		FallbackName: "Unknown",
	}

	// Look for common name field patterns
	namePatterns := [][]string{
		{"personal", "info", "name"},
		{"contacts", "name"},
		{"name"},
		{"personal", "name"},
	}

	for _, pattern := range namePatterns {
		if p.fieldPathExists(fields, pattern) {
			extractor.NamePath = pattern
			break
		}
	}

	// Look for common email field patterns
	emailPatterns := [][]string{
		{"personal", "info", "email"},
		{"contacts", "email"},
		{"email"},
		{"personal", "email"},
	}

	for _, pattern := range emailPatterns {
		if p.fieldPathExists(fields, pattern) {
			extractor.EmailPath = pattern
			break
		}
	}

	return extractor
}

// fieldPathExists checks if a nested field path exists in the field definitions
func (p *TemplateParser) fieldPathExists(fields map[string]FieldDefinition, path []string) bool {
	if len(path) == 0 {
		return false
	}

	current := fields
	for i, key := range path {
		if field, exists := current[key]; exists {
			if i == len(path)-1 {
				return true // Found the complete path
			}
			if field.Type == "object" && field.Fields != nil {
				current = field.Fields
			} else {
				return false // Path broken
			}
		} else {
			return false // Key not found
		}
	}

	return false
}

// GenerateForm creates a form structure from template configuration
func (p *TemplateParser) GenerateForm(config *TemplateConfig) (*TemplateForm, error) {
	form := &TemplateForm{
		TemplateID: config.ID,
		Title:      fmt.Sprintf("%s CV", config.Name),
		Sections:   make(map[string]FormSection),
		Order:      []string{},
	}

	// Group fields into logical sections
	sections := p.groupFieldsIntoSections(config.Fields)

	for sectionID, sectionFields := range sections {
		section := FormSection{
			ID:     sectionID,
			Title:  p.generateLabel(sectionID),
			Fields: sectionFields,
			Order:  p.generateFieldOrder(sectionFields),
		}

		form.Sections[sectionID] = section
		form.Order = append(form.Order, sectionID)
	}

	return form, nil
}

// groupFieldsIntoSections organizes fields into logical sections
func (p *TemplateParser) groupFieldsIntoSections(fields map[string]FieldDefinition) map[string]map[string]FieldDefinition {
	sections := make(map[string]map[string]FieldDefinition)

	for fieldName, fieldDef := range fields {
		sectionName := p.determineSectionForField(fieldName)

		if sections[sectionName] == nil {
			sections[sectionName] = make(map[string]FieldDefinition)
		}

		sections[sectionName][fieldName] = fieldDef
	}

	return sections
}

// determineSectionForField determines which section a field belongs to
func (p *TemplateParser) determineSectionForField(fieldName string) string {
	// Define section mappings based on field names
	sectionMappings := map[string]string{
		"personal":   "personal",
		"contacts":   "personal",
		"name":       "personal",
		"email":      "personal",
		"phone":      "personal",
		"address":    "personal",
		"experience": "experience",
		"jobs":       "experience",
		"work":       "experience",
		"education":  "education",
		"skills":     "skills",
		"technical":  "skills",
		"projects":   "projects",
		"settings":   "settings",
		"layout":     "settings",
	}

	fieldLower := strings.ToLower(fieldName)

	for keyword, section := range sectionMappings {
		if strings.Contains(fieldLower, keyword) {
			return section
		}
	}

	return "other"
}

// generateFieldOrder creates a logical order for fields within a section
func (p *TemplateParser) generateFieldOrder(fields map[string]FieldDefinition) []string {
	// Define priority order for common fields
	priorityOrder := []string{
		"name", "title", "email", "phone", "address",
		"position", "company", "location", "start_date", "end_date",
		"degree", "institution", "graduation_date",
	}

	var ordered []string
	var remaining []string

	// Add priority fields first
	for _, priority := range priorityOrder {
		if _, exists := fields[priority]; exists {
			ordered = append(ordered, priority)
		}
	}

	// Add remaining fields
	for fieldName := range fields {
		found := false
		for _, existing := range ordered {
			if existing == fieldName {
				found = true
				break
			}
		}
		if !found {
			remaining = append(remaining, fieldName)
		}
	}

	return append(ordered, remaining...)
}

// ValidateData validates template data against field definitions
func (p *TemplateParser) ValidateData(config *TemplateConfig, data TemplateData) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	// Validate each field
	for fieldName, fieldDef := range config.Fields {
		errors := p.validateFieldValue(data.Data, fieldName, fieldDef, fieldName)
		for _, err := range errors {
			result.Errors = append(result.Errors, ValidationError{
				Field:   err,
				Message: fmt.Sprintf("Field '%s': %s", fieldName, err),
			})
			result.Valid = false
		}
	}

	return result
}

// validateFieldValue validates a single field value
func (p *TemplateParser) validateFieldValue(data map[string]interface{}, fieldName string, fieldDef FieldDefinition, path string) []string {
	var errors []string

	value, exists := data[fieldName]

	// Check required fields
	if fieldDef.Required && (!exists || value == nil || value == "") {
		errors = append(errors, fmt.Sprintf("required field missing"))
		return errors
	}

	if !exists || value == nil {
		return errors // Optional field not provided
	}

	// Validate by type
	switch fieldDef.Type {
	case "string":
		if str, ok := value.(string); ok {
			if fieldDef.Pattern != "" {
				if matched, _ := regexp.MatchString(fieldDef.Pattern, str); !matched {
					errors = append(errors, fmt.Sprintf("does not match required pattern"))
				}
			}
			if fieldDef.Options != nil && len(fieldDef.Options) > 0 {
				valid := false
				for _, option := range fieldDef.Options {
					if str == option {
						valid = true
						break
					}
				}
				if !valid {
					errors = append(errors, fmt.Sprintf("must be one of: %s", strings.Join(fieldDef.Options, ", ")))
				}
			}
		} else {
			errors = append(errors, "must be a string")
		}

	case "integer":
		if num, ok := value.(int); ok {
			if fieldDef.Min != nil && num < *fieldDef.Min {
				errors = append(errors, fmt.Sprintf("must be at least %d", *fieldDef.Min))
			}
			if fieldDef.Max != nil && num > *fieldDef.Max {
				errors = append(errors, fmt.Sprintf("must be at most %d", *fieldDef.Max))
			}
		} else if numFloat, ok := value.(float64); ok {
			num := int(numFloat)
			if fieldDef.Min != nil && num < *fieldDef.Min {
				errors = append(errors, fmt.Sprintf("must be at least %d", *fieldDef.Min))
			}
			if fieldDef.Max != nil && num > *fieldDef.Max {
				errors = append(errors, fmt.Sprintf("must be at most %d", *fieldDef.Max))
			}
		} else {
			errors = append(errors, "must be an integer")
		}

	case "boolean":
		if _, ok := value.(bool); !ok {
			errors = append(errors, "must be a boolean")
		}

	case "array":
		if arr, ok := value.([]interface{}); ok {
			if fieldDef.Items != nil {
				for i, item := range arr {
					itemPath := fmt.Sprintf("%s[%d]", path, i)
					itemErrors := p.validateFieldValue(map[string]interface{}{"item": item}, "item", *fieldDef.Items, itemPath)
					errors = append(errors, itemErrors...)
				}
			}
		} else {
			errors = append(errors, "must be an array")
		}

	case "object":
		if obj, ok := value.(map[string]interface{}); ok {
			if fieldDef.Fields != nil {
				for subFieldName, subFieldDef := range fieldDef.Fields {
					subPath := fmt.Sprintf("%s.%s", path, subFieldName)
					subErrors := p.validateFieldValue(obj, subFieldName, subFieldDef, subPath)
					errors = append(errors, subErrors...)
				}
			}
		} else {
			errors = append(errors, "must be an object")
		}
	}

	return errors
}

// ExtractValue extracts a value from nested data using a path
func (p *TemplateParser) ExtractValue(data map[string]interface{}, path []string) (interface{}, bool) {
	if len(path) == 0 {
		return nil, false
	}

	current := data
	for i, key := range path {
		if value, exists := current[key]; exists {
			if i == len(path)-1 {
				return value, true
			}
			if nextMap, ok := value.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	}

	return nil, false
}

// GetDisplayName extracts a display name from template data
func (p *TemplateParser) GetDisplayName(config *TemplateConfig, data TemplateData) string {
	if name, found := p.ExtractValue(data.Data, config.DataExtractor.NamePath); found {
		if nameStr, ok := name.(string); ok && nameStr != "" {
			return nameStr
		}
	}

	return config.DataExtractor.FallbackName
}

// ConvertValue converts a string value to the appropriate type based on field definition
func (p *TemplateParser) ConvertValue(value string, fieldDef FieldDefinition) (interface{}, error) {
	switch fieldDef.Type {
	case "string", "text":
		return value, nil
	case "integer":
		return strconv.Atoi(value)
	case "boolean":
		return strconv.ParseBool(value)
	case "array":
		// Simple comma-separated parsing for arrays
		if value == "" {
			return []string{}, nil
		}
		return strings.Split(value, ","), nil
	default:
		return value, nil
	}
}
