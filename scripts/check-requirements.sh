#!/bin/bash

# Focus Helper Requirements Check Script
# This script checks if all requirements are met for development

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REQUIRED_GO_VERSION="1.21"
REQUIRED_CMAKE_VERSION="3.15"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check version
check_version() {
    local cmd="$1"
    local version_cmd="$2"
    local required_version="$3"
    local version_name="$4"
    
    if command_exists "$cmd"; then
        local version=$($version_cmd 2>/dev/null | head -n1 | awk '{print $3}' | sed 's/[^0-9.]//g')
        if [[ -n "$version" ]]; then
            print_success "$version_name version $version found"
            return 0
        else
            print_warning "$version_name found but version could not be determined"
            return 1
        fi
    else
        print_error "$version_name is not installed"
        return 1
    fi
}

# Function to check Go version
check_go_version() {
    if command_exists go; then
        local go_version=$(go version | awk '{print $3}' | sed 's/go//')
        local go_major=$(echo $go_version | cut -d. -f1)
        local go_minor=$(echo $go_version | cut -d. -f2)
        local required_major=$(echo $REQUIRED_GO_VERSION | cut -d. -f1)
        local required_minor=$(echo $REQUIRED_GO_VERSION | cut -d. -f2)
        
        if [[ $go_major -gt $required_major ]] || [[ $go_major -eq $required_major && $go_minor -ge $required_minor ]]; then
            print_success "Go version $go_version is compatible (required: $REQUIRED_GO_VERSION+)"
            return 0
        else
            print_error "Go version $go_version is too old. Required: $REQUIRED_GO_VERSION+"
            return 1
        fi
    else
        print_error "Go is not installed. Please install Go $REQUIRED_GO_VERSION or later"
        return 1
    fi
}

# Function to check CMake version
check_cmake_version() {
    if command_exists cmake; then
        local cmake_version=$(cmake --version | head -n1 | awk '{print $3}')
        local cmake_major=$(echo $cmake_version | cut -d. -f1)
        local cmake_minor=$(echo $cmake_version | cut -d. -f2)
        local required_major=$(echo $REQUIRED_CMAKE_VERSION | cut -d. -f1)
        local required_minor=$(echo $REQUIRED_CMAKE_VERSION | cut -d. -f2)
        
        if [[ $cmake_major -gt $required_major ]] || [[ $cmake_major -eq $required_major && $cmake_minor -ge $required_minor ]]; then
            print_success "CMake version $cmake_version is compatible (required: $REQUIRED_CMAKE_VERSION+)"
            return 0
        else
            print_error "CMake version $cmake_version is too old. Required: $REQUIRED_CMAKE_VERSION+"
            return 1
        fi
    else
        print_error "CMake is not installed. Please install CMake $REQUIRED_CMAKE_VERSION or later"
        return 1
    fi
}

# Main function
main() {
    echo "Focus Helper Requirements Check"
    echo "==============================="
    echo
    
    local all_good=true
    
    # Check Go
    print_status "Checking Go..."
    if ! check_go_version; then
        all_good=false
    fi
    echo
    
    # Check CMake
    print_status "Checking CMake..."
    if ! check_cmake_version; then
        all_good=false
    fi
    echo
    
    # Check Make
    print_status "Checking Make..."
    if ! check_version "make" "make --version" "" "Make"; then
        all_good=false
    fi
    echo
    
    # Check Git
    print_status "Checking Git..."
    if ! check_version "git" "git --version" "" "Git"; then
        all_good=false
    fi
    echo
    
    # Check Python (optional)
    print_status "Checking Python (optional for voice features)..."
    if command_exists python3; then
        local python_version=$(python3 --version | awk '{print $2}')
        print_success "Python version $python_version found"
    else
        print_warning "Python3 not found (optional for voice features)"
    fi
    echo
    
    # Check if we're in the right directory
    print_status "Checking project directory..."
    if [[ -f "go.mod" ]] && [[ -d "src" ]]; then
        print_success "Running from correct project directory"
    else
        print_error "Please run this script from the Focus Helper project root directory"
        all_good=false
    fi
    echo
    
    # Summary
    if [[ "$all_good" == true ]]; then
        print_success "All requirements met! You can run 'make quick-start' to begin development."
    else
        print_error "Some requirements are missing. Please install the missing dependencies and try again."
        echo
        echo "Installation help:"
        echo "  - Go: https://golang.org/dl/"
        echo "  - CMake: https://cmake.org/download/"
        echo "  - Make: Usually pre-installed on Linux/macOS"
        echo "  - Git: https://git-scm.com/downloads"
        echo "  - Python: https://python.org/downloads/ (optional)"
        exit 1
    fi
}

# Run main function
main "$@"
