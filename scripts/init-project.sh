#!/bin/bash

# Project initialization script for go-base-ms template
# This script customizes the template based on user preferences

set -e

# Source utility functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils.sh"

# Global variables
PROJECT_NAME=""           # Technical name (kebab-case): go-base-ms
PROJECT_DISPLAY_NAME=""   # Display name (Title Case): Go Base Microservice  
PROJECT_DESCRIPTION=""
GITHUB_USERNAME=""
GITHUB_REPO=""
MODULE_NAME=""
USE_POSTGRES=false
USE_KAFKA=false
USE_SCHEMA_REGISTRY=false
CREATE_GITHUB_REPO=false
PRIVATE_REPO=false


# Check prerequisites
check_prerequisites() {
    print_step "Checking prerequisites"
    
    # Check if we're in the go-base-ms directory
    if [[ ! -f ".goreleaser.yaml" || ! -d "cmd/server" ]]; then
        print_error "This script must be run from the go-base-ms project root directory"
        exit 1
    fi
    
    # Check required tools
    local missing_tools=()
    
    if ! command -v go &> /dev/null; then
        missing_tools+=("go")
    fi
    
    if ! command -v git &> /dev/null; then
        missing_tools+=("git")
    fi
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        print_error "Missing required tools: ${missing_tools[*]}"
        echo "Please install the missing tools and try again."
        exit 1
    fi
    
    # Check Go version
    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    local required_version="1.21"
    if ! printf '%s\n%s\n' "$required_version" "$go_version" | sort -C -V; then
        print_error "Go version $go_version is too old. Required: $required_version or higher"
        exit 1
    fi
    
    print_success "All prerequisites met"
}

# Collect user input
collect_user_input() {
    print_step "Project Configuration"
    
    # Project name (technical name)
    while [[ -z "$PROJECT_NAME" ]]; do
        echo -n -e "${BLUE}Enter project name (technical name, e.g., my-awesome-service): ${NC}"
        read -r PROJECT_NAME
        if [[ -z "$PROJECT_NAME" ]]; then
            print_warning "Project name cannot be empty"
        elif [[ ! "$PROJECT_NAME" =~ ^[a-z0-9-]+$ ]]; then
            print_warning "Project name should only contain lowercase letters, numbers, and hyphens"
            PROJECT_NAME=""
        fi
    done
    
    # Auto-generate display name from project name
    DEFAULT_DISPLAY_NAME=$(kebab_to_title "$PROJECT_NAME")
    echo -n -e "${BLUE}Enter display name (default: $DEFAULT_DISPLAY_NAME): ${NC}"
    read -r PROJECT_DISPLAY_NAME
    if [[ -z "$PROJECT_DISPLAY_NAME" ]]; then
        PROJECT_DISPLAY_NAME="$DEFAULT_DISPLAY_NAME"
    fi
    
    # Project description
    echo -n -e "${BLUE}Enter project description (optional): ${NC}"
    read -r PROJECT_DESCRIPTION
    if [[ -z "$PROJECT_DESCRIPTION" ]]; then
        PROJECT_DESCRIPTION="A Go microservice built with go-base-ms template"
    fi
    
    # GitHub configuration
    echo ""
    echo -e "${BLUE}GitHub Configuration:${NC}"
    echo -n -e "${BLUE}Enter your GitHub username: ${NC}"
    read -r GITHUB_USERNAME
    
    if [[ -n "$GITHUB_USERNAME" ]]; then
        echo -n -e "${BLUE}Create GitHub repository? (y/N): ${NC}"
        read -r create_repo
        if [[ "$create_repo" =~ ^[Yy]$ ]]; then
            CREATE_GITHUB_REPO=true
            
            echo -n -e "${BLUE}Repository name (default: $PROJECT_NAME): ${NC}"
            read -r GITHUB_REPO
            if [[ -z "$GITHUB_REPO" ]]; then
                GITHUB_REPO="$PROJECT_NAME"
            fi
            
            echo -n -e "${BLUE}Make repository private? (y/N): ${NC}"
            read -r private_repo
            if [[ "$private_repo" =~ ^[Yy]$ ]]; then
                PRIVATE_REPO=true
            fi
        fi
        MODULE_NAME="github.com/$GITHUB_USERNAME/$GITHUB_REPO"
    else
        echo -n -e "${BLUE}Enter Go module name (e.g., github.com/user/repo): ${NC}"
        read -r MODULE_NAME
        if [[ -z "$MODULE_NAME" ]]; then
            MODULE_NAME="github.com/user/$PROJECT_NAME"
        fi
    fi
    
    # Dependencies
    print_step "Dependency Configuration"
    
    echo -e "${BLUE}Select the dependencies you need:${NC}"
    
    echo -n -e "${BLUE}Include PostgreSQL database support? (Y/n): ${NC}"
    read -r use_postgres
    if [[ ! "$use_postgres" =~ ^[Nn]$ ]]; then
        USE_POSTGRES=true
        print_success "PostgreSQL support enabled"
    else
        print_warning "PostgreSQL support disabled"
    fi
    
    echo -n -e "${BLUE}Include Kafka message broker support? (Y/n): ${NC}"
    read -r use_kafka
    if [[ ! "$use_kafka" =~ ^[Nn]$ ]]; then
        USE_KAFKA=true
        print_success "Kafka support enabled"
        
        if [[ "$USE_KAFKA" == true ]]; then
            echo -n -e "${BLUE}Include Schema Registry (Avro) support? (Y/n): ${NC}"
            read -r use_schema_registry
            if [[ ! "$use_schema_registry" =~ ^[Nn]$ ]]; then
                USE_SCHEMA_REGISTRY=true
                print_success "Schema Registry support enabled"
            else
                print_warning "Schema Registry support disabled"
            fi
        fi
    else
        print_warning "Kafka support disabled"
    fi
    
    # Confirmation
    print_step "Configuration Summary"
    echo -e "${CYAN}Technical Name:${NC} $PROJECT_NAME"
    echo -e "${CYAN}Display Name:${NC} $PROJECT_DISPLAY_NAME"
    echo -e "${CYAN}Description:${NC} $PROJECT_DESCRIPTION"
    echo -e "${CYAN}Module Name:${NC} $MODULE_NAME"
    echo -e "${CYAN}PostgreSQL:${NC} $([ "$USE_POSTGRES" == true ] && echo "✅ Enabled" || echo "❌ Disabled")"
    echo -e "${CYAN}Kafka:${NC} $([ "$USE_KAFKA" == true ] && echo "✅ Enabled" || echo "❌ Disabled")"
    echo -e "${CYAN}Schema Registry:${NC} $([ "$USE_SCHEMA_REGISTRY" == true ] && echo "✅ Enabled" || echo "❌ Disabled")"
    echo -e "${CYAN}GitHub Repo:${NC} $([ "$CREATE_GITHUB_REPO" == true ] && echo "✅ Will create $GITHUB_USERNAME/$GITHUB_REPO" || echo "❌ Skip")"
    
    echo ""
    echo -n -e "${BLUE}Proceed with this configuration? (Y/n): ${NC}"
    read -r confirm
    if [[ "$confirm" =~ ^[Nn]$ ]]; then
        echo -e "${YELLOW}Configuration cancelled. Exiting...${NC}"
        exit 0
    fi
}

# Update go.mod with new module name and replace all hardcoded references
update_go_mod() {
    print_step "Updating Go module and replacing template references"
    
    # Update go.mod
    sed -i.bak "s|module github.com/dks0523168/go-base-ms|module $MODULE_NAME|g" go.mod
    rm go.mod.bak
    
    # Update all import statements in Go files
    find . -name "*.go" -type f -exec sed -i.bak "s|github.com/dks0523168/go-base-ms|$MODULE_NAME|g" {} \;
    find . -name "*.go.bak" -delete
    
    # Replace all instances of hardcoded GitHub username
    find . -type f \( -name "*.go" -o -name "*.yaml" -o -name "*.yml" -o -name "*.md" -o -name "*.json" -o -name "Dockerfile*" \) \
        -exec sed -i.bak "s/DKS0523168/$GITHUB_USERNAME/g" {} \;
    find . -name "*.bak" -delete
    
    # Replace all instances of technical project name
    find . -type f \( -name "*.go" -o -name "*.yaml" -o -name "*.yml" -o -name "*.md" -o -name "*.json" -o -name "Dockerfile*" -o -name "Makefile" \) \
        -exec sed -i.bak "s/go-base-ms/$PROJECT_NAME/g" {} \;
    find . -name "*.bak" -delete
    
    # Replace instances of display name (Go Base Microservice)
    find . -type f \( -name "*.go" -o -name "*.yaml" -o -name "*.yml" -o -name "*.md" -o -name "*.json" \) \
        -exec sed -i.bak "s/Go Base Microservice/$PROJECT_DISPLAY_NAME/g" {} \;
    find . -name "*.bak" -delete
    
    print_success "Go module updated to $MODULE_NAME"
    print_success "Replaced GitHub username: DKS0523168 → $GITHUB_USERNAME"
    print_success "Replaced project name: go-base-ms → $PROJECT_NAME"
    print_success "Replaced display name: Go Base Microservice → $PROJECT_DISPLAY_NAME"
}

# Remove PostgreSQL dependencies
remove_postgres_support() {
    print_step "Removing PostgreSQL support"
    
    # Remove PostgreSQL dependency from go.mod
    go mod edit -droprequire github.com/lib/pq
    
    # Remove database files
    rm -rf internal/db/
    
    # Update main.go to remove database initialization
    sed -i.bak '/database/d; /db\./d; /DB/d' cmd/server/main.go
    sed -i.bak '/github.com\/.*\/internal\/db/d' cmd/server/main.go
    rm cmd/server/main.go.bak
    
    # Update config to remove database config
    sed -i.bak '/Database.*DatabaseConfig/d; /type DatabaseConfig/,/^}/d' internal/config/config.go
    sed -i.bak '/Database:/,/},/d' internal/config/config.go
    sed -i.bak '/DB_/d' internal/config/config.go
    rm internal/config/config.go.bak
    
    # Update health checker
    if [[ "$USE_KAFKA" == true ]]; then
        sed -i.bak 's/New(database, kafkaClient)/New(kafkaClient)/g' cmd/server/main.go
        sed -i.bak 's/func New(db Checker, kafka Checker)/func New(kafka Checker)/g' internal/health/health.go
        sed -i.bak 's/"database": db,/"kafka": kafka,/g' internal/health/health.go
        sed -i.bak '/database.*db/d' internal/health/health.go
    else
        sed -i.bak '/healthChecker := health.New/d' cmd/server/main.go
        sed -i.bak 's/NewRouter(log, healthChecker)/NewRouter(log, health.New())/g' cmd/server/main.go
        sed -i.bak 's/func New(db Checker, kafka Checker)/func New()/g' internal/health/health.go
        sed -i.bak '/checks.*map.*Checker/,/}/d' internal/health/health.go
        sed -i.bak 's/h.checks\[name\], checker := range h.checks/range []string{}/g' internal/health/health.go
    fi
    rm cmd/server/main.go.bak internal/health/health.go.bak 2>/dev/null || true
    
    # Update docker-compose and Makefile
    sed -i.bak '/postgres/,/^$/d' docker-compose.yml
    sed -i.bak '/postgres-dev/d; /DB_/d' Makefile
    rm docker-compose.yml.bak Makefile.bak
    
    print_success "PostgreSQL support removed"
}

# Remove Kafka dependencies
remove_kafka_support() {
    print_step "Removing Kafka support"
    
    # Remove Kafka dependencies from go.mod
    go mod edit -droprequire github.com/confluentinc/confluent-kafka-go/v2
    
    # Remove kafka files
    rm -rf internal/kafka/
    
    # Update main.go to remove kafka initialization
    sed -i.bak '/kafka/d; /Kafka/d; /KAFKA/d' cmd/server/main.go
    sed -i.bak '/github.com\/.*\/internal\/kafka/d' cmd/server/main.go
    rm cmd/server/main.go.bak
    
    # Update config to remove kafka config
    sed -i.bak '/Kafka.*KafkaConfig/d; /SchemaRegistry.*SchemaRegistryConfig/d' internal/config/config.go
    sed -i.bak '/type KafkaConfig/,/^}/d; /type SchemaRegistryConfig/,/^}/d' internal/config/config.go
    sed -i.bak '/Kafka:/,/},/d; /SchemaRegistry:/,/},/d' internal/config/config.go
    sed -i.bak '/KAFKA_/d; /SCHEMA_REGISTRY_/d' internal/config/config.go
    rm internal/config/config.go.bak
    
    # Update health checker
    if [[ "$USE_POSTGRES" == true ]]; then
        sed -i.bak 's/New(database, kafkaClient)/New(database)/g' cmd/server/main.go
        sed -i.bak 's/func New(db Checker, kafka Checker)/func New(db Checker)/g' internal/health/health.go
        sed -i.bak 's/"kafka": kafka,/"database": db,/g' internal/health/health.go
    else
        sed -i.bak '/healthChecker := health.New/d' cmd/server/main.go
        sed -i.bak 's/NewRouter(log, healthChecker)/NewRouter(log, health.New())/g' cmd/server/main.go
        sed -i.bak 's/func New(kafka Checker)/func New()/g' internal/health/health.go
        sed -i.bak '/checks.*map.*Checker/,/}/d' internal/health/health.go
    fi
    sed -i.bak '/kafka.*kafka/d' internal/health/health.go
    rm cmd/server/main.go.bak internal/health/health.go.bak 2>/dev/null || true
    
    # Update docker-compose and Makefile
    sed -i.bak '/kafka/,/^$/d; /zookeeper/,/^$/d; /schema-registry/,/^$/d' docker-compose.yml
    sed -i.bak '/kafka-dev/d; /zookeeper-dev/d; /schema-registry-dev/d; /KAFKA_/d; /SCHEMA_REGISTRY_/d' Makefile
    rm docker-compose.yml.bak Makefile.bak
    
    print_success "Kafka support removed"
}

# Remove Schema Registry but keep Kafka
remove_schema_registry_support() {
    print_step "Removing Schema Registry support"
    
    # Replace kafka.go with simplified version
    if [[ -f "templates/kafka-simple.go.template" ]]; then
        cp templates/kafka-simple.go.template internal/kafka/kafka.go
        sed -i.bak "s/MODULE_NAME/$MODULE_NAME/g" internal/kafka/kafka.go
        sed -i.bak "s/PROJECT_NAME/$PROJECT_NAME/g" internal/kafka/kafka.go
        rm internal/kafka/kafka.go.bak
    else
        # Fallback: remove schema registry imports and functionality
        sed -i.bak '/schemaregistry/d; /serde/d; /avro/d' internal/kafka/kafka.go
        sed -i.bak '/SchemaRegistry/,/^}/d; /Avro/,/^}/d' internal/kafka/kafka.go
        sed -i.bak '/schemaRegistry/d; /avroSerializer/d; /avroDeserializer/d' internal/kafka/kafka.go
        rm internal/kafka/kafka.go.bak
    fi
    
    # Update config to remove schema registry config
    sed -i.bak '/SchemaRegistry.*SchemaRegistryConfig/d' internal/config/config.go
    sed -i.bak '/type SchemaRegistryConfig/,/^}/d' internal/config/config.go
    sed -i.bak '/SchemaRegistry:/,/},/d' internal/config/config.go
    sed -i.bak '/SCHEMA_REGISTRY_/d' internal/config/config.go
    rm internal/config/config.go.bak
    
    # Update main.go
    sed -i.bak 's/kafka.New(cfg.Kafka, cfg.SchemaRegistry, log)/kafka.New(cfg.Kafka, log)/g' cmd/server/main.go
    rm cmd/server/main.go.bak
    
    print_success "Schema Registry support removed"
}

# Update project files with new name and configuration
update_project_files() {
    print_step "Updating project files"
    
    # Generate README from template
    export PROJECT_NAME="$PROJECT_NAME"
    export PROJECT_DISPLAY_NAME="$PROJECT_DISPLAY_NAME"
    export PROJECT_DESCRIPTION="$PROJECT_DESCRIPTION"
    export USE_POSTGRES="$USE_POSTGRES"
    export USE_KAFKA="$USE_KAFKA"
    export USE_SCHEMA_REGISTRY="$USE_SCHEMA_REGISTRY"
    ./scripts/generate-readme.sh
    
    # Regenerate OpenAPI spec after replacements
    if [[ -f "api/openapi/base.yaml" ]]; then
        # Regenerate merged spec with updated names
        ./scripts/merge-openapi.sh >/dev/null 2>&1 || echo "Warning: Could not regenerate OpenAPI spec"
    fi
    
    print_success "Project files updated with $PROJECT_NAME"
}

# Initialize git repository
init_git_repository() {
    print_step "Initializing Git repository"
    
    # Remove existing git history if it exists
    if [[ -d ".git" ]]; then
        rm -rf .git
    fi
    
    # Initialize new git repository
    git init
    git add .
    git commit -m "feat: initial commit of $PROJECT_DISPLAY_NAME ($PROJECT_NAME)
    
Generated from go-base-ms template with:
- PostgreSQL: $([ "$USE_POSTGRES" == true ] && echo "enabled" || echo "disabled")
- Kafka: $([ "$USE_KAFKA" == true ] && echo "enabled" || echo "disabled")
- Schema Registry: $([ "$USE_SCHEMA_REGISTRY" == true ] && echo "enabled" || echo "disabled")

🤖 Generated with go-base-ms template"
    
    print_success "Git repository initialized"
}

# Create GitHub repository
create_github_repository() {
    if [[ "$CREATE_GITHUB_REPO" != true ]]; then
        return
    fi
    
    print_step "Creating GitHub repository"
    
    # Check if gh CLI is installed
    if ! command -v gh &> /dev/null; then
        print_warning "GitHub CLI (gh) not found. Please install it and authenticate:"
        echo "  brew install gh  # macOS"
        echo "  gh auth login"
        echo ""
        echo "Then create the repository manually:"
        echo "  gh repo create $GITHUB_USERNAME/$GITHUB_REPO --public --description \"$PROJECT_DESCRIPTION\""
        return
    fi
    
    # Check if user is authenticated
    if ! gh auth status &> /dev/null; then
        print_warning "GitHub CLI not authenticated. Please run:"
        echo "  gh auth login"
        return
    fi
    
    # Create repository
    local visibility_flag="--public"
    if [[ "$PRIVATE_REPO" == true ]]; then
        visibility_flag="--private"
    fi
    
    echo "Creating GitHub repository: $GITHUB_USERNAME/$GITHUB_REPO"
    if gh repo create "$GITHUB_USERNAME/$GITHUB_REPO" $visibility_flag --description "$PROJECT_DESCRIPTION" --clone=false; then
        print_success "GitHub repository created: https://github.com/$GITHUB_USERNAME/$GITHUB_REPO"
        
        # Add remote and push
        git remote add origin "https://github.com/$GITHUB_USERNAME/$GITHUB_REPO.git"
        git branch -M main
        git push -u origin main
        
        print_success "Code pushed to GitHub"
    else
        print_error "Failed to create GitHub repository"
    fi
}

# Clean up and finalize
finalize_project() {
    print_step "Finalizing project setup"
    
    # Clean up Go modules
    go mod tidy
    
    # Run tests to make sure everything works
    if go test ./... > /dev/null 2>&1; then
        print_success "All tests pass"
    else
        print_warning "Some tests failed. This might be expected if external services are not running."
    fi
    
    # Remove the init script and template-specific files
    rm -f scripts/init-project.sh
    rm -f scripts/init-release.sh
    rm -f scripts/generate-readme.sh
    rm -rf templates/
    
    # Update the scripts directory or remove if empty
    if [[ -d "scripts" && ! "$(ls -A scripts)" ]]; then
        rmdir scripts
    fi
    
    print_success "Project initialization completed"
}

# Display final instructions
show_final_instructions() {
    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                                                            ║${NC}"
    echo -e "${GREEN}║                    🎉 Setup Complete! 🎉                   ║${NC}"
    echo -e "${GREEN}║                                                            ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${CYAN}Your project '$PROJECT_DISPLAY_NAME' ($PROJECT_NAME) is ready!${NC}"
    echo ""
    echo -e "${BLUE}Next steps:${NC}"
    echo "1. Start development environment:"
    echo -e "   ${YELLOW}make dev-env${NC}"
    echo ""
    echo "2. Run the application:"
    echo -e "   ${YELLOW}make run-dev${NC}"
    echo ""
    echo "3. Run tests:"
    echo -e "   ${YELLOW}make test${NC}"
    echo ""
    echo "4. Create your first release:"
    echo -e "   ${YELLOW}make release-init${NC}"
    echo ""
    if [[ "$CREATE_GITHUB_REPO" == true ]]; then
        echo -e "${BLUE}GitHub Repository:${NC}"
        echo -e "   ${CYAN}https://github.com/$GITHUB_USERNAME/$GITHUB_REPO${NC}"
        echo ""
    fi
    echo -e "${BLUE}Available endpoints:${NC}"
    echo "   • Health: http://localhost:8080/health/live"
    echo "   • Version: http://localhost:8080/version"
    echo "   • API Docs: http://localhost:8080/openapi.json"
    echo ""
    echo -e "${BLUE}For help:${NC}"
    echo -e "   ${YELLOW}make help${NC}"
    echo ""
    echo -e "${GREEN}Happy coding! 🚀${NC}"
}

# Main execution
main() {
    print_header
    check_prerequisites
    collect_user_input
    
    # Update module name first
    update_go_mod
    
    # Remove unwanted dependencies
    if [[ "$USE_POSTGRES" != true ]]; then
        remove_postgres_support
    fi
    
    if [[ "$USE_KAFKA" != true ]]; then
        remove_kafka_support
    elif [[ "$USE_SCHEMA_REGISTRY" != true ]]; then
        remove_schema_registry_support
    fi
    
    # Update project files
    update_project_files
    
    # Initialize git and create GitHub repo
    init_git_repository
    create_github_repository
    
    # Finalize
    finalize_project
    show_final_instructions
}

# Run main function
main "$@"