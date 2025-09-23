package utils

import (
	"strings"
)

// SanitizeForTypst sanitizes user input to prevent Typst syntax errors.
func SanitizeForTypst(input string) string {
	if input == "" {
		return input
	}

	result := input

	// Apply backslash escaping first to avoid double-escaping
	result = strings.ReplaceAll(result, `\`, `\\`)

	// Then apply other escapes
	result = strings.ReplaceAll(result, `"`, `\"`)  // Escape double quotes for Typst strings
	result = strings.ReplaceAll(result, `$`, `\$`)  // Escape dollar sign (Typst uses $ for math)
	result = strings.ReplaceAll(result, `#`, `\#`)  // Escape hash (Typst uses # for functions)
	result = strings.ReplaceAll(result, "`", "\\`") // Escape backticks (using regular string literals)

	return result
}

// SanitizeFormValue is a convenience function for sanitizing HTTP form values.
func SanitizeFormValue(value string) string {
	return SanitizeForTypst(strings.TrimSpace(value))
}

// NormalizeURL ensures a URL has a proper protocol prefix.
func NormalizeURL(url string) string {
	if url == "" {
		return url
	}

	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "https://" + url
	}

	return url
}
