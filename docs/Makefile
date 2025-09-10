# Focus Helper Documentation Makefile

.PHONY: help setup serve build clean install-theme

# Default target
help:
	@echo "Focus Helper Documentation"
	@echo "========================="
	@echo ""
	@echo "Available targets:"
	@echo "  setup          - Set up Hugo and install theme"
	@echo "  serve          - Start development server"
	@echo "  build          - Build static site for production"
	@echo "  clean          - Clean build artifacts"
	@echo "  install-theme  - Install Hugo TechDoc theme"
	@echo "  help           - Show this help message"
	@echo ""

# Set up Hugo and install theme
setup: install-theme
	@echo "✅ Documentation setup complete!"
	@echo "Run 'make serve' to start the development server"

# Install Hugo TechDoc theme
install-theme:
	@echo "📦 Installing Hugo TechDoc theme..."
	@mkdir -p themes
	@if [ ! -d "themes/hugo-theme-techdoc" ]; then \
		git clone https://github.com/thingsym/hugo-theme-techdoc.git themes/hugo-theme-techdoc; \
	else \
		echo "✅ Theme already exists"; \
	fi

# Start development server
serve:
	@echo "🚀 Starting Hugo development server..."
	@echo "Open http://localhost:1313 in your browser"
	@hugo server -D --bind 0.0.0.0

# Build static site for production
build:
	@echo "🏗️  Building static site..."
	@hugo --minify
	@echo "✅ Site built in public/ directory"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf public/
	@rm -rf resources/
	@echo "✅ Clean complete"

# Install Hugo (if not installed)
install-hugo:
	@echo "📦 Installing Hugo..."
	@if command -v brew >/dev/null 2>&1; then \
		brew install hugo; \
	elif command -v snap >/dev/null 2>&1; then \
		sudo snap install hugo; \
	elif command -v apt >/dev/null 2>&1; then \
		sudo apt update && sudo apt install hugo; \
	else \
		echo "❌ Please install Hugo manually from https://gohugo.io/installation/"; \
		exit 1; \
	fi

# Check if Hugo is installed
check-hugo:
	@if ! command -v hugo &> /dev/null; then \
		echo "❌ Hugo is not installed. Run 'make install-hugo' first."; \
		exit 1; \
	else \
		echo "✅ Hugo is installed: $$(hugo version)"; \
	fi

# Deploy to GitHub Pages (if configured)
deploy: build
	@echo "🚀 Deploying to GitHub Pages..."
	@if [ -d ".git" ]; then \
		git add public/; \
		git commit -m "Deploy documentation $(shell date)"; \
		git subtree push --prefix public origin gh-pages; \
	else \
		echo "❌ Not a git repository. Cannot deploy."; \
	fi

# Validate content
validate:
	@echo "🔍 Validating content..."
	@hugo --quiet
	@echo "✅ Content validation passed"

# Create new content page
new-page:
	@echo "📝 Creating new content page..."
	@read -p "Enter page title: " title; \
	read -p "Enter page filename (slug): " slug; \
	hugo new content/$$slug.md; \
	echo "✅ Created content/$$slug.md"

# Show site statistics
stats:
	@echo "📊 Site Statistics:"
	@echo "Pages: $$(find content -name "*.md" | wc -l)"
	@echo "Images: $$(find static -name "*.png" -o -name "*.jpg" -o -name "*.gif" | wc -l)"
	@echo "Total size: $$(du -sh . | cut -f1)"
