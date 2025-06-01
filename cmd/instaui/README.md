# Insta-Infra UI

A modern, native desktop application for managing data infrastructure services built with [Wails](https://wails.io/).

## Overview

Insta-Infra UI is a cross-platform desktop application that provides a graphical interface for managing data infrastructure services like Apache Airflow, Jupyter, PostgreSQL, and more. It's built as a native application using Go for the backend and React for the frontend.

## Features

### 🖥️ Native Desktop Application
- **No Command Line Required**: Double-click to launch like any other desktop app
- **Native Menus**: Platform-specific menu bars and keyboard shortcuts
- **Window Management**: Resizable windows with proper minimize/maximize behavior
- **System Integration**: Appears in taskbar/dock like other native applications

### 🎨 Modern UI
- **Responsive Design**: Clean, modern interface built with React and Tailwind CSS
- **Real-time Updates**: Live status monitoring of services
- **Interactive Controls**: Start, stop, and manage services with visual feedback
- **Dependency Visualization**: Graphical representation of service dependencies

### 🔧 Service Management
- **Multi-Service Support**: Manage multiple data infrastructure services
- **Status Monitoring**: Real-time service health and status checking
- **Log Viewing**: Built-in log viewer for troubleshooting
- **Connection Info**: Easy access to service URLs and connection details

## Installation & Usage

### Download Pre-built Binaries
1. Go to the [GitHub Releases](https://github.com/data-catering/insta-infra/releases) page
2. Download the appropriate binary for your platform:
   - **macOS**: `instaui-darwin-amd64` (Intel) or `instaui-darwin-arm64` (Apple Silicon)
   - **Windows**: `instaui-windows-amd64.exe`
   - **Linux**: `instaui-linux-amd64` or `instaui-linux-arm64`

### Running the Application

#### macOS
```bash
# Make executable and run
chmod +x instaui
./instaui
```

#### Windows
Simply double-click `instaui.exe` or run from command prompt:
```cmd
instaui.exe
```

#### Linux
```bash
# Make executable and run
chmod +x instaui
./instaui
```

### Building from Source

#### Prerequisites
- Go 1.21 or later
- Node.js 18 or later
- Wails CLI v2

#### Build Steps
```bash
# Clone the repository
git clone https://github.com/data-catering/insta-infra.git
cd insta-infra

# Install dependencies
make deps

# Build the UI application
make build-ui
```

The built application will be available in `cmd/instaui/build/bin/`.

## Development

### Development Mode
To run the application in development mode with hot reload:

```bash
make dev-ui
```

This will start the Wails development server with:
- Hot reload for frontend changes
- Live backend recompilation
- Developer tools enabled

### Project Structure
```
cmd/instaui/
├── app.go              # Main application logic
├── main.go             # Application entry point
├── wails.json          # Wails configuration
├── frontend/           # React frontend
│   ├── src/
│   │   ├── components/ # React components
│   │   ├── App.jsx     # Main app component
│   │   └── main.jsx    # Frontend entry point
│   ├── package.json    # Frontend dependencies
│   └── vite.config.js  # Build configuration
└── build/              # Build output (generated)
```

### Testing
```bash
# Run all tests
make test

# Run only UI tests
make test-ui

# Run with coverage
make test-ui-coverage
```

## Platform-Specific Features

### macOS
- **Native Menu Bar**: File, View, and Help menus with standard shortcuts
- **About Dialog**: Accessible from the application menu
- **App Bundle**: Properly packaged as `.app` bundle
- **Dock Integration**: Shows in dock with proper icon

### Windows
- **No Console Window**: Runs as a proper Windows application without command prompt
- **Native Window Controls**: Standard minimize, maximize, close buttons
- **Taskbar Integration**: Appears in taskbar like other Windows apps
- **NSIS Installer**: Can be packaged as Windows installer

### Linux
- **Desktop Integration**: Works with most Linux desktop environments
- **Window Manager Support**: Compatible with GNOME, KDE, XFCE, etc.
- **System Tray**: Can be configured to minimize to system tray

## Configuration

The application automatically detects and configures:
- **Container Runtime**: Docker or Podman
- **Service Discovery**: Available services and their configurations
- **Network Settings**: Port mappings and service URLs

## Troubleshooting

### Common Issues

**Application won't start:**
- Ensure Docker or Podman is installed and running
- Check that no other services are using required ports
- Verify the binary has execute permissions (Linux/macOS)

**Services not appearing:**
- Confirm `insta` CLI is installed and working
- Check that service definitions are in the correct location
- Verify container runtime is accessible

**UI not loading:**
- Try refreshing the application (Cmd/Ctrl+R)
- Check browser console for errors (if in development mode)
- Ensure frontend dependencies are installed

### Getting Help
- Check the [main documentation](../../README.md)
- Open an issue on [GitHub](https://github.com/data-catering/insta-infra/issues)
- Review logs in the application's log viewer

## Contributing

Contributions are welcome! Please see the [main contributing guide](../../CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file for details.
