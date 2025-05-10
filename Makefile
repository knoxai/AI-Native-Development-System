.PHONY: all clean build run-web run-native bundle macos-app

# Default target
all: build

# Clean the project
clean:
	rm -rf bin/*
	rm -rf *.app

# Build all applications
build: build-web build-native

# Build the web application
build-web:
	go build -o bin/ai-dev-env ./cmd/ai-dev-env

# Build the native application
build-native:
	go build -o bin/ai-dev-env-native ./cmd/ai-dev-env-native

# Build a native macOS application bundle with fyne's tools
bundle:
	fyne package -os darwin -icon ./assets/appicon.png -name "AI-Native Dev" ./cmd/ai-dev-env-native

# Create a full-featured macOS .app bundle manually (more control)
macos-app: build-native
	@echo "Creating macOS .app bundle..."
	mkdir -p "AI-Native Dev.app/Contents/MacOS"
	mkdir -p "AI-Native Dev.app/Contents/Resources"
	cp bin/ai-dev-env-native "AI-Native Dev.app/Contents/MacOS/ai-dev-env-native"
	cp assets/Info.plist "AI-Native Dev.app/Contents/Info.plist"
	# If we had a proper icon, we would use:
	# cp assets/icon.icns "AI-Native Dev.app/Contents/Resources/icon.icns"
	@echo "Bundle created: AI-Native Dev.app"

# Run the web application
run-web:
	./bin/ai-dev-env

# Run the native application
run-native:
	./bin/ai-dev-env-native

# Install Fyne CLI tools (required for bundling)
install-fyne-cli:
	go install fyne.io/fyne/v2/cmd/fyne@latest

# Help
help:
	@echo "Available targets:"
	@echo "  all           - Build all applications (default)"
	@echo "  clean         - Remove build artifacts"
	@echo "  build         - Build both web and native applications"
	@echo "  build-web     - Build just the web application"
	@echo "  build-native  - Build just the native application"
	@echo "  bundle        - Create a macOS .app bundle using Fyne"
	@echo "  macos-app     - Create a more customized macOS .app bundle"
	@echo "  run-web       - Run the web application"
	@echo "  run-native    - Run the native application"
	@echo "  install-fyne-cli - Install the Fyne CLI tools"