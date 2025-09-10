#!/bin/bash

# Focus Helper Voice Download Script
# This script helps users download and configure voices for Focus Helper

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
VOICES_DIR="./voices"
TEMP_DIR="./temp_voices"
PYTHON_SCRIPT="./scripts/convert_voice.py"

# Available voice models
declare -A VOICE_MODELS=(
    ["pt-br-male"]="coqui-ai/TTS-models:tts_models/pt/cv/vits"
    ["pt-br-female"]="coqui-ai/TTS-models:tts_models/pt/cv/vits"
    ["en-us-male"]="coqui-ai/TTS-models:tts_models/en/ljspeech/tacotron2-DDC"
    ["en-us-female"]="coqui-ai/TTS-models:tts_models/en/ljspeech/tacotron2-DDC"
    ["multilingual"]="coqui-ai/TTS-models:tts_models/multilingual/multi-dataset/xtts_v2"
)

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

# Function to check dependencies
check_dependencies() {
    print_status "Checking dependencies..."
    
    # Check if Python is installed
    if ! command -v python3 &> /dev/null; then
        print_error "Python 3 is required but not installed."
        exit 1
    fi
    
    # Check if pip is installed
    if ! command -v pip3 &> /dev/null; then
        print_error "pip3 is required but not installed."
        exit 1
    fi
    
    # Check if wget is installed
    if ! command -v wget &> /dev/null; then
        print_error "wget is required but not installed."
        exit 1
    fi
    
    print_success "All dependencies are available."
}

# Function to install Python dependencies
install_python_deps() {
    print_status "Installing Python dependencies..."
    
    pip3 install --user TTS onnx onnxruntime
    
    print_success "Python dependencies installed."
}

# Function to create directories
create_directories() {
    print_status "Creating directories..."
    
    mkdir -p "$VOICES_DIR"
    mkdir -p "$TEMP_DIR"
    
    print_success "Directories created."
}

# Function to download a voice model
download_voice() {
    local voice_name="$1"
    local model_path="$2"
    
    print_status "Downloading voice: $voice_name"
    
    # Create temporary directory for this voice
    local voice_temp_dir="$TEMP_DIR/$voice_name"
    mkdir -p "$voice_temp_dir"
    
    # Download model using TTS
    python3 -c "
import os
from TTS.api import TTS

# Set working directory
os.chdir('$voice_temp_dir')

# Download model
tts = TTS('$model_path')
print('Model downloaded successfully')
"
    
    # Move files to voices directory
    if [ -d "$voice_temp_dir" ]; then
        cp -r "$voice_temp_dir"/* "$VOICES_DIR/"
        print_success "Voice downloaded: $voice_name"
    else
        print_error "Failed to download voice: $voice_name"
        return 1
    fi
}

# Function to convert voice to ONNX format
convert_to_onnx() {
    local voice_name="$1"
    local model_path="$2"
    
    print_status "Converting voice to ONNX format: $voice_name"
    
    python3 -c "
import os
from TTS.api import TTS

# Set working directory
os.chdir('$VOICES_DIR')

# Load and convert model
tts = TTS('$model_path')
tts.export_onnx('$voice_name.onnx')

print('Voice converted to ONNX format')
"
    
    print_success "Voice converted to ONNX: $voice_name"
}

# Function to create voice configuration
create_voice_config() {
    local voice_name="$1"
    local language="$2"
    
    print_status "Creating voice configuration: $voice_name"
    
    cat > "$VOICES_DIR/$voice_name.onnx.json" << EOF
{
  "model_name": "$voice_name",
  "language": "$language",
  "speaker_id": "default",
  "sample_rate": 22050,
  "chunk_size": 1024,
  "speed": 1.0,
  "pitch": 1.0,
  "volume": 0.8
}
EOF
    
    print_success "Voice configuration created: $voice_name"
}

# Function to show available voices
show_available_voices() {
    print_status "Available voice models:"
    echo
    for voice in "${!VOICE_MODELS[@]}"; do
        echo "  - $voice"
    done
    echo
}

# Function to list downloaded voices
list_downloaded_voices() {
    print_status "Downloaded voices:"
    echo
    
    if [ -d "$VOICES_DIR" ] && [ "$(ls -A $VOICES_DIR)" ]; then
        for voice_file in "$VOICES_DIR"/*.onnx; do
            if [ -f "$voice_file" ]; then
                voice_name=$(basename "$voice_file" .onnx)
                echo "  - $voice_name"
            fi
        done
    else
        echo "  No voices downloaded yet."
    fi
    echo
}

# Function to test a voice
test_voice() {
    local voice_name="$1"
    
    if [ ! -f "$VOICES_DIR/$voice_name.onnx" ]; then
        print_error "Voice not found: $voice_name"
        return 1
    fi
    
    print_status "Testing voice: $voice_name"
    
    python3 -c "
import onnxruntime as ort
import numpy as np

# Load ONNX model
session = ort.InferenceSession('$VOICES_DIR/$voice_name.onnx')

# Test with sample text
print('Voice model loaded successfully')
print('Model input names:', [input.name for input in session.get_inputs()])
print('Model output names:', [output.name for output in session.get_outputs()])
"
    
    print_success "Voice test completed: $voice_name"
}

# Function to clean up temporary files
cleanup() {
    print_status "Cleaning up temporary files..."
    rm -rf "$TEMP_DIR"
    print_success "Cleanup completed."
}

# Main function
main() {
    echo "Focus Helper Voice Download Script"
    echo "=================================="
    echo
    
    # Check if voices directory exists
    if [ ! -d "$VOICES_DIR" ]; then
        create_directories
    fi
    
    # Show current status
    list_downloaded_voices
    
    # Check dependencies
    check_dependencies
    
    # Install Python dependencies if needed
    if ! python3 -c "import TTS" 2>/dev/null; then
        install_python_deps
    fi
    
    # Show menu
    while true; do
        echo "Voice Download Menu:"
        echo "1. Show available voices"
        echo "2. Download a voice"
        echo "3. List downloaded voices"
        echo "4. Test a voice"
        echo "5. Clean up temporary files"
        echo "6. Exit"
        echo
        read -p "Choose an option (1-6): " choice
        
        case $choice in
            1)
                show_available_voices
                ;;
            2)
                echo
                show_available_voices
                read -p "Enter voice name to download: " voice_name
                
                if [ -n "${VOICE_MODELS[$voice_name]}" ]; then
                    download_voice "$voice_name" "${VOICE_MODELS[$voice_name]}"
                    convert_to_onnx "$voice_name" "${VOICE_MODELS[$voice_name]}"
                    
                    # Determine language based on voice name
                    if [[ "$voice_name" == *"pt-br"* ]]; then
                        language="pt-BR"
                    elif [[ "$voice_name" == *"en-us"* ]]; then
                        language="en-US"
                    else
                        language="multilingual"
                    fi
                    
                    create_voice_config "$voice_name" "$language"
                else
                    print_error "Invalid voice name: $voice_name"
                fi
                ;;
            3)
                list_downloaded_voices
                ;;
            4)
                echo
                list_downloaded_voices
                read -p "Enter voice name to test: " voice_name
                test_voice "$voice_name"
                ;;
            5)
                cleanup
                ;;
            6)
                print_status "Exiting..."
                cleanup
                exit 0
                ;;
            *)
                print_error "Invalid option. Please choose 1-6."
                ;;
        esac
        echo
    done
}

# Run main function
main "$@"
