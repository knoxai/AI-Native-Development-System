.PHONY: all clean build bundle macos-app

# Default target
all: build

# Clean the project
clean:
	rm -rf bin/*
	rm -rf *.app

# Build the native application
build:
	go build -o bin/ai-native-dev ./cmd/ai-native-dev

# Build a native macOS application bundle with fyne's tools
bundle:
	fyne package -os darwin -icon ./assets/appicon.png -name "AI-Native Dev" ./cmd/ai-native-dev

# Create a full-featured macOS .app bundle manually (more control)
macos-app: build
	@echo "Creating macOS .app bundle..."
	mkdir -p "AI-Native Dev.app/Contents/MacOS"
	mkdir -p "AI-Native Dev.app/Contents/Resources"
	cp bin/ai-native-dev "AI-Native Dev.app/Contents/MacOS/ai-native-dev"
	cp assets/Info.plist "AI-Native Dev.app/Contents/Info.plist"
	# If we had a proper icon, we would use:
	# cp assets/icon.icns "AI-Native Dev.app/Contents/Resources/icon.icns"
	@echo "Bundle created: AI-Native Dev.app"

# Run the native application
run:
	./bin/ai-native-dev

# Install Fyne CLI tools (required for bundling)
install-fyne-cli:
	go install fyne.io/fyne/v2/cmd/fyne@latest

# Help
help:
	@echo "Available targets:"
	@echo "  all           - Build the application (default)"
	@echo "  clean         - Remove build artifacts"
	@echo "  build         - Build the native application"
	@echo "  bundle        - Create a macOS .app bundle using Fyne"
	@echo "  macos-app     - Create a more customized macOS .app bundle"
	@echo "  run           - Run the native application"
	@echo "  install-fyne-cli - Install the Fyne CLI tools"