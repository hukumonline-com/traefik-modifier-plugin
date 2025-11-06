package traefik_modifier_plugin

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestHeaderModifier_ModifyHeaders(t *testing.T) {
	tests := []struct {
		name            string
		config          HeaderConfig
		requestHeaders  map[string]string
		expectedHeaders map[string]string
	}{
		{
			name: "Simple Authorization header modification",
			config: HeaderConfig{
				"Authorization": `[[ if eq (index .request.headers "x-api-key") "sk-didin" ]]Bearer sk-didin[[ else ]]Bearer sk-default[[ end ]]`,
			},
			requestHeaders: map[string]string{
				"X-Api-Key": "sk-didin",
			},
			expectedHeaders: map[string]string{
				"Authorization": "Bearer sk-didin",
			},
		},
		{
			name: "Default Authorization when no API key",
			config: HeaderConfig{
				"Authorization": `[[ if eq (index .request.headers "x-api-key") "sk-didin" ]]Bearer sk-didin[[ else ]]Bearer sk-default[[ end ]]`,
			},
			requestHeaders: map[string]string{},
			expectedHeaders: map[string]string{
				"Authorization": "Bearer sk-default",
			},
		},
		{
			name: "Multiple header modifications",
			config: HeaderConfig{
				"X-Request-ID": "req_[[ .context.unixtime ]]",
				"X-Method":     "[[ .request.method ]]",
			},
			requestHeaders: map[string]string{},
			expectedHeaders: map[string]string{
				"X-Method": "GET",
			},
		},
		{
			name: "Conditional header based on existing header",
			config: HeaderConfig{
				"X-Debug": `[[ if eq .request.headers.debug "true" ]]enabled[[ else ]]disabled[[ end ]]`,
			},
			requestHeaders: map[string]string{
				"Debug": "true",
			},
			expectedHeaders: map[string]string{
				"X-Debug": "enabled",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create header modifier
			hm := NewHeaderModifier(tt.config)

			// Create test request
			req := httptest.NewRequest("GET", "http://example.com/test", nil)

			// Set request headers
			for key, value := range tt.requestHeaders {
				req.Header.Set(key, value)
			}

			// Create context
			context := &TemplateContext{
				"unixtime": time.Now().UnixNano(),
			}

			// Modify headers
			err := hm.ModifyHeaders(req, context)
			if err != nil {
				t.Fatalf("ModifyHeaders() error = %v", err)
			}

			// Check expected headers
			for key, expectedValue := range tt.expectedHeaders {
				actualValue := req.Header.Get(key)
				if key == "X-Request-ID" {
					// For timestamp-based headers, just check if it's not empty
					if actualValue == "" {
						t.Errorf("Expected %s to be set, but it was empty", key)
					}
				} else if actualValue != expectedValue {
					t.Errorf("Expected header %s = %s, got %s", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestHeaderModifier_SetHeader(t *testing.T) {
	hm := NewHeaderModifier(HeaderConfig{})
	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	context := &TemplateContext{"unixtime": time.Now().UnixNano()}

	// Test simple header setting
	err := hm.SetHeader(req, "X-Test", "test-value", context)
	if err != nil {
		t.Fatalf("SetHeader() error = %v", err)
	}

	if req.Header.Get("X-Test") != "test-value" {
		t.Errorf("Expected X-Test = test-value, got %s", req.Header.Get("X-Test"))
	}

	// Test template header setting
	req.Header.Set("Original", "original-value")
	err = hm.SetHeader(req, "X-Template", "prefix-[[ .request.headers.original ]]", context)
	if err != nil {
		t.Fatalf("SetHeader() with template error = %v", err)
	}

	if req.Header.Get("X-Template") != "prefix-original-value" {
		t.Errorf("Expected X-Template = prefix-original-value, got %s", req.Header.Get("X-Template"))
	}
}

func TestHeaderModifier_AddHeader(t *testing.T) {
	hm := NewHeaderModifier(HeaderConfig{})
	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	context := &TemplateContext{"unixtime": time.Now().UnixNano()}

	// Add first header
	err := hm.AddHeader(req, "X-Multi", "value1", context)
	if err != nil {
		t.Fatalf("AddHeader() error = %v", err)
	}

	// Add second header with same name
	err = hm.AddHeader(req, "X-Multi", "value2", context)
	if err != nil {
		t.Fatalf("AddHeader() error = %v", err)
	}

	values := req.Header.Values("X-Multi")
	if len(values) != 2 {
		t.Errorf("Expected 2 values for X-Multi, got %d", len(values))
	}
	if values[0] != "value1" || values[1] != "value2" {
		t.Errorf("Expected [value1, value2], got %v", values)
	}
}

func TestHeaderModifier_RemoveHeader(t *testing.T) {
	hm := NewHeaderModifier(HeaderConfig{})
	req := httptest.NewRequest("GET", "http://example.com/test", nil)

	// Set header first
	req.Header.Set("X-Remove", "to-be-removed")

	// Verify it exists
	if req.Header.Get("X-Remove") != "to-be-removed" {
		t.Errorf("Header X-Remove was not set properly")
	}

	// Remove header
	hm.RemoveHeader(req, "X-Remove")

	// Verify it's removed
	if req.Header.Get("X-Remove") != "" {
		t.Errorf("Header X-Remove was not removed")
	}
}

func TestHeaderModifier_DynamicHeaderHandling(t *testing.T) {
	tests := []struct {
		name              string
		config            HeaderConfig
		existingHeaders   map[string]string
		expectedOperation string // "set" or "add"
		expectedValue     string
	}{
		{
			name: "Set header when header exists originally",
			config: HeaderConfig{
				"Authorization": "Bearer new-token",
			},
			existingHeaders: map[string]string{
				"Authorization": "Bearer old-token",
			},
			expectedOperation: "set",
			expectedValue:     "Bearer new-token",
		},
		{
			name: "Add header when header doesn't exist",
			config: HeaderConfig{
				"X-New-Header": "new-value",
			},
			existingHeaders:   map[string]string{},
			expectedOperation: "add",
			expectedValue:     "new-value",
		},
		{
			name: "Set header when template references the same header",
			config: HeaderConfig{
				// "Authorization": `[[ if eq (index .request.headers "authorization") "Bearer old" ]]Bearer updated[[ else ]]Bearer default[[ end ]]`,
				"Authorization": `[[ if eq (index .request.headers "authorization") "Bearer old" ]]Bearer updated[[ else ]]Bearer default[[ end ]]`,
			},
			existingHeaders: map[string]string{
				"Authorization": "Bearer old",
			},
			expectedOperation: "set",
			expectedValue:     "Bearer updated",
		},
		{
			name: "Set header when template references original headers",
			config: HeaderConfig{
				"X-Modified": `[[ (index .request.headers "authorization") ]]`,
			},
			existingHeaders: map[string]string{
				"Authorization": "Bearer test",
			},
			expectedOperation: "set",
			expectedValue:     "Bearer test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create header modifier
			hm := NewHeaderModifier(tt.config)

			// Create test request
			req := httptest.NewRequest("GET", "http://example.com/test", nil)

			// Set existing headers
			for key, value := range tt.existingHeaders {
				req.Header.Set(key, value)
			}

			// Create context
			context := &TemplateContext{
				"unixtime": time.Now().UnixNano(),
			}

			// Modify headers
			err := hm.ModifyHeaders(req, context)
			if err != nil {
				t.Fatalf("ModifyHeaders() error = %v", err)
			}

			// Verify the header value is as expected
			for headerName := range tt.config {
				actualValue := req.Header.Get(headerName)
				if actualValue != tt.expectedValue {
					t.Errorf("Expected header %s = %s, got %s", headerName, tt.expectedValue, actualValue)
				}
			}
		})
	}
}

func TestContainsTemplate(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"simple string", false},
		{"[[ .request.headers.test ]]", true},
		{"prefix [[ .context.time ]] suffix", true},
		{"[[incomplete", false},
		{"incomplete]]", false},
		{"", false},
		{"[[]]", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := containsTemplate(tt.input)
			if result != tt.expected {
				t.Errorf("containsTemplate(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
