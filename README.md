# Traefik Modifier Plugin

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   REQUEST API   │───▶│ MODIFIER REQUEST│───▶│     SERVICE     │───▶│MODIFIER RESPONSE│───▶│  RESPONSE API   │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘

```

A powerful Traefik middleware plugin that allows you to modify HTTP requests and responses using Go templates. This plugin supports query parameter transformation, request body modification, and response body masking with dynamic context injection.

## Features

- **Query Parameter Transformation**: Transform query parameters using templates
- **Request Body Modification**: Modify request bodies with template-based transformations
- **Response Body Masking**: Transform response bodies based on status codes
- **Context Injection**: Access dynamic context variables like timestamps in templates
- **Template Engine**: Uses Go templates with custom function maps
- **JSON Formatting**: Automatic JSON formatting (minified output)

## Installation

### Using Docker Compose

Add the plugin to your `docker-compose.yml`:

```yaml
version: '3.8'
services:
  traefik:
    image: traefik:latest
    command:
      - "--experimental.plugins.modifier.modulename=github.com/hukumonline-com/traefik-modifier"
      - "--experimental.plugins.modifier.version=v1.0.0"
    # ... other configuration
```

### Plugin Configuration

Configure the plugin in your Traefik dynamic configuration:

```yaml
http:
  middlewares:
    my-modifier:
      plugin:
        modifier:
          Query:
            transform:
              new_param: "[[ .request.query.old_param ]]_[[ .context.unixtime ]]"
          Request: |
            {
              "question": "[[ .request.body.ask ]]",
              "timestamp": [[ .context.unixtime ]]
            }
          Response:
            "200": |
              {
                "id": [[ .response.body.id ]],
                "data": "[[ .request.body.question ]]",
                "timestamp": [[ .context.unixtime ]]
              }
```

## Configuration Options

### Query Configuration

Transform query parameters using templates:

```yaml
Query:
  transform:
    question_id: "[[ .request.query.ask_id ]]_[[ .context.unixtime ]]"
    user_id: "[[ .request.query.uid ]]"
```

### Request Configuration

Modify request bodies sent to upstream services:

```yaml
Request: |
  {
    "question": "[[ .request.body.ask ]]",
    "timestamp": [[ .context.unixtime ]],
    "user": "[[ .request.body.user ]]"
  }
```

### Response Configuration

Transform response bodies based on HTTP status codes:

```yaml
Response:
  "200": |
    {
      "success": true,
      "data": [[ .response.body ]],
      "timestamp": [[ .context.unixtime ]]
    }
  "404": |
    {
      "error": "Resource not found",
      "timestamp": [[ .context.unixtime ]]
    }
```

## Template Context

The plugin provides several context variables accessible in templates:

### Request Context

- `.request.body` - Request body (parsed as JSON)
- `.request.query` - Query parameters as key-value pairs
- `.request.header` - Request headers
- `.request.method` - HTTP method
- `.request.path` - Request path

### Response Context

- `.response.body` - Response body (parsed as JSON)

### Dynamic Context

- `.context.unixtime` - Current Unix timestamp (updated for each request)

## Template Functions

The plugin includes custom template functions via `pkg.SimpleFuncMap()`:

- Standard Go template functions
- Custom utility functions for data manipulation

## Examples

### Basic Query Parameter Transformation

```yaml
middlewares:
  query-transformer:
    plugin:
      modifier:
        Query:
          transform:
            # Transform ask_id to question_id with timestamp
            question_id: "[[ .request.query.ask_id ]]_[[ .context.unixtime ]]"
```

### Request Body Modification

```yaml
middlewares:
  request-modifier:
    plugin:
      modifier:
        Request: |
          {
            "question": "[[ .request.body.ask ]]",
            "timestamp": [[ .context.unixtime ]],
            "source": "traefik-modifier"
          }
```

### Response Data Transformation

```yaml
middlewares:
  response-modifier:
    plugin:
      modifier:
        Response:
          "200": |
            {
              "id": [[ .response.body.id ]],
              "answer": "[[ .request.body.ask ]]",
              "timestamp": [[ .context.unixtime ]],
              "datas": [
                [[ range .response.body.data_array_of_maps ]]
                {
                  "id": "[[ .key1 ]]",
                  "value": "[[ .key2 ]]"
                }[[ if not (isLast .) ]],[[ end ]]
                [[ end ]]
              ]
            }
```

### Complete Configuration Example

```yaml
http:
  routers:
    chat-service:
      rule: "Host(`chat.localhost`)"
      entryPoints:
        - web
      middlewares:
        - chat-modifier
      service: chat-service

  middlewares:
    chat-modifier:
      plugin:
        modifier:
          Query: 
            transform:
              question_id: "[[ .request.query.ask_id ]]_[[ .context.unixtime ]]"
          Request: |
            {
              "question": "[[ .request.body.ask ]]",
              "timestamp": [[ .context.unixtime ]]
            }
          Response:
            "200": |
              {
                "id": [[ .response.body.id ]],
                "answer": "[[ .request.body.ask ]]",
                "timestamp": [[ .context.unixtime ]],
                "datas": [
                  [[ range .response.body.data_array_of_maps ]]
                  {
                    "id": "[[ .key1 ]]",
                    "value": "[[ .key2 ]]"
                  }[[ if not (isLast .) ]],[[ end ]]
                  [[ end ]]
                ]
              }

  services:
    chat-service:
      loadBalancer:
        servers:
          - url: "http://chat-service:3000"
```

### Template Processing

- Uses Go's `text/template` package
- Custom delimiters: `[[` and `]]`
- Context injection for dynamic values
- JSON parsing and formatting

## Development

### Project Structure

```
traefik-modifier/
├── modifier.go          # Main plugin logic
├── body.go             # Body modification handlers
├── query.go            # Query parameter handlers
├── pkg/
│   └── simplefunc.go   # Template functions
├── traefik/
│   ├── traefik.yaml    # Traefik configuration
│   └── dynamic.yaml    # Dynamic configuration
├── docker/
│   ├── Dockerfile      # Plugin container
│   └── test-backend.js # Test backend service
├── docker-compose.yml  # Development environment
├── go.mod             # Go module definition
└── Makefile          # Build automation
```

### Building

```bash
# Build the plugin
make build

# Run tests
make test

# Start development environment
docker-compose up
```

### Testing

The plugin includes a test backend service for development and testing:

```bash
# Start the development environment
docker-compose up

# Test query transformation
curl "http://chat.localhost/test?ask_id=123"

# Test request/response modification
curl -X POST http://chat.localhost/test \
  -H "Content-Type: application/json" \
  -d '{"ask": "What is the weather?"}'
```

## Error Handling

The plugin includes comprehensive error handling:

- Template parsing errors are logged
- JSON parsing failures fall back to raw string handling
- Network errors are propagated appropriately
- Configuration validation on startup

## Performance Considerations

- Templates are parsed once during initialization
- JSON parsing/formatting is optimized for common use cases
- Context is lightweight and created per request
- Memory usage is minimized through efficient buffering

## Security Considerations

- Input validation on all template data
- Safe template execution with controlled context
- No arbitrary code execution capabilities
- Header manipulation is controlled and logged

## Troubleshooting

### Common Issues

1. **Template Execution Errors**:
   ```
   Failed to execute query template: template: query:1:39: executing "query" at <.context.unixtime>: can't evaluate field unixtime in type interface {}
   ```
   - Ensure context types are properly defined
   - Check template syntax and variable references

2. **JSON Parsing Errors**:
   - Verify input data is valid JSON
   - Check template output produces valid JSON

3. **Configuration Errors**:
   - Validate YAML syntax in dynamic configuration
   - Ensure all required fields are present

### Debug Mode

Enable debug logging by setting log level in Traefik configuration:

```yaml
log:
  level: DEBUG
```

## License

This plugin is released under the MIT License. See LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## Support

For issues and questions:
- Check the troubleshooting section
- Review Traefik plugin documentation
- Submit issues on the project repository