# Getting Started with Focus Helper

Welcome to Focus Helper! This guide will help you get up and running quickly.

## 🚀 Quick Start (5 minutes)

### 1. Clone and Setup
```bash
git clone https://github.com/robertocpaes/focus-helper.git
cd focus-helper
make quick-start
```

### 2. That's it!
The application is now running in development mode. You'll see:
- Voice listener starting
- Activity monitor running
- MCP server on port 8089
- Web interface available

## 🎯 What You Can Do

### Basic Commands
- **Voice Commands**: Say your activation phrase to interact
- **Time Check**: Ask "What time is it?" or "How long have I been working?"
- **Status Check**: Say "Status" to get current information
- **Emergency**: Say "Emergency" for immediate assistance

### Configuration
- **Edit Settings**: Modify `profiles.json` for your preferences
- **Voice Setup**: Use `make download-voices` to get voice models
- **Integration**: Set up n8n workflows for notifications

## 🔧 Development

### Daily Development
```bash
make dev-run    # Start development
make dev-test   # Run tests
make dev-clean  # Clean up
```

### First Time Development Setup
```bash
make quick-start  # Complete setup
```

## 📚 Learn More

- **Full Documentation**: `make docs-serve` then visit http://localhost:1313
- **Development Guide**: See `DEVELOPMENT.md`
- **Use Cases**: See `docs/content/use-cases/`
- **Integrations**: See `docs/content/integrations/`

## 🆘 Need Help?

- **Issues**: Check GitHub Issues
- **Documentation**: Run `make docs-serve`
- **Community**: Join discussions

## 🎉 You're Ready!

Focus Helper is now running and ready to help you manage your hyperfocus. The system will:
- Monitor your activity
- Provide gentle reminders
- Send alerts when needed
- Help you maintain healthy routines

Enjoy your new focus management assistant! 🚀
