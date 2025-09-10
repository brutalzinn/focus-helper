#!/bin/bash

# Hugo Documentation Setup Script
# This script sets up Hugo for the Focus Helper documentation

set -e

echo "🚀 Setting up Hugo documentation for Focus Helper..."

# Check if Hugo is installed
if ! command -v hugo &> /dev/null; then
    echo "❌ Hugo is not installed. Please install Hugo first:"
    echo "   - Go to https://gohugo.io/installation/"
    echo "   - Or use: brew install hugo (macOS)"
    echo "   - Or use: snap install hugo (Linux)"
    exit 1
fi

echo "✅ Hugo is installed: $(hugo version)"

# Check if we're in the docs directory
if [ ! -f "config.yaml" ]; then
    echo "❌ Please run this script from the docs directory"
    exit 1
fi

# Install the theme
echo "📦 Installing Hugo TechDoc theme..."

# Create themes directory if it doesn't exist
mkdir -p themes

# Clone the theme
if [ ! -d "themes/hugo-theme-techdoc" ]; then
    echo "📥 Cloning Hugo TechDoc theme..."
    git clone https://github.com/thingsym/hugo-theme-techdoc.git themes/hugo-theme-techdoc
else
    echo "✅ Theme already exists"
fi

# Create necessary directories
echo "📁 Creating necessary directories..."
mkdir -p static/images
mkdir -p static/css
mkdir -p static/js
mkdir -p layouts/partials
mkdir -p layouts/shortcodes

# Create a simple README for the docs
cat > README.md << 'EOF'
# Focus Helper Documentation

This directory contains the Hugo-based documentation for Focus Helper.

## Quick Start

1. **Install Hugo** (if not already installed):
   ```bash
   # macOS
   brew install hugo
   
   # Linux
   snap install hugo
   
   # Windows
   choco install hugo
   ```

2. **Start the development server**:
   ```bash
   hugo server -D
   ```

3. **Open your browser** to `http://localhost:1313`

## Building for Production

```bash
hugo --minify
```

The built site will be in the `public/` directory.

## Content Structure

- `content/` - Markdown content files
- `static/` - Static assets (images, CSS, JS)
- `themes/` - Hugo themes
- `config.yaml` - Site configuration

## Adding New Content

1. Create a new markdown file in `content/`
2. Add front matter with title, description, date, etc.
3. The file will automatically appear in the navigation

## Customization

- Edit `config.yaml` for site-wide settings
- Modify theme files in `themes/hugo-theme-techdoc/`
- Add custom CSS in `static/css/`
- Add custom JavaScript in `static/js/`
EOF

echo "✅ Documentation setup complete!"
echo ""
echo "🎉 Next steps:"
echo "1. Start the development server: hugo server -D"
echo "2. Open http://localhost:1313 in your browser"
echo "3. Edit content in the content/ directory"
echo "4. Build for production: hugo --minify"
echo ""
echo "📚 Happy documenting!"
