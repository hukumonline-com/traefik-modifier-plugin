package traefik_modifier_plugin

import (
	"bytes"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"github.com/hukumonline-com/traefik-modifier-plugin/pkg"
)

// QueryConfig holds the query transformation configuration
type QueryConfig struct {
	Transform map[string]string `json:"transform,omitempty"`
}

// QueryModifier handles query parameter transformations
type QueryModifier struct {
	transforms map[string]string
}

// NewQueryModifier creates a new query modifier instance
func NewQueryModifier(transforms map[string]string) *QueryModifier {
	return &QueryModifier{
		transforms: transforms,
	}
}

// ModifyQueryWithContext handles query parameter modification using templates with context
func (qm *QueryModifier) ModifyQueryWithContext(req *http.Request, ctx *TemplateContext) error {
	if len(qm.transforms) == 0 {
		return nil
	}

	// Get current query parameters
	values := req.URL.Query()

	// Create template data from request
	templateData := map[string]interface{}{
		"request": map[string]interface{}{
			"query":  queryParamsToMap(values),
			"header": headerToMap(req.Header),
			"method": req.Method,
			"path":   req.URL.Path,
		},
	}

	// Add context if provided
	if ctx != nil {
		templateData["context"] = ctx
	}

	log.Printf("Query modifier template data: %+v", templateData)

	// Apply transformations
	for targetParam, templateStr := range qm.transforms {
		// Parse and execute template
		tmpl, err := template.New("query").Funcs(pkg.SimpleFuncMap()).Delims("[[", "]]").Parse(templateStr)
		if err != nil {
			log.Printf("Failed to parse query template for %s: %v", targetParam, err)
			continue
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, templateData); err != nil {
			log.Printf("Failed to execute query template for %s: %v", targetParam, err)
			continue
		}

		result := buf.String()

		// Clean the result by removing "<no value>" strings
		result = strings.ReplaceAll(result, "<no value>", "")

		if result != "" {
			if values.Has(targetParam) {
				log.Printf("Overwriting existing query parameter %s", targetParam)
				values.Set(targetParam, result)
			} else {
				log.Printf("Setting new query parameter %s", targetParam)
				values.Add(targetParam, result)
			}

			// Set the transformed value
			log.Printf("Query parameter %s transformed to: %s", targetParam, result)
		}
	}

	// Update the request URL with modified query parameters
	req.URL.RawQuery = values.Encode()
	req.RequestURI = req.URL.RequestURI()

	return nil
}

// queryParamsToMap converts url.Values to a simple map for template usage
func queryParamsToMap(values url.Values) map[string]interface{} {
	result := make(map[string]interface{})
	for key, vals := range values {
		if len(vals) == 1 {
			result[key] = vals[0]
		} else {
			result[key] = vals
		}
	}
	return result
}

// headerToMap converts http.Header to a simple map for template usage
func headerToMap(headers http.Header) map[string]interface{} {
	result := make(map[string]interface{})
	for key, vals := range headers {
		if len(vals) == 1 {
			result[key] = vals[0]
		} else {
			result[key] = vals
		}
	}
	return result
}
