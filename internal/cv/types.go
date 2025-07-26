package cv

import (
	"time"
)

// CVData represents the complete CV data structure
type CVData struct {
	Contacts           Contacts         `json:"contacts" yaml:"contacts"`
	Position           string           `json:"position" yaml:"position"`
	Tagline            string           `json:"tagline,omitempty" yaml:"tagline,omitempty"`
	Jobs               []Job            `json:"jobs,omitempty" yaml:"jobs,omitempty"`
	Objective          string           `json:"objective,omitempty" yaml:"objective,omitempty"`
	Education          []Education      `json:"education,omitempty" yaml:"education,omitempty"`
	TechnicalExpertise []TechnicalSkill `json:"technical_expertise,omitempty" yaml:"technical_expertise,omitempty"`
	Skills             []string         `json:"skills,omitempty" yaml:"skills,omitempty"`
	Methodology        []string         `json:"methodology,omitempty" yaml:"methodology,omitempty"`
	Tools              []string         `json:"tools,omitempty" yaml:"tools,omitempty"`
	Achievements       []Achievement    `json:"achievements,omitempty" yaml:"achievements,omitempty"`
}

// Contacts represents contact information
type Contacts struct {
	Name     string    `json:"name" yaml:"name"`
	Title    string    `json:"title" yaml:"title"`
	Email    string    `json:"email" yaml:"email"`
	Address  string    `json:"address,omitempty" yaml:"address,omitempty"`
	Location string    `json:"location,omitempty" yaml:"location,omitempty"`
	LinkedIn *LinkInfo `json:"linkedin,omitempty" yaml:"linkedin,omitempty"`
	GitHub   *LinkInfo `json:"github,omitempty" yaml:"github,omitempty"`
	Website  *LinkInfo `json:"website,omitempty" yaml:"website,omitempty"`
}

// LinkInfo represents a link with display text
type LinkInfo struct {
	URL         string `json:"url" yaml:"url"`
	DisplayText string `json:"displayText" yaml:"displayText"`
}

// Job represents work experience
type Job struct {
	Position    string   `json:"position" yaml:"position"`
	Company     Company  `json:"company" yaml:"company"`
	Product     *Product `json:"product,omitempty" yaml:"product,omitempty"`
	Description []string `json:"description,omitempty" yaml:"description,omitempty"`
	From        string   `json:"from" yaml:"from"`
	To          string   `json:"to" yaml:"to"`
	Tags        []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Location    string   `json:"location,omitempty" yaml:"location,omitempty"`
}

// Company represents company information
type Company struct {
	Name string `json:"name" yaml:"name"`
	Link string `json:"link,omitempty" yaml:"link,omitempty"`
}

// Product represents product information
type Product struct {
	Name string `json:"name" yaml:"name"`
	Link string `json:"link,omitempty" yaml:"link,omitempty"`
}

// Education represents educational background
type Education struct {
	Place    Institution `json:"place" yaml:"place"`
	Degree   string      `json:"degree" yaml:"degree"`
	Major    string      `json:"major" yaml:"major"`
	Track    string      `json:"track,omitempty" yaml:"track,omitempty"`
	From     string      `json:"from" yaml:"from"`
	To       string      `json:"to" yaml:"to"`
	Location string      `json:"location,omitempty" yaml:"location,omitempty"`
}

// Institution represents educational institution
type Institution struct {
	Name string `json:"name" yaml:"name"`
	Link string `json:"link,omitempty" yaml:"link,omitempty"`
}

// TechnicalSkill represents technical expertise with proficiency level
type TechnicalSkill struct {
	Name  string `json:"name" yaml:"name"`
	Level int    `json:"level" yaml:"level"` // 1-5 scale
}

// Achievement represents achievements or certifications
type Achievement struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
}

// Template represents a CV template
type Template struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Author      string                 `json:"author"`
	Fields      map[string]interface{} `json:"fields"`
}

// GenerationRequest represents a CV generation request
type GenerationRequest struct {
	TemplateID string `json:"template_id"`
	Data       CVData `json:"data"`
	Format     string `json:"format"` // "pdf", "png", etc.
}

// GenerationResult represents the result of CV generation
type GenerationResult struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message,omitempty"`
	Filename  string    `json:"filename,omitempty"`
	Data      []byte    `json:"data,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
