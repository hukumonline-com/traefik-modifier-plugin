package traefik_modifier_plugin

import (
	"bytes"
	"log"
	"net/http"
	"strings"
	"text/template"
)

// HeaderConfig holds header modification configuration
type HeaderConfig map[string]string

// HeaderModifier handles header modifications
type HeaderModifier struct {
	templates       map[string]*template.Template
	templateStrings map[string]string // Store original template strings
}

// NewHeaderModifier creates a new header modifier with the given configuration
func NewHeaderModifier(config HeaderConfig) *HeaderModifier {
	hm := &HeaderModifier{
		templates:       make(map[string]*template.Template),
		templateStrings: make(map[string]string),
	}

	// Parse all header templates
	for headerName, templateStr := range config {
		if templateStr != "" {
			tmpl, err := template.New("header_"+headerName).
				Delims("[[", "]]").
				Parse(templateStr)
			if err != nil {
				log.Printf("Error parsing header template for %s: %v", headerName, err)
				continue
			}
			hm.templates[headerName] = tmpl
			hm.templateStrings[headerName] = templateStr // Store original template string
		}
	}

	return hm
}

// ModifyHeaders modifies request headers based on the configured templates and context
// Uses original headers map to determine whether to Set (replace) or Add (append)
func (hm *HeaderModifier) ModifyHeaders(req *http.Request, context *TemplateContext) error {
	if len(hm.templates) == 0 {
		return nil
	}

	// Capture original headers before any modifications
	originalHeaders := make(map[string]string)
	for name, values := range req.Header {
		if len(values) > 0 {
			originalHeaders[name] = values[0] // Keep original case for comparison
		}
	}

	// Create template data combining request info and context
	templateData := map[string]interface{}{
		"request": map[string]interface{}{
			"headers": convertHeaders(req.Header),
			"method":  req.Method,
			"url":     req.URL.String(),
			"path":    req.URL.Path,
		},
		"context": *context,
	}

	// Create modified headers map
	modifiedHeaders := make(map[string]string)

	// Process each header template to generate modified headers
	for headerName, tmpl := range hm.templates {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, templateData); err != nil {
			log.Printf("Error executing header template for %s: %v", headerName, err)
			continue
		}

		headerValue := strings.TrimSpace(buf.String())
		if headerValue != "" {
			modifiedHeaders[headerName] = headerValue
		}
	}

	// Apply headers: Set if exists in original, Add if new
	for headerName, headerValue := range modifiedHeaders {
		// Check if header exists in original headers (case-insensitive)
		headerExistsInOriginal := false
		var originalValue string

		for origName, origValue := range originalHeaders {
			if strings.EqualFold(origName, headerName) {
				headerExistsInOriginal = true
				originalValue = origValue
				break
			}
		}

		if headerExistsInOriginal {
			// Use Set (replace) for existing headers
			req.Header.Set(headerName, headerValue)
			log.Printf("Set header %s: %s (was: %s)", headerName, headerValue, originalValue)
		} else {
			// Use Add (append) for new headers
			req.Header.Add(headerName, headerValue)
			log.Printf("Added header %s: %s", headerName, headerValue)
		}
	}

	return nil
}

// AddHeader adds a new header without replacing existing ones
func (hm *HeaderModifier) AddHeader(req *http.Request, headerName, headerValue string, context *TemplateContext) error {
	if headerValue == "" {
		return nil
	}

	// Check if it's a template
	if containsTemplate(headerValue) {
		tmpl, err := template.New("dynamic").
			Delims("[[", "]]").
			Parse(headerValue)
		if err != nil {
			return err
		}

		templateData := map[string]interface{}{
			"request": map[string]interface{}{
				"headers": convertHeaders(req.Header),
				"method":  req.Method,
				"url":     req.URL.String(),
				"path":    req.URL.Path,
			},
			"context": *context,
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, templateData); err != nil {
			return err
		}
		headerValue = buf.String()
	}

	req.Header.Add(headerName, headerValue)
	log.Printf("Added header %s: %s", headerName, headerValue)
	return nil
}

// SetHeader sets a header value, optionally using templates
func (hm *HeaderModifier) SetHeader(req *http.Request, headerName, headerValue string, context *TemplateContext) error {
	if headerValue == "" {
		return nil
	}

	// Check if it's a template
	if containsTemplate(headerValue) {
		tmpl, err := template.New("dynamic").
			Delims("[[", "]]").
			Parse(headerValue)
		if err != nil {
			return err
		}

		templateData := map[string]interface{}{
			"request": map[string]interface{}{
				"headers": convertHeaders(req.Header),
				"method":  req.Method,
				"url":     req.URL.String(),
				"path":    req.URL.Path,
			},
			"context": *context,
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, templateData); err != nil {
			return err
		}
		headerValue = buf.String()
	}

	req.Header.Set(headerName, headerValue)
	log.Printf("Set header %s: %s", headerName, headerValue)
	return nil
}

// RemoveHeader removes a header from the request
func (hm *HeaderModifier) RemoveHeader(req *http.Request, headerName string) {
	req.Header.Del(headerName)
	log.Printf("Removed header %s", headerName)
}

// convertHeaders converts http.Header to map[string]string for template access
func convertHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	for name, values := range headers {
		if len(values) > 0 {
			// Use lowercase key for consistent access
			result[strings.ToLower(name)] = values[0] // Take first value
		}
	}
	return result
}

// containsTemplate checks if a string contains template syntax
func containsTemplate(s string) bool {
	return contains(s, "[[") && contains(s, "]]")
}

// contains checks if a string contains a substring (simple implementation)
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
