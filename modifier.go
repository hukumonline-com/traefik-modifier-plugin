package traefik_modifier_plugin

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func init() {
	log.SetOutput(os.Stdout)
}

// Config holds the plugin configuration
type Config struct {
	ModifierRequest  string         `json:"modifier_request,omitempty"`
	ModifierResponse map[int]string `json:"modifier_response,omitempty"`
	ModifierQuery    *QueryConfig   `json:"modifier_query,omitempty"`
	ModifierHeader   HeaderConfig   `json:"modifier_header,omitempty"`
}

// TemplateContext holds context data for templates
type TemplateContext map[string]interface{}

// CreateConfig creates and initializes the plugin configuration
func CreateConfig() *Config {
	return &Config{}
}

// modifier holds the plugin instance
type modifier struct {
	name           string
	next           http.Handler
	bodyModifier   *BodyModifier
	queryModifier  *QueryModifier
	headerModifier *HeaderModifier
	context        *TemplateContext
}

// New creates and returns a new modifier plugin instance
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	// Initialize body modifier
	bodyModifier := NewBodyModifier(config.ModifierRequest, config.ModifierResponse)

	// Initialize query modifier
	var queryModifier *QueryModifier
	if config.ModifierQuery != nil && len(config.ModifierQuery.Transform) > 0 {
		queryModifier = NewQueryModifier(config.ModifierQuery.Transform)
	}

	// Initialize header modifier
	var headerModifier *HeaderModifier
	if len(config.ModifierHeader) > 0 {
		headerModifier = NewHeaderModifier(config.ModifierHeader)
	}

	// Initialize template context
	templateContext := &TemplateContext{}

	plugin := &modifier{
		name:           name,
		next:           next,
		bodyModifier:   bodyModifier,
		queryModifier:  queryModifier,
		headerModifier: headerModifier,
		context:        templateContext,
	}

	return plugin, nil
}

// ServeHTTP processes the HTTP request and response
func (m *modifier) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var err error
	var originalRequestBody, modifiedRequestBody []byte

	m.context = &TemplateContext{
		"unixtime": time.Now().UnixNano(),
	}

	// Handle header modification
	if m.headerModifier != nil {
		if err := m.headerModifier.ModifyHeaders(req, m.context); err != nil {
			log.Printf("Header modification error: %v", err)
		}
	}

	// Handle query parameter modification
	if m.queryModifier != nil {
		if err := m.queryModifier.ModifyQueryWithContext(req, m.context); err != nil {
			log.Printf("Query modification error: %v", err)
		}
	}

	// Handle request body masking
	if m.bodyModifier != nil {
		originalRequestBody, modifiedRequestBody, err = m.bodyModifier.ModifyRequestBodyWithContext(req, m.context)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Request masking error: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Handle response masking if configured
	if m.bodyModifier != nil && len(m.bodyModifier.templateResponse) > 0 {
		m.handleResponseMasking(rw, req, originalRequestBody, modifiedRequestBody)
		return
	}

	// No response masking, proceed normally
	m.next.ServeHTTP(rw, req)
}

// handleResponseMasking handles response body modification
func (m *modifier) handleResponseMasking(rw http.ResponseWriter, req *http.Request, originalRequestBody, modifiedRequestBody []byte) {
	// Create a response writer to capture the response
	captureWriter := NewResponseWriter(rw)

	// Call next handler
	m.next.ServeHTTP(captureWriter, req)

	// Use body modifier to handle response modification with context
	if err := m.bodyModifier.ModifyResponseWithContext(rw, captureWriter, originalRequestBody, modifiedRequestBody, m.context); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}
