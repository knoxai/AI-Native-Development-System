#!/bin/bash

# Check if API key was provided as an argument
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <your_openrouter_api_key>"
    echo "Example: $0 sk-or-v1-..."
    exit 1
fi

API_KEY="$1"

# Set the API key in the current shell
export OPENROUTER_API_KEY="$API_KEY"
echo "OpenRouter API Key set for current shell: ${API_KEY:0:4}****${API_KEY: -4}"

# Append to .env file if it exists, or create it
if [ -f .env ]; then
    # Update existing key or add new one
    if grep -q "OPENROUTER_API_KEY=" .env; then
        # Use sed to replace existing key
        sed -i '' 's|OPENROUTER_API_KEY=.*|OPENROUTER_API_KEY='"$API_KEY"'|' .env
        echo "Updated API key in .env file"
    else
        # Append to file
        echo "OPENROUTER_API_KEY=$API_KEY" >> .env
        echo "Added API key to .env file"
    fi
else
    # Create new .env file
    echo "OPENROUTER_API_KEY=$API_KEY" > .env
    echo "Created .env file with API key"
fi

# Suggest next steps
echo ""
echo "Next steps:"
echo "1. To test the API connection, run: go run test_openrouter.go"
echo "2. To start the application with the key, run: ./ai-dev-env"
echo "3. For Docker, build with: docker build -t ai-dev-env ."
echo "4. Run with Docker: docker run -p 8080:8080 -e OPENROUTER_API_KEY=$API_KEY ai-dev-env"
echo ""
echo "Note: For new shells, you'll need to export the API key again or source the .env file"
echo "To make it permanent, add to your shell profile (~/.bash_profile, ~/.zshrc, etc.):"
echo "export OPENROUTER_API_KEY=$API_KEY" 