# Documentation Summary

This document provides an overview of the Focus Helper documentation structure and how to use it.

## 📁 Documentation Structure

```
docs/
├── config.yaml              # Hugo site configuration
├── Makefile                 # Documentation build commands
├── setup-hugo.sh           # Hugo setup script
├── content/                 # Markdown content files
│   ├── _index.md           # Home page
│   ├── getting-started/    # Getting started guide
│   ├── features/           # Features overview
│   ├── mcp-integration/    # MCP integration guide
│   ├── docker/             # Docker documentation
│   └── examples/           # Examples and use cases
│       ├── _index.md       # Examples overview
│       ├── n8n-integration.md  # n8n workflow automation
│       ├── mcp_client.go   # Go client example
│       └── mcp_client.py   # Python client example
├── static/                  # Static assets
├── themes/                  # Hugo themes
│   └── hugo-theme-techdoc/ # TechDoc theme
├── public/                  # Built static site
└── README.md               # Documentation README
```

## 🚀 Quick Start

### 1. Setup Hugo Documentation

```bash
cd docs
make setup    # First time setup
make serve    # Start development server
```

### 2. View Documentation

Open `http://localhost:1313` in your browser.

### 3. Build for Production

```bash
make build    # Build static site
```

## 📚 Content Overview

### Main Sections

1. **[Home](content/_index.md)** - Project overview and quick start
2. **[Getting Started](content/getting-started/)** - Installation and setup
3. **[Features](content/features/)** - Complete feature overview
4. **[MCP Integration](content/mcp-integration/)** - External tool integration
5. **[Docker](content/docker/)** - Containerized deployment
6. **[Examples](content/examples/)** - Practical examples and use cases

### Examples Included

- **n8n Integration** - Workflow automation with n8n
- **Go Client** - Go client library for MCP
- **Python Client** - Python client library for MCP
- **Cursor IDE Integration** - Code editor integration
- **Docker Compose** - Container orchestration
- **Webhook Integration** - External service integration

## 🛠️ Available Commands

### Documentation Commands

```bash
make help          # Show available commands
make setup         # Set up Hugo and install theme
make serve         # Start development server
make build         # Build static site for production
make clean         # Clean build artifacts
make validate      # Validate content
make stats         # Show site statistics
```

### Content Management

```bash
make new-page      # Create new content page
make install-theme # Install Hugo TechDoc theme
```

## 🎨 Theme Customization

The documentation uses the Hugo TechDoc theme with the following customizations:

### Configuration (config.yaml)

- **Theme**: Hugo TechDoc
- **Base URL**: https://focus-helper.dev
- **Language**: English
- **Navigation**: Custom menu structure
- **Features**: Highlighted key features
- **Social Links**: GitHub integration

### Custom Features

- **Responsive Design**: Mobile-friendly layout
- **Search**: Built-in search functionality
- **Syntax Highlighting**: Code syntax highlighting
- **Table of Contents**: Auto-generated TOC
- **Breadcrumbs**: Navigation breadcrumbs
- **Dark Mode**: Automatic theme switching

## 📝 Adding New Content

### 1. Create New Page

```bash
make new-page
# Enter page title and filename
```

### 2. Edit Existing Page

Edit files in `content/` directory:

```markdown
---
title: "Page Title"
description: "Page description"
date: 2025-09-09T19:00:00Z
draft: false
weight: 10
---

# Page Content

Your content here...
```

### 3. Add Images

Place images in `static/images/` and reference them:

```markdown
![Alt text](/images/image-name.png)
```

## 🔧 Configuration

### Site Configuration (config.yaml)

- **Title**: Focus Helper
- **Base URL**: https://focus-helper.dev
- **Theme**: hugo-theme-techdoc
- **Language**: en-us
- **Menu**: Custom navigation structure
- **Features**: Key features showcase
- **Social**: GitHub links

### Content Structure

- **Front Matter**: Required for all pages
- **Weight**: Controls page order
- **Draft**: Set to false for published content
- **Date**: Publication date
- **Description**: Page description for SEO

## 🚀 Deployment

### GitHub Pages

```bash
make deploy    # Deploy to GitHub Pages
```

### Custom Hosting

```bash
make build     # Build static site
# Upload public/ directory to your web server
```

### Docker

```bash
# Build documentation container
docker build -t focus-helper-docs .

# Run documentation server
docker run -p 1313:1313 focus-helper-docs
```

## 📊 Statistics

Current documentation includes:

- **Pages**: 22+ content pages
- **Examples**: 5+ practical examples
- **Languages**: Go, Python, JavaScript, YAML, Markdown
- **Features**: Complete API documentation
- **Integrations**: n8n, Cursor, Docker, MCP

## 🔍 Search and Navigation

### Built-in Features

- **Search**: Full-text search across all content
- **Categories**: Content categorization
- **Tags**: Content tagging system
- **Breadcrumbs**: Navigation breadcrumbs
- **Table of Contents**: Auto-generated TOC
- **Responsive Menu**: Mobile-friendly navigation

### SEO Features

- **Meta Tags**: Automatic meta tag generation
- **Sitemap**: Auto-generated sitemap
- **RSS Feed**: RSS feed for updates
- **Open Graph**: Social media sharing
- **Structured Data**: Rich snippets support

## 🐛 Troubleshooting

### Common Issues

1. **Hugo not found**: Run `make install-hugo`
2. **Theme missing**: Run `make install-theme`
3. **Build errors**: Check content syntax
4. **Server not starting**: Check port availability

### Debug Mode

```bash
hugo server -D --verbose
```

### Content Validation

```bash
make validate
```

## 📈 Future Enhancements

### Planned Features

- **Multi-language Support**: Portuguese documentation
- **Interactive Examples**: Live code examples
- **Video Tutorials**: Embedded video content
- **API Explorer**: Interactive API documentation
- **Community Contributions**: User-generated content

### Content Roadmap

- **Advanced Integrations**: More tool integrations
- **Case Studies**: Real-world usage examples
- **Best Practices**: Development guidelines
- **Troubleshooting**: Common issues and solutions
- **Performance**: Optimization guides

---

**Ready to contribute?** Check out our [Contributing Guidelines](../CONTRIBUTING.md) and start adding to the documentation!
