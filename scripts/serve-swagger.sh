#!/bin/bash

# Script to serve Swagger documentation
# This works around Go 1.20 compatibility issues with newer swagger dependencies

set -e

echo "Generating Swagger documentation..."
~/go/bin/swag init -g cmd/local/main.go -o docs

echo "Swagger documentation generated successfully!"
echo ""
echo "To view the documentation:"
echo "1. Start the API server: make run-local"
echo "2. Open your browser to: http://localhost:8080/swagger/index.html"
echo ""
echo "Note: The swagger endpoint is currently commented out due to Go 1.20 compatibility."
echo "To enable it, uncomment the swagger imports and route in internal/routes/routes.go"
echo "and the docs import in cmd/local/main.go, but you'll need Go 1.22+ to build."
echo ""
echo "Alternatively, you can view the generated documentation files:"
echo "- JSON: docs/swagger.json"
echo "- YAML: docs/swagger.yaml"
echo "- Go docs: docs/docs.go"