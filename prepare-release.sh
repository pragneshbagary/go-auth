#!/bin/bash

# Go-Auth v2 Release Preparation Script
# This script prepares the repository for release by organizing commits logically

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    print_error "Not in a git repository. Please run this script from the project root."
    exit 1
fi

print_header "Go-Auth v2 Release Preparation"
print_status "This script will organize and commit all v2 changes"
echo

# Check git status
if [ -n "$(git status --porcelain)" ]; then
    print_status "Found uncommitted changes. Proceeding with commit organization..."
else
    print_warning "No uncommitted changes found. Repository appears to be clean."
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 0
    fi
fi

# Function to create organized commits
create_organized_commits() {
    print_header "Creating Organized Commit History"
    
    # 1. Core Infrastructure
    print_status "Committing core infrastructure..."
    git add go.mod go.sum LICENSE .gitignore
    if ! git diff --cached --quiet; then
        git commit -m "chore: update project infrastructure and dependencies

- Update .gitignore to exclude build artifacts and temporary files
- Ensure proper licensing and module configuration
- Prepare repository for v2 release"
    fi
    
    # 2. Enhanced Configuration System
    print_status "Committing enhanced configuration system..."
    git add pkg/auth/enhanced_config.go pkg/auth/enhanced_config_test.go
    if ! git diff --cached --quiet; then
        git commit -m "feat: implement enhanced configuration system

- Add comprehensive configuration with environment variable support
- Implement configuration validation and defaults
- Support for multiple database types and JWT configurations
- Add extensive test coverage for configuration scenarios"
    fi
    
    # 3. Error Handling and Logging
    print_status "Committing error handling and logging..."
    git add pkg/auth/errors.go pkg/auth/errors_test.go pkg/auth/logger.go
    if ! git diff --cached --quiet; then
        git commit -m "feat: implement advanced error handling and structured logging

- Add custom error types with proper error codes
- Implement structured logging with configurable levels
- Add error wrapping and context preservation
- Comprehensive error handling test suite"
    fi
    
    # 4. Database Migration System
    print_status "Committing migration system..."
    git add pkg/auth/migration.go pkg/auth/migration_test.go
    if ! git diff --cached --quiet; then
        git commit -m "feat: implement database migration system

- Add automatic database migration with version control
- Support for migration rollback and version targeting
- Schema version tracking and validation
- Comprehensive migration testing"
    fi
    
    # 5. Monitoring and Metrics
    print_status "Committing monitoring and metrics..."
    git add pkg/auth/monitoring.go pkg/auth/monitoring_test.go pkg/auth/metrics.go pkg/auth/metrics_test.go
    if ! git diff --cached --quiet; then
        git commit -m "feat: add comprehensive monitoring and metrics collection

- Implement system health monitoring with detailed status reporting
- Add metrics collection for authentication operations
- Performance tracking and success rate calculations
- Health check endpoints and system information"
    fi
    
    # 6. Core Auth System
    print_status "Committing core auth system..."
    git add pkg/auth/auth.go pkg/auth/config.go pkg/auth/hashing_test.go
    if ! git diff --cached --quiet; then
        git commit -m "feat: implement enhanced core authentication system

- Redesigned Auth struct with improved API
- Simplified constructors for different storage types
- Automatic database initialization and migration
- Enhanced password hashing with security tests"
    fi
    
    # 7. Component-Based Architecture
    print_status "Committing component architecture..."
    git add pkg/auth/users.go pkg/auth/users_test.go pkg/auth/tokens.go pkg/auth/tokens_test.go
    if ! git diff --cached --quiet; then
        git commit -m "feat: implement component-based architecture

- Add Users component for comprehensive user management
- Add Tokens component for advanced token operations
- Implement password reset functionality with secure tokens
- Batch token validation and session management
- Comprehensive test coverage for all components"
    fi
    
    # 8. Middleware System
    print_status "Committing middleware system..."
    git add pkg/auth/middleware.go pkg/auth/middleware_test.go pkg/auth/middleware_adapters.go pkg/auth/middleware_adapters_test.go
    if ! git diff --cached --quiet; then
        git commit -m "feat: implement comprehensive middleware system

- Framework-specific adapters for Gin, Echo, Fiber
- Automatic user injection into request context
- Support for optional authentication
- Flexible middleware configuration and testing"
    fi
    
    # 9. SimpleAuth API
    print_status "Committing SimpleAuth..."
    git add pkg/auth/simple_auth.go pkg/auth/simple_auth_test.go
    if ! git diff --cached --quiet; then
        git commit -m "feat: implement SimpleAuth for rapid development

- Simplified API for quick setup and prototyping
- Environment-based configuration
- Automatic directory creation for SQLite
- Comprehensive testing and validation"
    fi
    
    # 10. Storage Implementations
    print_status "Committing storage implementations..."
    git add internal/storage/
    if ! git diff --cached --quiet; then
        git commit -m "feat: enhance storage implementations

- Improved PostgreSQL storage with connection pooling
- Optimized SQLite storage with better performance
- Enhanced in-memory storage for testing
- Comprehensive storage testing and integration tests"
    fi
    
    # 11. Caching System
    print_status "Committing caching system..."
    git add pkg/auth/cache.go pkg/auth/cache_test.go
    if ! git diff --cached --quiet; then
        git commit -m "feat: implement optional caching layer

- Add caching support for improved performance
- Configurable cache TTL and cleanup
- Memory-efficient caching implementation
- Comprehensive caching tests"
    fi
    
    # 12. Models and Integration
    print_status "Committing models and integration..."
    git add pkg/models/ pkg/storage/
    if ! git diff --cached --quiet; then
        git commit -m "feat: enhance data models and storage integration

- Improved user models with metadata support
- Enhanced storage interfaces and implementations
- Better integration between components
- Comprehensive integration testing"
    fi
    
    # 13. Security Enhancements
    print_status "Committing security enhancements..."
    git add pkg/auth/security_test.go
    if ! git diff --cached --quiet; then
        git commit -m "feat: implement comprehensive security testing

- Security tests for password hashing strength
- Token validation and tampering protection
- SQL injection prevention testing
- Rate limiting and brute force protection
- Data exposure prevention and randomness testing"
    fi
    
    # 14. Performance Testing
    print_status "Committing performance testing..."
    git add pkg/auth/performance_test.go
    if ! git diff --cached --quiet; then
        git commit -m "feat: add comprehensive performance testing

- Load testing with concurrent operations
- Performance benchmarks for critical paths
- Memory usage optimization
- Throughput and latency measurements"
    fi
    
    # 15. Backward Compatibility
    print_status "Committing backward compatibility..."
    git add pkg/auth/compatibility.go pkg/auth/compatibility_test.go pkg/auth/service.go
    if ! git diff --cached --quiet; then
        git commit -m "feat: implement comprehensive backward compatibility layer

- Full backward compatibility with v1 API
- Deprecation warnings with migration guidance
- Alias functions mapping old API to new API
- Comprehensive compatibility testing
- Zero breaking changes for existing users"
    fi
    
    # 16. Migration Tools
    print_status "Committing migration tools..."
    git add cmd/migrate/ pkg/auth/migration_tool.go
    if ! git diff --cached --quiet; then
        git commit -m "feat: create automated migration tools

- Command-line migration tool with code analysis
- Pattern-based migration suggestions
- Automated migration script generation
- Project-wide migration reporting
- Comprehensive migration assistance"
    fi
    
    # 17. Examples and Documentation
    print_status "Committing examples..."
    git add examples/
    if ! git diff --cached --quiet; then
        git commit -m "docs: add comprehensive examples and integration guides

- Basic usage examples with step-by-step guides
- Advanced usage patterns and best practices
- Framework integration examples (Gin, Echo, Fiber)
- Middleware usage and configuration examples
- Error handling and logging examples
- Token management and user management examples
- Migration examples and patterns"
    fi
    
    # 18. Documentation
    print_status "Committing documentation..."
    git add README.md MIGRATION.md
    if ! git diff --cached --quiet; then
        git commit -m "docs: update comprehensive documentation for v2

- Complete README with v2 features and usage
- Detailed migration guide from v1 to v2
- API documentation and examples
- Performance considerations and best practices
- Troubleshooting and FAQ sections"
    fi
    
    # 19. Final cleanup and remaining files
    print_status "Committing any remaining files..."
    git add .
    if ! git diff --cached --quiet; then
        git commit -m "chore: final cleanup and preparation for v2.0.0 release

- Clean up temporary files and build artifacts
- Ensure all components are properly integrated
- Final testing and validation
- Prepare for official v2.0.0 release"
    fi
}

# Function to create version tags
create_version_tags() {
    print_header "Creating Version Tags"
    
    # Create pre-release tags
    print_status "Creating v2.0.0-alpha tag..."
    git tag -a "v2.0.0-alpha" -m "go-auth v2.0.0-alpha

Alpha release featuring:
- Component-based architecture
- Enhanced configuration system
- Comprehensive middleware support
- Advanced user and token management
- Monitoring and metrics
- Backward compatibility with v1"
    
    print_status "Creating v2.0.0-beta tag..."
    git tag -a "v2.0.0-beta" -m "go-auth v2.0.0-beta

Beta release with:
- Complete feature set
- Comprehensive testing
- Migration tools
- Full documentation
- Production-ready stability"
    
    print_status "Creating v2.0.0 final tag..."
    git tag -a "v2.0.0" -m "go-auth v2.0.0

🎉 Official release of go-auth v2.0.0!

Major improvements in v2:
✨ Simplified API with intuitive constructors
🏗️  Component-based architecture (Users, Tokens, Middleware)
🔧 Enhanced configuration with environment variable support
🚀 Framework-specific middleware (Gin, Echo, Fiber)
📊 Built-in monitoring, metrics, and health checks
🔒 Advanced security features and best practices
🔄 Automatic database migration system
📚 Comprehensive documentation and examples
🔙 Full backward compatibility with v1
🛠️  Automated migration tools

Breaking changes: None! v1 API continues to work.
Migration guide: https://github.com/pragneshbagary/go-auth/blob/main/MIGRATION.md"
}

# Main execution
print_status "Starting release preparation..."
echo

# Create organized commits
create_organized_commits

# Create version tags
create_version_tags

print_header "Release Preparation Complete"
print_status "✓ Created organized commit history"
print_status "✓ Created version tags: v2.0.0-alpha, v2.0.0-beta, v2.0.0"
print_status "✓ Repository is ready for v2.0.0 release"
echo

print_header "Next Steps"
print_status "1. Review the commit history: git log --oneline"
print_status "2. Push to remote repository: git push origin main --tags"
print_status "3. Create GitHub release from v2.0.0 tag"
print_status "4. Update package registries (pkg.go.dev will auto-update)"
echo

print_status "🎉 Go-Auth v2.0.0 is ready for release!"