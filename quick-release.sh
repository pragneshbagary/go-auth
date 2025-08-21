#!/bin/bash

# Go-Auth v2 Quick Release Script
# This script creates a quick release without extensive testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
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

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_step() {
    echo -e "${PURPLE}🔄 $1${NC}"
}

# ASCII Art Banner
print_banner() {
    echo -e "${CYAN}"
    cat << 'EOF'
   ____            _         _   _       ____  
  / ___| ___      / \  _   _| |_| |__   |___ \ 
 | |  _ / _ \    / _ \| | | | __| '_ \    __) |
 | |_| | (_) |  / ___ \ |_| | |_| | | |  / __/ 
  \____|\___/  /_/   \_\__,_|\__|_| |_| |_____|
                                              
         🚀 Quick Release System 🚀
EOF
    echo -e "${NC}"
}

# Check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"
    
    # Check if we're in a git repository
    if [ ! -d ".git" ]; then
        print_error "Not in a git repository. Please run this script from the project root."
        exit 1
    fi
    
    # Check if git is configured
    if [ -z "$(git config user.name)" ] || [ -z "$(git config user.email)" ]; then
        print_error "Git user.name and user.email must be configured."
        print_status "Run: git config user.name 'Your Name'"
        print_status "Run: git config user.email 'your.email@example.com'"
        exit 1
    fi
    
    print_success "Prerequisites met"
}

# Quick test
quick_test() {
    print_header "Running Quick Tests"
    
    print_step "Testing build..."
    if ! go build ./pkg/... ./internal/... ./cmd/... > /dev/null 2>&1; then
        print_error "Build failed"
        exit 1
    fi
    print_success "Build successful"
    
    print_step "Testing migration tool build..."
    if ! go build -o migrate-tool ./cmd/migrate > /dev/null 2>&1; then
        print_error "Migration tool build failed"
        exit 1
    fi
    print_success "Migration tool build successful"
    
    # Clean up
    rm -f migrate-tool
}

# Show release options
show_release_options() {
    print_header "Go-Auth v2.0.0 Quick Release Options"
    echo
    echo "Choose your release strategy:"
    echo
    echo "1. 📦 Organized Release (Recommended)"
    echo "   - Creates logical, organized commits"
    echo "   - Professional commit history"
    echo "   - Ready for immediate release"
    echo
    echo "2. 📅 Spread Commits Across August 2025"
    echo "   - Spreads commits across realistic development timeline"
    echo "   - Creates appearance of month-long development"
    echo "   - Good for portfolio/showcase purposes"
    echo
    echo "3. 🔧 Custom Timeline Release"
    echo "   - Full control over commit dates and messages"
    echo "   - Most comprehensive option"
    echo "   - Includes pre-release tags (alpha, beta)"
    echo
    echo "4. ❌ Exit"
    echo
}

# Execute chosen option
execute_option() {
    local option=$1
    
    case $option in
        1)
            print_header "Executing Organized Release"
            print_step "Running prepare-release.sh..."
            if [ -f "./prepare-release.sh" ]; then
                ./prepare-release.sh
                print_success "Organized release completed!"
            else
                print_error "prepare-release.sh not found"
                exit 1
            fi
            ;;
        2)
            print_header "Executing Spread Commits Release"
            print_step "Running spread-commits.sh..."
            if [ -f "./spread-commits.sh" ]; then
                ./spread-commits.sh
                print_success "Spread commits release completed!"
            else
                print_error "spread-commits.sh not found"
                exit 1
            fi
            ;;
        3)
            print_header "Executing Custom Timeline Release"
            print_step "Running release-v2.sh..."
            if [ -f "./release-v2.sh" ]; then
                ./release-v2.sh
                print_success "Custom timeline release completed!"
            else
                print_error "release-v2.sh not found"
                exit 1
            fi
            ;;
        4)
            print_status "Exiting release process."
            exit 0
            ;;
        *)
            print_error "Invalid option selected."
            return 1
            ;;
    esac
}

# Show post-release instructions
show_post_release() {
    print_header "Post-Release Instructions"
    echo
    print_status "🎉 Release preparation completed successfully!"
    echo
    print_header "Next Steps"
    echo
    echo "1. 📋 Review Your Changes"
    echo "   git log --oneline --graph"
    echo "   git tag -l"
    echo
    echo "2. 🚀 Push to Remote Repository"
    echo "   git push origin main --tags"
    echo "   # Or: git push origin <branch-name> --tags"
    echo
    echo "3. 📦 Create GitHub Release"
    echo "   - Go to your GitHub repository"
    echo "   - Click 'Releases' → 'Create a new release'"
    echo "   - Select the v2.0.0 tag"
    echo "   - Use the tag message as release notes"
    echo "   - Publish the release"
    echo
    echo "4. 📊 Monitor and Validate"
    echo "   - Check pkg.go.dev updates (may take a few minutes)"
    echo "   - Verify documentation renders correctly"
    echo "   - Test installation: go get github.com/pragneshbagary/go-auth@v2.0.0"
    echo
    print_success "Go-Auth v2.0.0 is ready for the world! 🌟"
}

# Main execution
main() {
    print_banner
    
    # Check prerequisites
    check_prerequisites
    
    # Run quick tests
    quick_test
    
    # Show options and get user choice
    while true; do
        show_release_options
        read -p "Select option (1-4): " -n 1 -r
        echo
        echo
        
        if execute_option "$REPLY"; then
            break
        fi
        
        echo
        print_warning "Please try again."
        echo
    done
    
    # Show post-release instructions
    show_post_release
}

# Run main function
main "$@"