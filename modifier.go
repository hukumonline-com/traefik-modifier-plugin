package modifier

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
	Request  string         `json:"request,omitempty"`
	Response map[int]string `json:"response,omitempty"`
	Query    *QueryConfig   `json:"query,omitempty"`
}

// TemplateContext holds context data for templates
type TemplateContext map[string]interface{}

// CreateConfig creates and initializes the plugin configuration
func CreateConfig() *Config {
	return &Config{}
}

// modifier holds the plugin instance
type modifier struct {
	name          string
	next          http.Handler
	bodyModifier  *BodyModifier
	queryModifier *QueryModifier
	context       *TemplateContext
}

// New creates and returns a new modifier plugin instance
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	// Initialize body modifier
	bodyModifier := NewBodyModifier(config.Request, config.Response)

	// Initialize query modifier
	var queryModifier *QueryModifier
	if config.Query != nil && len(config.Query.Transform) > 0 {
		queryModifier = NewQueryModifier(config.Query.Transform)
	}

	// Initialize template context
	templateContext := &TemplateContext{}

	plugin := &modifier{
		name:          name,
		next:          next,
		bodyModifier:  bodyModifier,
		queryModifier: queryModifier,
		context:       templateContext,
	}

	return plugin, nil
}

// ServeHTTP processes the HTTP request and response
func (m *modifier) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var err error
	var originalRequestBody []byte

	// Update context timestamp for each request
	m.context = &TemplateContext{
		"unixtime": time.Now().Unix(),
	}

	// Handle query parameter modification
	if m.queryModifier != nil {
		if err := m.queryModifier.ModifyQueryWithContext(req, m.context); err != nil {
			log.Printf("Query modification error: %v", err)
		}
	}

	// Handle request body masking
	if m.bodyModifier != nil {
		originalRequestBody, err = m.bodyModifier.ModifyRequestBodyWithContext(req, m.context)
		if err != nil {
			http.Error(rw, fmt.Sprintf("Request masking error: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Handle response masking if configured
	if m.bodyModifier != nil && len(m.bodyModifier.templateResponse) > 0 {
		m.handleResponseMasking(rw, req, originalRequestBody)
		return
	}

	// No response masking, proceed normally
	m.next.ServeHTTP(rw, req)
}

// handleResponseMasking handles response body modification
func (m *modifier) handleResponseMasking(rw http.ResponseWriter, req *http.Request, originalRequestBody []byte) {
	// Create a response writer to capture the response
	captureWriter := NewResponseWriter(rw)

	// Call next handler
	m.next.ServeHTTP(captureWriter, req)

	// Use body modifier to handle response modification with context
	if err := m.bodyModifier.ModifyResponseWithContext(rw, captureWriter, originalRequestBody, m.context); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}
