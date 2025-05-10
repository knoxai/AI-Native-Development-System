# AI-Oriented Software Development Environment

Currently, most AI coding systems like GitHub Copilot, Cursor, and Buddy work through existing human-oriented IDEs:

- They interact with text-based code representations
- They generate code that fits into file structures designed for humans
- They're constrained by interfaces optimized for human cognition and interaction

## Why This Isn't Optimal for AI

Human-oriented IDEs aren't necessarily ideal for AI because:

1. AIs don't benefit from many IDE features designed for human limitations (syntax highlighting, visual organization)
2. AIs can process abstract code representations that might be more efficient than text
3. The file-based organization of code is an artifact of human cognitive constraints

## Better Approaches for AI-Driven Development

Several alternative paradigms could better leverage AI's capabilities:

### Direct Abstract Syntax Tree (AST) Manipulation
```
AI could work directly with structured code representations rather than generating text that must be parsed.
```

### Semantic Code Models
```python
# Instead of text files like this
def calculate_total(items):
    return sum(item.price for item in items)
```

AI could interact with semantic models representing the same functionality as relationships, constraints, and behaviors.

### Intent-Based Development Systems

Rather than writing code, developers could express intentions, and AI could:
- Understand the high-level goal
- Generate implementation automatically
- Reason about correctness without the text-based intermediate step

### API-Native Development

AI systems could interact directly with compilers, build systems, and deployment pipelines through purpose-built APIs rather than simulating human interaction patterns.

### Automated Verification Integration

AI-specific environments could integrate formal verification methods that continuously validate that generated code meets specifications.

## Concept

Traditional IDEs are designed for human developers with features like syntax highlighting, visual organization, and file-based structure. However, AI systems don't have the same cognitive constraints and can work with more abstract representations of code.

This project explores a new paradigm with:

1. **Intent-based Development**: Express what you want to build in natural language, and the AI interprets and implements it
2. **Abstract Syntax Tree (AST) Manipulation**: Direct work with code structure rather than text files
3. **Semantic Code Models**: Understanding code as relationships, constraints, and behaviors
4. **API-Native Development**: Direct interaction with compilers and build systems

## Architecture

The system consists of:

- **Intent Processor**: Interprets natural language development requests
- **AST Processor**: Works with abstract code representations
- **Semantic Model**: Maintains relationships between code entities
- **HTTP API Server**: Provides endpoints for client interaction
- **Web UI**: A simple interface to interact with the system
- **LLM Integration**: Uses OpenRouter API to connect to various AI models

## OpenRouter Integration

This project includes integration with the [OpenRouter API](https://docs.openrouter.co), allowing the system to use a variety of LLMs to process intent and generate code.

### API Key Requirements

The system has two modes of operation:

1. **Browse-only Mode**: Without an API key, you can browse all available models but cannot generate code
2. **Full Mode**: With an API key, you can browse models and generate code

### Setting Up OpenRouter

1. Register for an account at [OpenRouter](https://openrouter.co/)
2. Get your API key from the dashboard (only needed for code generation)
3. You can provide your API key in one of two ways:
   - **Directly in the web interface**: Enter your API key in the web UI (recommended for most users)
   - **Environment variable**: Set the key in your environment before starting the server
     ```bash
     export OPENROUTER_API_KEY="your_api_key_here"
     ```

When you use the web interface to enter your API key:
- The key is stored securely in your browser's localStorage
- It's never stored on the server
- It's sent with each request that requires it
- You can clear it at any time from the web interface

### Model Selection

The system includes a model selection UI that allows you to:

1. View and select from all available models in OpenRouter (no API key required)
2. See model information including context length and pricing
3. Save your preferred model for future sessions

Model data is cached in your browser for 12 hours to improve performance.

### How It Works

1. **Intent Parsing**: Natural language intents are sent to the selected LLM, which parses them into structured representations
2. **Code Generation**: The LLM generates code based on the intent, along with AST and semantic representations
3. **Response Processing**: The system processes the LLM's response, extracting code, AST, and semantic information

## Getting Started

### Prerequisites

- Go 1.16 or higher
- OpenRouter API key (only for AI code generation features)

### Installation

1. Clone the repository
2. Build the project:

```bash
cd ai-dev-env
go build -o ai-dev-env ./cmd/ai-dev-env
```

### Running

```bash
# To enable AI code generation, set your OpenRouter API key
export OPENROUTER_API_KEY="your_api_key_here"

# Run the application
./ai-dev-env
```

The server will start on http://localhost:8080

### Using Docker

#### Building and Running with Docker

1. Build the Docker image:
```bash
docker build -t ai-dev-env .
```

2. Run the container:
```bash
# To run with AI code generation enabled:
docker run -p 8080:8080 -e OPENROUTER_API_KEY="your_api_key_here" ai-dev-env

# To run in browse-only mode:
docker run -p 8080:8080 ai-dev-env
```

The server will be available at http://localhost:8080

#### Using Docker Compose

1. Create a `.env` file with your API key (optional for browse-only mode):
```
OPENROUTER_API_KEY=your_api_key_here
```

2. Start the application:
```bash
docker-compose up
```

3. For background mode:
```bash
docker-compose up -d
```

4. Stop the application:
```bash
docker-compose down
```

### Using the System

1. Open your browser to http://localhost:8080
2. Select the AI model you want to use from the dropdown
3. Enter your development intent in natural language (e.g., "Create a login function that validates user credentials")
4. Click "Execute" to generate code (requires API key)
5. View the generated code, AST representation, and semantic model

## API Endpoints

- `POST /api/intent`: Process a natural language development intent
- `POST /api/ast`: Directly manipulate the abstract syntax tree
- `POST /api/semantics`: Query the semantic model
- `POST /api/models/select`: Select an AI model to use

## Limitations

This is a proof-of-concept that demonstrates the potential for AI-native development environments. Some features are simplified or mocked for demonstration purposes.

## Future Directions

- Integration with real LLM services for more sophisticated intent parsing
- Live code generation and compilation
- Support for multiple programming languages
- Collaborative development features
- Integration with version control systems

## License

MIT