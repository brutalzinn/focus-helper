#!/bin/bash

# Focus Helper Development Setup Script
# This script sets up everything needed to develop Focus Helper on any machine

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="focus-helper"
REQUIRED_GO_VERSION="1.21"
REQUIRED_CMAKE_VERSION="3.15"
REQUIRED_MAKE_VERSION="4.0"

# Function to print colored output
print_header() {
    echo -e "${PURPLE}================================${NC}"
    echo -e "${PURPLE}  Focus Helper Development Setup${NC}"
    echo -e "${PURPLE}================================${NC}"
    echo
}

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

print_step() {
    echo -e "${CYAN}[STEP]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check OS
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
    elif [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
        OS="windows"
    else
        OS="unknown"
    fi
    echo "$OS"
}

# Function to check if running as root
check_root() {
    if [[ $EUID -eq 0 ]]; then
        print_error "This script should not be run as root. Please run as a regular user."
        exit 1
    fi
}

# Function to check system requirements
check_requirements() {
    print_step "Checking system requirements..."
    
    local os=$(detect_os)
    print_status "Detected OS: $os"
    
    # Check if we're in the right directory
    if [[ ! -f "go.mod" ]] || [[ ! -d "src" ]]; then
        print_error "Please run this script from the Focus Helper project root directory."
        exit 1
    fi
    
    # Check Go version
    if command_exists go; then
        local go_version=$(go version | awk '{print $3}' | sed 's/go//')
        local go_major=$(echo $go_version | cut -d. -f1)
        local go_minor=$(echo $go_version | cut -d. -f2)
        local required_major=$(echo $REQUIRED_GO_VERSION | cut -d. -f1)
        local required_minor=$(echo $REQUIRED_GO_VERSION | cut -d. -f2)
        
        if [[ $go_major -gt $required_major ]] || [[ $go_major -eq $required_major && $go_minor -ge $required_minor ]]; then
            print_success "Go version $go_version is compatible (required: $REQUIRED_GO_VERSION+)"
        else
            print_error "Go version $go_version is too old. Required: $REQUIRED_GO_VERSION+"
            print_status "Please install Go $REQUIRED_GO_VERSION or later from https://golang.org/dl/"
            exit 1
        fi
    else
        print_error "Go is not installed. Please install Go $REQUIRED_GO_VERSION or later from https://golang.org/dl/"
        exit 1
    fi
    
    # Check CMake
    if command_exists cmake; then
        local cmake_version=$(cmake --version | head -n1 | awk '{print $3}')
        print_success "CMake version $cmake_version found"
    else
        print_error "CMake is not installed. Please install CMake $REQUIRED_CMAKE_VERSION or later."
        print_status "Install instructions:"
        if [[ "$os" == "linux" ]]; then
            echo "  Ubuntu/Debian: sudo apt-get install cmake"
            echo "  CentOS/RHEL: sudo yum install cmake"
            echo "  Fedora: sudo dnf install cmake"
        elif [[ "$os" == "macos" ]]; then
            echo "  brew install cmake"
        fi
        exit 1
    fi
    
    # Check Make
    if command_exists make; then
        local make_version=$(make --version | head -n1 | awk '{print $3}')
        print_success "Make version $make_version found"
    else
        print_error "Make is not installed. Please install Make."
        exit 1
    fi
    
    # Check Git
    if command_exists git; then
        print_success "Git found"
    else
        print_error "Git is not installed. Please install Git."
        exit 1
    fi
    
    print_success "All basic requirements met!"
}

# Function to install system dependencies
install_system_deps() {
    print_step "Installing system dependencies..."
    
    local os=$(detect_os)
    
    if [[ "$os" == "linux" ]]; then
        print_status "Installing Linux dependencies..."
        
        # Detect package manager
        if command_exists apt-get; then
            print_status "Using apt-get..."
            sudo apt-get update
            sudo apt-get install -y \
                build-essential \
                cmake \
                portaudio19-dev \
                libasound2-dev \
                pkg-config \
                wget \
                curl \
                git \
                make \
                gcc \
                g++ \
                libssl-dev \
                libffi-dev \
                python3-dev \
                python3-pip
        elif command_exists yum; then
            print_status "Using yum..."
            sudo yum groupinstall -y "Development Tools"
            sudo yum install -y \
                cmake \
                portaudio-devel \
                alsa-lib-devel \
                pkgconfig \
                wget \
                curl \
                git \
                make \
                gcc \
                gcc-c++ \
                openssl-devel \
                libffi-devel \
                python3-devel \
                python3-pip
        elif command_exists dnf; then
            print_status "Using dnf..."
            sudo dnf groupinstall -y "Development Tools"
            sudo dnf install -y \
                cmake \
                portaudio-devel \
                alsa-lib-devel \
                pkgconfig \
                wget \
                curl \
                git \
                make \
                gcc \
                gcc-c++ \
                openssl-devel \
                libffi-devel \
                python3-devel \
                python3-pip
        else
            print_warning "Unknown package manager. Please install the following manually:"
            echo "  - build-essential (or equivalent)"
            echo "  - cmake"
            echo "  - portaudio19-dev (or equivalent)"
            echo "  - libasound2-dev (or equivalent)"
            echo "  - pkg-config"
            echo "  - wget, curl, git, make"
            echo "  - gcc, g++"
            echo "  - libssl-dev, libffi-dev"
            echo "  - python3-dev, python3-pip"
        fi
        
    elif [[ "$os" == "macos" ]]; then
        print_status "Installing macOS dependencies..."
        
        if command_exists brew; then
            print_status "Using Homebrew..."
            brew install \
                cmake \
                portaudio \
                pkg-config \
                wget \
                curl \
                git \
                make \
                gcc \
                openssl \
                libffi \
                python3
        else
            print_warning "Homebrew not found. Please install the following manually:"
            echo "  - cmake"
            echo "  - portaudio"
            echo "  - pkg-config"
            echo "  - wget, curl, git, make"
            echo "  - gcc"
            echo "  - openssl, libffi"
            echo "  - python3"
        fi
        
    elif [[ "$os" == "windows" ]]; then
        print_warning "Windows setup requires manual installation. Please install:"
        echo "  - Go from https://golang.org/dl/"
        echo "  - CMake from https://cmake.org/download/"
        echo "  - Git from https://git-scm.com/download/win"
        echo "  - Visual Studio Build Tools"
        echo "  - PortAudio (via vcpkg or manual compilation)"
        echo "  - Python 3 from https://python.org/"
        
    else
        print_error "Unsupported operating system: $os"
        exit 1
    fi
    
    print_success "System dependencies installed!"
}

# Function to setup Go environment
setup_go() {
    print_step "Setting up Go environment..."
    
    # Check if GOPATH is set
    if [[ -z "$GOPATH" ]]; then
        print_status "Setting up GOPATH..."
        echo 'export GOPATH=$HOME/go' >> ~/.bashrc
        echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
        export GOPATH=$HOME/go
        export PATH=$PATH:$GOPATH/bin
    fi
    
    # Create Go directories
    mkdir -p $GOPATH/{bin,src,pkg}
    
    print_success "Go environment configured!"
}

# Function to install Python dependencies
install_python_deps() {
    print_step "Installing Python dependencies..."
    
    if command_exists python3; then
        print_status "Installing Python packages..."
        python3 -m pip install --user --upgrade pip
        python3 -m pip install --user \
            TTS \
            onnx \
            onnxruntime \
            numpy \
            scipy \
            librosa \
            soundfile
        print_success "Python dependencies installed!"
    else
        print_warning "Python3 not found. Skipping Python dependencies."
    fi
}

# Function to setup git hooks
setup_git_hooks() {
    print_step "Setting up Git hooks..."
    
    # Create pre-commit hook
    cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
# Pre-commit hook for Focus Helper

echo "Running pre-commit checks..."

# Run go fmt
echo "Running go fmt..."
go fmt ./src/...

# Run go vet
echo "Running go vet..."
go vet ./src/...

# Run tests
echo "Running tests..."
go test ./src/... -short

echo "Pre-commit checks completed!"
EOF
    
    chmod +x .git/hooks/pre-commit
    
    print_success "Git hooks configured!"
}

# Function to create development configuration
create_dev_config() {
    print_step "Creating development configuration..."
    
    # Create development profile
    cat > profiles-dev.json << 'EOF'
{
  "name": "development",
  "debug": true,
  "log_level": "debug",
  "voice": {
    "enabled": false
  },
  "mcp": {
    "enabled": true,
    "port": 8089
  },
  "webhook": {
    "enabled": false
  },
  "llm": {
    "enabled": false
  }
}
EOF
    
    print_success "Development configuration created!"
}

# Function to run initial build
run_initial_build() {
    print_step "Running initial build..."
    
    # Update submodules
    print_status "Updating git submodules..."
    git submodule update --init --recursive
    
    # Build whisper dependencies
    print_status "Building whisper dependencies..."
    make build-whisper
    
    # Build the application
    print_status "Building Focus Helper..."
    make build
    
    print_success "Initial build completed!"
}

# Function to run tests
run_tests() {
    print_step "Running tests..."
    
    make test
    
    print_success "Tests completed!"
}

# Function to create development scripts
create_dev_scripts() {
    print_step "Creating development scripts..."
    
    # Create dev-run script
    cat > scripts/dev-run.sh << 'EOF'
#!/bin/bash
# Development run script

echo "Starting Focus Helper in development mode..."
make run
EOF
    
    chmod +x scripts/dev-run.sh
    
    # Create dev-test script
    cat > scripts/dev-test.sh << 'EOF'
#!/bin/bash
# Development test script

echo "Running development tests..."
make test-coverage
echo "Coverage report: coverage.html"
EOF
    
    chmod +x scripts/dev-test.sh
    
    # Create dev-clean script
    cat > scripts/dev-clean.sh << 'EOF'
#!/bin/bash
# Development clean script

echo "Cleaning development artifacts..."
make clean
rm -f profiles-dev.json
rm -f focus_helper*.log
rm -f focus_helper*.db
EOF
    
    chmod +x scripts/dev-clean.sh
    
    print_success "Development scripts created!"
}

# Function to show next steps
show_next_steps() {
    print_step "Setup completed! Next steps:"
    echo
    echo -e "${GREEN}1. Start developing:${NC}"
    echo "   make run                    # Run in development mode"
    echo "   make test                  # Run tests"
    echo "   make build                 # Build production binary"
    echo
    echo -e "${GREEN}2. Development scripts:${NC}"
    echo "   ./scripts/dev-run.sh       # Quick development run"
    echo "   ./scripts/dev-test.sh      # Run tests with coverage"
    echo "   ./scripts/dev-clean.sh     # Clean development artifacts"
    echo
    echo -e "${GREEN}3. Configuration:${NC}"
    echo "   Edit profiles-dev.json     # Development configuration"
    echo "   Edit profiles.json         # Production configuration"
    echo
    echo -e "${GREEN}4. Voice setup:${NC}"
    echo "   ./scripts/download-voices.sh  # Download voice models"
    echo
    echo -e "${GREEN}5. Documentation:${NC}"
    echo "   cd docs && hugo server     # Start documentation server"
    echo
    echo -e "${GREEN}6. Integration:${NC}"
    echo "   Check docs/content/integrations/ for n8n setup"
    echo
    echo -e "${YELLOW}Happy coding! 🚀${NC}"
}

# Main function
main() {
    print_header
    
    # Check if running as root
    check_root
    
    # Check requirements
    check_requirements
    
    # Install system dependencies
    install_system_deps
    
    # Setup Go environment
    setup_go
    
    # Install Python dependencies
    install_python_deps
    
    # Setup git hooks
    setup_git_hooks
    
    # Create development configuration
    create_dev_config
    
    # Create development scripts
    create_dev_scripts
    
    # Run initial build
    run_initial_build
    
    # Run tests
    run_tests
    
    # Show next steps
    show_next_steps
}

# Run main function
main "$@"
