package modifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"text/template"

	"github.com/hukumonline-com/traefik-modifier/pkg"
)

// BodyModifier handles request and response body modifications
type BodyModifier struct {
	templateRequest  string
	templateResponse map[int]string
}

// NewBodyModifier creates a new body modifier instance
func NewBodyModifier(templateRequest string, templateResponse map[int]string) *BodyModifier {
	return &BodyModifier{
		templateRequest:  templateRequest,
		templateResponse: templateResponse,
	}
}

// ModifyRequestBodyWithContext handles request body modification using templates with context
func (bm *BodyModifier) ModifyRequestBodyWithContext(req *http.Request, ctx *TemplateContext) ([]byte, error) {
	if bm.templateRequest == "" || req.Body == nil {
		return nil, nil
	}

	// Read original body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	req.Body.Close()

	// Parse JSON body
	var requestData interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &requestData); err != nil {
			return nil, fmt.Errorf("failed to parse request JSON: %w", err)
		}
	}

	// Parse and execute template
	tmpl := template.Must(template.New("request").Funcs(pkg.SimpleFuncMap()).Delims("[[", "]]").Parse(bm.templateRequest))

	var buf bytes.Buffer
	templateData := map[string]interface{}{
		"request": map[string]interface{}{
			"body": requestData,
		},
	}

	// Add context if provided
	if ctx != nil {
		templateData["context"] = ctx
	}

	if err := tmpl.Execute(&buf, templateData); err != nil {
		return nil, fmt.Errorf("failed to execute request template: %w", err)
	}

	// Update request body
	newBody := buf.Bytes()
	req.Body = io.NopCloser(bytes.NewReader(newBody))
	req.ContentLength = int64(len(newBody))
	req.Header.Set("Content-Length", strconv.Itoa(len(newBody)))

	return body, nil
}

// ResponseWriter wraps http.ResponseWriter to capture response
type ResponseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

// NewResponseWriter creates a new response writer wrapper
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
		statusCode:     http.StatusOK,
	}
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	return rw.body.Write(b)
}

func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

func (rw *ResponseWriter) GetBody() []byte {
	return rw.body.Bytes()
}

func (rw *ResponseWriter) GetStatusCode() int {
	return rw.statusCode
}

// ModifyResponse handles response body modification
func (bm *BodyModifier) ModifyResponse(originalWriter http.ResponseWriter, capturedResponse *ResponseWriter, originalRequestBody []byte) error {
	return bm.ModifyResponseWithContext(originalWriter, capturedResponse, originalRequestBody, nil)
}

// ModifyResponseWithContext handles response body modification with context
func (bm *BodyModifier) ModifyResponseWithContext(originalWriter http.ResponseWriter, capturedResponse *ResponseWriter, originalRequestBody []byte, ctx *TemplateContext) error {
	if len(bm.templateResponse) == 0 {
		// No response masking configured, write original response
		originalWriter.WriteHeader(capturedResponse.statusCode)
		originalWriter.Write(capturedResponse.body.Bytes())
		return nil
	}

	// Check if we have a template for this status code
	templateStr, exists := bm.templateResponse[capturedResponse.statusCode]
	if !exists {
		// No masking for this status code, write original response
		originalWriter.WriteHeader(capturedResponse.statusCode)
		originalWriter.Write(capturedResponse.body.Bytes())
		return nil
	}

	// Parse original request body
	var requestData interface{}
	if len(originalRequestBody) > 0 {
		json.Unmarshal(originalRequestBody, &requestData)
	}

	// Parse response body
	var responseData interface{}
	responseBody := capturedResponse.body.Bytes()
	if len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, &responseData); err != nil {
			// If we can't parse as JSON, use raw string
			responseData = string(responseBody)
		}
	}

	// Parse and execute response template
	tmpl := template.Must(template.New("response").Funcs(pkg.SimpleFuncMap()).Delims("[[", "]]").Parse(templateStr))

	var buf bytes.Buffer
	templateData := map[string]interface{}{
		"request": map[string]interface{}{
			"body": requestData,
		},
		"response": map[string]interface{}{
			"body": responseData,
		},
	}

	// Add context if provided
	if ctx != nil {
		templateData["context"] = ctx
	}

	if err := tmpl.Execute(&buf, templateData); err != nil {
		return fmt.Errorf("response masking error: %v", err)
	}

	// Write modified response
	// Format JSON as minified (compact)
	var jsonData interface{}
	if err := json.Unmarshal(buf.Bytes(), &jsonData); err != nil {
		// If not valid JSON, use as is
		newBody := buf.Bytes()
		originalWriter.Header().Set("Content-Length", strconv.Itoa(len(newBody)))
		originalWriter.WriteHeader(capturedResponse.statusCode)
		originalWriter.Write(newBody)
		return nil
	}

	// Format JSON as minified (compact)
	formattedJSON, err := json.Marshal(jsonData)
	if err != nil {
		return fmt.Errorf("failed to format JSON: %v", err)
	}

	// Write formatted response
	originalWriter.Header().Set("Content-Length", strconv.Itoa(len(formattedJSON)))
	originalWriter.Header().Set("Content-Type", "application/json")
	originalWriter.WriteHeader(capturedResponse.statusCode)
	originalWriter.Write(formattedJSON)

	return nil
}
