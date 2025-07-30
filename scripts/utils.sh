#!/bin/bash

# Utility functions for project initialization
# This file contains reusable functions that can be sourced by other scripts

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Naming convention utilities
# Convert kebab-case to Title Case
kebab_to_title() {
    local input="$1"
    # Replace hyphens with spaces and capitalize each word
    echo "$input" | sed 's/-/ /g' | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) tolower(substr($i,2))}1'
}

# Convert Title Case to kebab-case
title_to_kebab() {
    local input="$1"
    # Convert to lowercase and replace spaces with hyphens
    echo "$input" | tr '[:upper:]' '[:lower:]' | sed 's/ /-/g'
}

# Print functions
print_header() {
    echo ""
    echo -e "${PURPLE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${PURPLE}║                                                            ║${NC}"
    echo -e "${PURPLE}║               🚀 Go Microservice Generator 🚀              ║${NC}"
    echo -e "${PURPLE}║                                                            ║${NC}"
    echo -e "${PURPLE}║          Initialize your Go microservice project          ║${NC}"
    echo -e "${PURPLE}║                                                            ║${NC}"
    echo -e "${PURPLE}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

print_step() {
    echo ""
    echo -e "${CYAN}📋 $1${NC}"
    echo "────────────────────────────────────────"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}