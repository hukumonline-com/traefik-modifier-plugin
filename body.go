package traefik_modifier_plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"text/template"

	"github.com/hukumonline-com/traefik-modifier-plugin/pkg"
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
func (bm *BodyModifier) ModifyRequestBodyWithContext(req *http.Request, ctx *TemplateContext) ([]byte, []byte, error) {
	if bm.templateRequest == "" || req.Body == nil {
		return nil, nil, nil
	}

	// Read original body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read request body: %w", err)
	}
	req.Body.Close()

	// Parse JSON body
	var requestData interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &requestData); err != nil {
			return nil, nil, fmt.Errorf("failed to parse request JSON: %w", err)
		}
	}

	// Parse and execute template
	tmpl := template.Must(template.New("request").Funcs(pkg.SimpleFuncMap()).Delims("[[", "]]").Parse(bm.templateRequest))

	var buf bytes.Buffer
	templateData := map[string]interface{}{
		"request": map[string]interface{}{
			"api": map[string]interface{}{
				"body": requestData,
			},
		},
	}

	// Add context if provided
	if ctx != nil {
		templateData["context"] = ctx
	}

	if err := tmpl.Execute(&buf, templateData); err != nil {
		return nil, nil, fmt.Errorf("failed to execute request template: %w", err)
	}

	// Clean and update request body
	newBody := buf.Bytes()

	// Clean JSON by removing "<no value>" strings
	cleanedBody := bytes.ReplaceAll(newBody, []byte(`"<no value>"`), []byte(`""`))

	req.Body = io.NopCloser(bytes.NewReader(cleanedBody))
	req.ContentLength = int64(len(cleanedBody))
	req.Header.Set("Content-Length", strconv.Itoa(len(cleanedBody)))

	return body, cleanedBody, nil
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
func (bm *BodyModifier) ModifyResponse(originalWriter http.ResponseWriter, capturedResponse *ResponseWriter, originalRequestBody, modifiedRequestBody []byte) error {
	return bm.ModifyResponseWithContext(originalWriter, capturedResponse, originalRequestBody, modifiedRequestBody, nil)
}

// ModifyResponseWithContext handles response body modification with context
func (bm *BodyModifier) ModifyResponseWithContext(originalWriter http.ResponseWriter, capturedResponse *ResponseWriter, originalRequestBody, modifiedRequestBody []byte, ctx *TemplateContext) error {
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
	var requestDataOriginal interface{}
	var requestDataModified interface{}

	if len(originalRequestBody) > 0 {
		json.Unmarshal(originalRequestBody, &requestDataOriginal)
	}

	if len(modifiedRequestBody) > 0 {
		json.Unmarshal(modifiedRequestBody, &requestDataModified)
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
			"api": map[string]interface{}{
				"body": requestDataOriginal,
			},
			"modified": map[string]interface{}{
				"body": requestDataModified,
			},
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
	// Check if response is valid JSON and clean it
	responseBytes := buf.Bytes()
	var formattedJSON []byte

	var jsonData interface{}
	if err := json.Unmarshal(responseBytes, &jsonData); err != nil {
		// If not valid JSON, use as is
		originalWriter.Header().Set("Content-Length", strconv.Itoa(len(responseBytes)))
		originalWriter.WriteHeader(capturedResponse.statusCode)
		originalWriter.Write(responseBytes)
		return nil
	}

	// Clean JSON by removing "<no value>" strings and format as minified
	cleanedJSON := bytes.ReplaceAll(responseBytes, []byte(`"<no value>"`), []byte(`""`))

	// Re-parse to ensure valid JSON structure after cleaning
	if err := json.Unmarshal(cleanedJSON, &jsonData); err != nil {
		// If cleaning broke JSON, fallback to original marshaling
		formattedJSON, err = json.Marshal(jsonData)
		if err != nil {
			return fmt.Errorf("failed to format JSON: %v", err)
		}
	} else {
		// Use cleaned JSON directly (already minified from template output)
		formattedJSON = cleanedJSON
	}

	// Write formatted response
	originalWriter.Header().Set("Content-Length", strconv.Itoa(len(formattedJSON)))
	originalWriter.Header().Set("Content-Type", "application/json")
	originalWriter.WriteHeader(capturedResponse.statusCode)
	originalWriter.Write(formattedJSON)

	return nil
}
