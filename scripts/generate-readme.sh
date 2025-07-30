#!/bin/bash

# README generator script
# Processes the README template with conditional sections based on project configuration

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Default values (set by init-project.sh)
PROJECT_NAME="${PROJECT_NAME:-go-base-ms}"
PROJECT_DESCRIPTION="${PROJECT_DESCRIPTION:-A Go microservice built with go-base-ms template}"
USE_POSTGRES="${USE_POSTGRES:-true}"
USE_KAFKA="${USE_KAFKA:-true}"
USE_SCHEMA_REGISTRY="${USE_SCHEMA_REGISTRY:-true}"

# Input and output files
TEMPLATE_FILE="${1:-templates/README.md.template}"
OUTPUT_FILE="${2:-README.md}"

echo -e "${GREEN}Generating README from template...${NC}"
echo "Template: $TEMPLATE_FILE"
echo "Output: $OUTPUT_FILE"
echo "Configuration:"
echo "  PROJECT_NAME: $PROJECT_NAME"
echo "  PROJECT_DESCRIPTION: $PROJECT_DESCRIPTION"
echo "  USE_POSTGRES: $USE_POSTGRES"
echo "  USE_KAFKA: $USE_KAFKA"
echo "  USE_SCHEMA_REGISTRY: $USE_SCHEMA_REGISTRY"

# Check if template exists
if [[ ! -f "$TEMPLATE_FILE" ]]; then
    echo -e "${RED}Error: Template file not found: $TEMPLATE_FILE${NC}"
    exit 1
fi

# Function to process conditional sections
process_template() {
    local content="$1"
    
    # Replace placeholders
    content="${content//PROJECT_NAME/$PROJECT_NAME}"
    content="${content//PROJECT_DESCRIPTION/$PROJECT_DESCRIPTION}"
    
    # Process conditional sections for PostgreSQL
    if [[ "$USE_POSTGRES" == "true" ]]; then
        # Keep content between {{#USE_POSTGRES}} and {{/USE_POSTGRES}}
        content=$(echo "$content" | sed 's/{{#USE_POSTGRES}}//g; s/{{\/USE_POSTGRES}}//g')
    else
        # Remove content between {{#USE_POSTGRES}} and {{/USE_POSTGRES}}
        content=$(echo "$content" | sed '/{{#USE_POSTGRES}}/,/{{\/USE_POSTGRES}}/d')
    fi
    
    # Process conditional sections for Kafka
    if [[ "$USE_KAFKA" == "true" ]]; then
        # Keep content between {{#USE_KAFKA}} and {{/USE_KAFKA}}
        content=$(echo "$content" | sed 's/{{#USE_KAFKA}}//g; s/{{\/USE_KAFKA}}//g')
    else
        # Remove content between {{#USE_KAFKA}} and {{/USE_KAFKA}}
        content=$(echo "$content" | sed '/{{#USE_KAFKA}}/,/{{\/USE_KAFKA}}/d')
    fi
    
    # Process conditional sections for Schema Registry
    if [[ "$USE_SCHEMA_REGISTRY" == "true" ]]; then
        # Keep content between {{#USE_SCHEMA_REGISTRY}} and {{/USE_SCHEMA_REGISTRY}}
        content=$(echo "$content" | sed 's/{{#USE_SCHEMA_REGISTRY}}//g; s/{{\/USE_SCHEMA_REGISTRY}}//g')
    else
        # Remove content between {{#USE_SCHEMA_REGISTRY}} and {{/USE_SCHEMA_REGISTRY}}
        content=$(echo "$content" | sed '/{{#USE_SCHEMA_REGISTRY}}/,/{{\/USE_SCHEMA_REGISTRY}}/d')
    fi
    
    # Clean up multiple consecutive blank lines (reduce to maximum 2)
    content=$(echo "$content" | sed '/^$/N;/^\n$/N;/^\n\n$/d')
    
    echo "$content"
}

# Read template file
template_content=$(cat "$TEMPLATE_FILE")

# Process the template
processed_content=$(process_template "$template_content")

# Write output file
echo "$processed_content" > "$OUTPUT_FILE"

echo -e "${GREEN}✅ README generated successfully: $OUTPUT_FILE${NC}"

# Show a summary of what was included
echo ""
echo -e "${GREEN}Generated README includes:${NC}"
echo "  ✅ Basic project structure and endpoints"
echo "  ✅ OpenAPI documentation system"
echo "  ✅ Development and deployment instructions"
if [[ "$USE_POSTGRES" == "true" ]]; then
    echo "  ✅ PostgreSQL database configuration"
else
    echo "  ❌ PostgreSQL sections removed"
fi
if [[ "$USE_KAFKA" == "true" ]]; then
    echo "  ✅ Kafka integration documentation"
    if [[ "$USE_SCHEMA_REGISTRY" == "true" ]]; then
        echo "  ✅ Schema Registry and Avro documentation"
    else
        echo "  ❌ Schema Registry sections removed"
    fi
else
    echo "  ❌ Kafka sections removed"
fi