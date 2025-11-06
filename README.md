# Traefik Modifier Plugin - Complete Documentation

Plugin Traefik untuk modifikasi request dan response secara dinamis menggunakan template Go dengan dukungan header, query, dan body modification.

## Table of Contents
- [Overview](#overview)
- [Template Variables](#template-variables)
- [Request API Variables](#request-api-variables)
- [Modifier Request Variables](#modifier-request-variables)
- [Modifier Response Variables](#modifier-response-variables)
- [Context Variables](#context-variables)
- [Configuration Examples](#configuration-examples)
- [Template Syntax](#template-syntax)

## Overview

Plugin ini mendukung tiga jenis modifikasi utama:
1. **Header Modification** - Memodifikasi HTTP headers
2. **Query Modification** - Memodifikasi query parameters
3. **Body Modification** - Memodifikasi request dan response body

Setiap modifikasi menggunakan template Go dengan delimiter `[[` dan `]]`.

## Template Variables

### Global Template Structure

Semua template memiliki akses ke struktur data berikut:

```go
{
  "request": {
    "headers": map[string]string,     // HTTP headers (lowercase keys)
    "method": string,                 // HTTP method (GET, POST, etc.)
    "url": string,                   // Full URL
    "path": string,                  // URL path
    "api": {                         // Available in modifier request/response
      "body": map[string]interface{} // Parsed request body
    },
    "modified": {                    // Available in modifier response
      "body": map[string]interface{} // Modified request body
    }
  },
  "response": {                      // Available in modifier response only
    "body": map[string]interface{},  // Original response body
    "status": int,                   // HTTP status code
    "headers": map[string]string     // Response headers
  },
  "context": {
    "unixtime": int64               // Current Unix timestamp (nanoseconds)
  }
}
```

## Request API Variables

### Dalam Modifier Header dan Query

Variables yang tersedia untuk template header dan query modification:

#### Request Information
```yaml
# HTTP Method
Method: "[[ .request.method ]]"

# URL Components  
FullURL: "[[ .request.url ]]"
Path: "[[ .request.path ]]"

# Headers (case-insensitive access)
ApiKey: "[[ index .request.headers \"x-api-key\" ]]"
UserAgent: "[[ index .request.headers \"user-agent\" ]]"
ContentType: "[[ index .request.headers \"content-type\" ]]"

# Conditional based on headers
ConditionalValue: |
  [[ if eq (index .request.headers "x-api-key") "secret" ]]
    authorized
  [[ else ]]
    unauthorized
  [[ end ]]
```

#### Context Data
```yaml
# Timestamp
RequestID: "req_[[ .context.unixtime ]]"
Timestamp: "[[ .context.unixtime ]]"
```

### Example Header Modification
```yaml
ModifierHeader:
  Authorization: |
    [[ if eq (index .request.headers "x-api-key") "sk-didin" ]]
      Bearer sk-didin
    [[ else if eq (index .request.headers "x-api-key") "sk-test" ]]
      Bearer sk-test
    [[ else ]]
      Bearer default-token
    [[ end ]]
  X-Request-ID: "req_[[ .context.unixtime ]]"
  X-Original-Method: "[[ .request.method ]]"
  X-Original-Path: "[[ .request.path ]]"
```

### Example Query Modification
```yaml
ModifierQuery:
  Transform:
    question_id: "ask_[[ .context.unixtime ]]"
    timestamp: "[[ .context.unixtime ]]"
    method: "[[ .request.method ]]"
    source: |
      [[ if eq .request.path "/api/v1/chat" ]]
        chat-api
      [[ else ]]
        other-api
      [[ end ]]
```

## Modifier Request Variables

### Request Body Modification

Variables yang tersedia dalam template modifier request:

#### Full Request Context
```go
{
  "request": {
    "headers": map[string]string,     // All request headers
    "method": string,                 // HTTP method
    "url": string,                   // Complete URL
    "path": string,                  // URL path
    "api": {
      "body": map[string]interface{} // Original parsed request body
    }
  },
  "context": {
    "unixtime": int64               // Current timestamp
  }
}
```

#### Accessing Request Body Data
```yaml
ModifierRequest: |
  {
    "question": "[[ .request.api.body.ask ]]",
    "user_id": "[[ .request.api.body.user_id ]]",
    "timestamp": "[[ .context.unixtime ]]",
    "method": "[[ .request.method ]]",
    "headers": {
      "authorization": "[[ index .request.headers \"authorization\" ]]",
      "user-agent": "[[ index .request.headers \"user-agent\" ]]"
    },
    "metadata": {
      "original_path": "[[ .request.path ]]",
      "request_id": "req_[[ .context.unixtime ]]"
    }
  }
```

#### Complex Request Body Transformation
```yaml
ModifierRequest: |
  {
    "query": {
      "text": "[[ .request.api.body.question ]]",
      "context": "[[ .request.api.body.context ]]",
      "parameters": {
        "max_tokens": [[ .request.api.body.max_tokens ]],
        "temperature": [[ .request.api.body.temperature ]]
      }
    },
    "metadata": {
      "source": "traefik-modifier",
      "timestamp": [[ .context.unixtime ]],
      "request_method": "[[ .request.method ]]",
      "api_key": "[[ index .request.headers \"x-api-key\" ]]"
    },
    "original_request": {
      "url": "[[ .request.url ]]",
      "path": "[[ .request.path ]]"
    }
  }
```

### Conditional Request Modification
```yaml
ModifierRequest: |
  [[ if eq .request.method "POST" ]]
  {
    "action": "create",
    "data": "[[ .request.api.body.data ]]",
    "timestamp": [[ .context.unixtime ]]
  }
  [[ else if eq .request.method "GET" ]]
  {
    "action": "read",
    "query": "[[ .request.api.body.query ]]",
    "timestamp": [[ .context.unixtime ]]
  }
  [[ else ]]
  {
    "action": "other",
    "method": "[[ .request.method ]]",
    "timestamp": [[ .context.unixtime ]]
  }
  [[ end ]]
```

## Modifier Response Variables

### Response Body Modification

Variables yang tersedia dalam template modifier response:

#### Full Response Context
```go
{
  "request": {
    "headers": map[string]string,     // Original request headers
    "method": string,                 // HTTP method
    "url": string,                   // Complete URL
    "path": string,                  // URL path
    "api": {
      "body": map[string]interface{} // Original request body
    },
    "modified": {
      "body": map[string]interface{} // Modified request body (after ModifierRequest)
    }
  },
  "response": {
    "body": map[string]interface{},  // Original response body from backend
    "status": int,                   // HTTP status code
    "headers": map[string]string     // Response headers
  },
  "context": {
    "unixtime": int64               // Current timestamp
  }
}
```

#### Accessing Response Data
```yaml
ModifierResponse:
  "200": |
    {
      "id": [[ .response.body.id ]],
      "answer": "[[ .response.body.text ]]",
      "metadata": {
        "original_question": "[[ .request.api.body.ask ]]",
        "modified_question": "[[ .request.modified.body.question ]]",
        "processing_time": [[ .context.unixtime ]],
        "status": [[ .response.status ]]
      },
      "request_info": {
        "method": "[[ .request.method ]]",
        "path": "[[ .request.path ]]",
        "user_agent": "[[ index .request.headers \"user-agent\" ]]"
      }
    }
```

#### Complex Response Transformation
```yaml
ModifierResponse:
  "200": |
    {
      "success": true,
      "data": {
        "response_id": [[ .response.body.id ]],
        "content": "[[ .response.body.message ]]",
        "confidence": [[ .response.body.confidence ]],
        "metadata": {
          "model": "[[ .response.body.model ]]",
          "tokens_used": [[ .response.body.usage.total_tokens ]]
        }
      },
      "request_context": {
        "original_query": "[[ .request.api.body.query ]]",
        "processed_query": "[[ .request.modified.body.question ]]",
        "timestamp": [[ .request.modified.body.timestamp ]],
        "request_id": "req_[[ .context.unixtime ]]"
      },
      "response_metadata": {
        "status_code": [[ .response.status ]],
        "content_type": "[[ index .response.headers \"content-type\" ]]",
        "server": "[[ index .response.headers \"server\" ]]"
      }
    }
  "400": |
    {
      "success": false,
      "error": {
        "message": "[[ .response.body.error ]]",
        "code": [[ .response.status ]],
        "timestamp": [[ .context.unixtime ]]
      },
      "request_info": {
        "method": "[[ .request.method ]]",
        "path": "[[ .request.path ]]",
        "original_body": "[[ .request.api.body ]]"
      }
    }
```

#### Array Processing in Response
```yaml
ModifierResponse:
  "200": |
    {
      "results": [
        [[ $dataList := .response.body.data_array ]]
        [[ $listLen := len $dataList ]]
        [[ range $index, $element := $dataList ]]
          [[ if $index ]], [[ end ]]
          {
            "id": "[[ $element.id ]]",
            "value": "[[ $element.value ]]",
            "processed": true,
            "index": [[ $index ]]
          }
        [[ end ]]
      ],
      "total_items": [[ len .response.body.data_array ]],
      "request_metadata": {
        "query": "[[ .request.api.body.search ]]",
        "timestamp": [[ .context.unixtime ]]
      }
    }
```

## Context Variables

### Available Context Data

```yaml
# Unix timestamp (nanoseconds)
Timestamp: "[[ .context.unixtime ]]"

# Derived values
RequestID: "req_[[ .context.unixtime ]]"
SessionID: "session_[[ .context.unixtime ]]"

# Formatted timestamp (you can add custom formatting)
HumanTime: "[[ .context.unixtime ]]"  # Note: This is raw nanoseconds
```

## Configuration Examples

### Complete Dynamic Configuration

```yaml
http:
  middlewares:
    full-modifier:
      plugin:
        modifier:
          # Header modification
          ModifierHeader:
            Authorization: |
              [[ if eq (index .request.headers "x-api-key") "sk-prod" ]]
                Bearer production-token
              [[ else if eq (index .request.headers "x-api-key") "sk-dev" ]]
                Bearer development-token
              [[ else ]]
                Bearer default-token
              [[ end ]]
            X-Request-ID: "req_[[ .context.unixtime ]]"
            X-Source-Method: "[[ .request.method ]]"
            X-Original-Path: "[[ .request.path ]]"
            
          # Query parameter modification
          ModifierQuery:
            Transform:
              request_id: "req_[[ .context.unixtime ]]"
              source_method: "[[ .request.method ]]"
              
          # Request body modification
          ModifierRequest: |
            {
              "query": "[[ .request.api.body.question ]]",
              "context": "[[ .request.api.body.context ]]",
              "metadata": {
                "request_id": "req_[[ .context.unixtime ]]",
                "source_ip": "[[ index .request.headers \"x-forwarded-for\" ]]",
                "user_agent": "[[ index .request.headers \"user-agent\" ]]",
                "api_version": "v1",
                "timestamp": [[ .context.unixtime ]]
              },
              "settings": {
                "max_tokens": [[ .request.api.body.max_tokens ]],
                "temperature": [[ .request.api.body.temperature ]]
              }
            }
            
          # Response body modification
          ModifierResponse:
            "200": |
              {
                "success": true,
                "response_id": [[ .response.body.id ]],
                "answer": "[[ .response.body.text ]]",
                "confidence": [[ .response.body.confidence ]],
                "request_context": {
                  "original_question": "[[ .request.api.body.question ]]",
                  "processed_question": "[[ .request.modified.body.query ]]",
                  "request_id": "[[ .request.modified.body.metadata.request_id ]]",
                  "timestamp": [[ .request.modified.body.metadata.timestamp ]]
                },
                "processing_info": {
                  "status": [[ .response.status ]],
                  "model": "[[ .response.body.model ]]",
                  "tokens": [[ .response.body.usage.total_tokens ]]
                }
              }
            "400": |
              {
                "success": false,
                "error": {
                  "message": "[[ .response.body.error.message ]]",
                  "code": "[[ .response.body.error.code ]]",
                  "type": "[[ .response.body.error.type ]]"
                },
                "request_info": {
                  "method": "[[ .request.method ]]",
                  "path": "[[ .request.path ]]",
                  "timestamp": [[ .context.unixtime ]]
                }
              }
```

## Template Syntax

### Basic Syntax Rules

1. **Delimiters**: Gunakan `[[` dan `]]` untuk template expressions
2. **Variable Access**: Gunakan dot notation (`.request.method`)
3. **Header Access**: Gunakan `index` function untuk header names dengan special characters
4. **Conditionals**: Gunakan `if`, `else if`, `else`, `end`
5. **Loops**: Gunakan `range` untuk array processing

### Header Access Patterns

```yaml
# ✅ CORRECT - Using index function for headers with special characters
Authorization: "[[ index .request.headers \"x-api-key\" ]]"

# ✅ CORRECT - Simple header names
SimpleHeader: "[[ .request.headers.simple ]]"

# ❌ INCORRECT - Will cause parsing errors
BadHeader: "[[ .request.headers.x-api-key ]]"
```

### Conditional Examples

```yaml
# Simple conditional
Value: |
  [[ if eq .request.method "POST" ]]
    create
  [[ else ]]
    read
  [[ end ]]

# Multiple conditions
Status: |
  [[ if eq .response.status 200 ]]
    success
  [[ else if eq .response.status 400 ]]
    client_error
  [[ else if eq .response.status 500 ]]
    server_error
  [[ else ]]
    unknown
  [[ end ]]

# Header-based conditions
Auth: |
  [[ if eq (index .request.headers "x-api-key") "secret" ]]
    authorized
  [[ else ]]
    unauthorized
  [[ end ]]
```

### Array Processing

```yaml
# Processing arrays in response
Results: |
  [
    [[ $items := .response.body.items ]]
    [[ range $index, $item := $items ]]
      [[ if $index ]], [[ end ]]
      {
        "id": "[[ $item.id ]]",
        "value": "[[ $item.value ]]",
        "index": [[ $index ]]
      }
    [[ end ]]
  ]
```

## Error Handling

### Template Errors
- Template parsing errors akan dicatat ke log
- Invalid templates akan diabaikan
- Processing akan tetap berlanjut meskipun ada template error

### Missing Data
- Missing variables akan menghasilkan `<no value>`
- Gunakan conditional checks untuk memvalidasi data

```yaml
# Safe access pattern
SafeValue: |
  [[ if .request.api.body.optional_field ]]
    [[ .request.api.body.optional_field ]]
  [[ else ]]
    default_value
  [[ end ]]
```

## Best Practices

1. **Always validate data**: Gunakan conditionals untuk check keberadaan data
2. **Use descriptive names**: Buat variable names yang jelas dan descriptive
3. **Handle errors gracefully**: Selalu provide fallback values
4. **Test thoroughly**: Test semua kondisi dan edge cases
5. **Log operations**: Monitor logs untuk debug dan troubleshooting

## Debugging

### Log Monitoring
```bash
# Monitor plugin logs
docker-compose logs -f traefik | grep modifier

# Look for patterns like:
# Set header Authorization: Bearer token (was: old-token)
# Added header X-Request-ID: req_123456789
# Modified request body: {...}
# Modified response body: {...}
```

### Common Issues
1. **Template parsing errors**: Check delimiter usage dan syntax
2. **Missing values**: Validate data availability dengan conditionals
3. **Header access**: Gunakan `index` function untuk special characters
4. **JSON syntax**: Pastikan valid JSON dalam body modifications