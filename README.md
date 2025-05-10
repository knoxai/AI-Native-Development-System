# AI-Native Development System

This project provides a native development environment with AI-driven code generation capabilities. It moves away from web-based interfaces to offer a true desktop application experience with direct system access.

## Why Native Instead of Web-Based?

While web interfaces offer wide accessibility, they impose several constraints that are at odds with the goal of creating a truly AI-native development environment:

1. **Browser Limitations**: Web browsers impose sandboxing restrictions that limit access to system resources needed for deep code manipulation
2. **Performance Overhead**: Web interfaces add an extra layer between the AI and the code
3. **Conceptual Dissonance**: Using a human-oriented interface (web browser) for an AI-oriented system creates friction

The native application approach provides:

- **Direct System Access**: Better integration with the filesystem and local resources
- **Lower Latency**: Direct communication between components without browser intermediation
- **Custom UI Paradigms**: Freedom to design interfaces specifically for AI-driven development rather than human web browsing

## Architecture

The application is built using:

- **Pure Go**: For robust, type-safe backend implementation
- **Fyne UI Toolkit**: A Go-native UI framework for cross-platform applications
- **OpenRouter Integration**: For connecting to various LLM providers

### Core Components

1. **Intent Processor**: Interprets natural language development requests
2. **AST Processor**: Works with abstract code representations
3. **Semantic Model**: Maintains relationships between code entities
4. **Native UI**: A three-column desktop-native interface optimized for AI-driven development:
   - File explorer on the left
   - Code output in the middle
   - Intent input on the right

## Getting Started

### Prerequisites

- Go 1.22 or higher
- Fyne dependencies (automatically handled by Go modules)
- OpenRouter API key (only for AI code generation features)

### Installation

1. Clone the repository
2. Install dependencies:

```bash
make build
```

### Building

The project includes a Makefile with several useful targets:

```bash
# Build the native application
make build

# Create a macOS application bundle
make bundle

# Run the application
make run
```

### Using the Application

1. Launch the application
2. Configure your OpenRouter API key in Settings (click the gear icon in the status bar)
3. Select an LLM model from the settings dialog
4. Enter your development intent in the right panel
5. View the generated code, AST representation, and semantic model in the middle panel

## Features

- **Intent-based Development**: Express what you want to build in natural language
- **AST Visualization**: See the abstract syntax tree representation of your code
- **Semantic Model**: Understand the relationships between code entities
- **Native Performance**: Fast, responsive UI without browser limitations
- **Responsive Three-Column Layout**: Efficiently organize your workflow
- **Dark and Light Themes**: Choose your preferred visual style

## OpenRouter Integration

This project uses the [OpenRouter API](https://openrouter.ai) to connect to various LLM providers. To use the AI code generation features:

1. Get an API key from [OpenRouter](https://openrouter.ai)
2. Enter your API key in the Settings dialog
3. Select your preferred LLM model

## Future Directions

- **Integration with Local LLMs**: Support for running models locally
- **Extended File System Access**: Deeper integration with the host system
- **Gesture and Voice Support**: More natural interaction methods
- **Multi-monitor Support**: Optimized layouts for developer workflows
- **Plugin System**: Extensibility for custom features

## License

MIT