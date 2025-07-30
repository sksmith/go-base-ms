#!/bin/bash

# OpenAPI spec merger script
# Merges base.yaml, standard.yaml, and application.yaml into a single openapi.yaml

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
API_DIR="$PROJECT_ROOT/api/openapi"
OUTPUT_FILE="$PROJECT_ROOT/api/openapi.yaml"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Merging OpenAPI specifications...${NC}"

# Check if required files exist
if [[ ! -f "$API_DIR/base.yaml" ]]; then
    echo -e "${RED}Error: base.yaml not found at $API_DIR/base.yaml${NC}"
    exit 1
fi

if [[ ! -f "$API_DIR/standard.yaml" ]]; then
    echo -e "${RED}Error: standard.yaml not found at $API_DIR/standard.yaml${NC}"
    exit 1
fi

if [[ ! -f "$API_DIR/application.yaml" ]]; then
    echo -e "${RED}Error: application.yaml not found at $API_DIR/application.yaml${NC}"
    exit 1
fi

# Check if yq is available
if ! command -v yq &> /dev/null; then
    echo -e "${YELLOW}Warning: yq not found. Installing via Go...${NC}"
    go install github.com/mikefarah/yq/v4@latest
    # Add Go bin to PATH
    export PATH=$PATH:$(go env GOPATH)/bin
    if ! command -v yq &> /dev/null; then
        echo -e "${RED}Error: Failed to install yq. Please install manually:${NC}"
        echo "  brew install yq  # macOS"
        echo "  Or download from: https://github.com/mikefarah/yq/releases"
        exit 1
    fi
fi

# Create output directory if it doesn't exist
mkdir -p "$(dirname "$OUTPUT_FILE")"

# Start with base spec
cp "$API_DIR/base.yaml" "$OUTPUT_FILE"

echo -e "${GREEN}Merging standard endpoints...${NC}"
# Merge standard paths
yq eval-all 'select(fileIndex == 0) * select(fileIndex == 1)' "$OUTPUT_FILE" "$API_DIR/standard.yaml" > "${OUTPUT_FILE}.tmp"
mv "${OUTPUT_FILE}.tmp" "$OUTPUT_FILE"

echo -e "${GREEN}Merging application endpoints...${NC}"
# Merge application paths
yq eval-all 'select(fileIndex == 0) * select(fileIndex == 1)' "$OUTPUT_FILE" "$API_DIR/application.yaml" > "${OUTPUT_FILE}.tmp"
mv "${OUTPUT_FILE}.tmp" "$OUTPUT_FILE"

echo -e "${GREEN}OpenAPI specification merged successfully!${NC}"
echo -e "${GREEN}Output: $OUTPUT_FILE${NC}"

# Validate the merged spec
echo -e "${GREEN}Validating merged specification...${NC}"
if yq eval '.' "$OUTPUT_FILE" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Validation successful${NC}"
else
    echo -e "${RED}❌ Validation failed${NC}"
    exit 1
fi

echo -e "${GREEN}🎉 OpenAPI merge completed successfully!${NC}"