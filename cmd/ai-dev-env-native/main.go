package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
	
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	
	"github.com/knoxai/AI-Native-Development-System/pkg/ast"
	"github.com/knoxai/AI-Native-Development-System/pkg/filesystem"
	"github.com/knoxai/AI-Native-Development-System/pkg/intent"
	"github.com/knoxai/AI-Native-Development-System/pkg/llm"
	"github.com/knoxai/AI-Native-Development-System/pkg/semantics"
)

// AppState stores the global state of the application
type AppState struct {
	llmClient       *llm.Client
	intentProcessor *intent.Processor
	astProcessor    *ast.Processor
	semanticModel   *semantics.Model
	fileSystem      *filesystem.FileSystem
	selectedModel   string
	apiKey          string
	models          []llm.Model
	ui              *uiElements
	isDarkTheme     bool
}

// OpenRouter API models response structure
type OpenRouterModelsResponse struct {
	Data []OpenRouterModel `json:"data"`
}

type OpenRouterModel struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Created       int64  `json:"created"`
	Description   string `json:"description"`
	ContextLength int    `json:"context_length"`
}

// Cache for storing models with a timestamp
type ModelsCache struct {
	Models    []OpenRouterModel
	Timestamp time.Time
}

// Global cache
var modelsCache ModelsCache

// fetchAvailableModels retrieves models from OpenRouter API without requiring API key
// and returns a list of model IDs, automatically refreshing cache every 12 hours
func fetchAvailableModels() ([]string, error) {
	// Check if cache is still valid (less than 12 hours old)
	if !modelsCache.Timestamp.IsZero() && time.Since(modelsCache.Timestamp) < 12*time.Hour && len(modelsCache.Models) > 0 {
		// Use cached models
		modelIDs := make([]string, len(modelsCache.Models))
		for i, model := range modelsCache.Models {
			modelIDs[i] = model.ID
		}
		return modelIDs, nil
	}

	// Cache expired or empty, fetch new data
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://openrouter.ai/api/v1/models", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var modelsResp OpenRouterModelsResponse
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return nil, err
	}

	// Update cache
	modelsCache.Models = modelsResp.Data
	modelsCache.Timestamp = time.Now()

	// Extract model IDs
	modelIDs := make([]string, len(modelsResp.Data))
	for i, model := range modelsResp.Data {
		modelIDs[i] = model.ID
	}

	return modelIDs, nil
}

// checkOpenRouterConnectivity checks if we can connect to the OpenRouter API
func checkOpenRouterConnectivity() error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://openrouter.ai/api/v1/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API responded with status code: %d", resp.StatusCode)
	}

	return nil
}

func main() {
	// Configure logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	fmt.Println("Starting AI-Native Development Environment...")
	
	// Initialize file system
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user home directory: %v", err)
	}
	
	workspaceDir := filepath.Join(homeDir, "AI-Native-Workspace")
	fs, err := filesystem.New(workspaceDir)
	if err != nil {
		log.Fatalf("Failed to initialize file system: %v", err)
	}
	
	// Initialize app state
	appState := &AppState{
		selectedModel: "openai/gpt-3.5-turbo", // Default model
		apiKey:        os.Getenv("OPENROUTER_API_KEY"),
		fileSystem:    fs,
		isDarkTheme:   true, // Default to dark theme
	}
	
	// Initialize the semantic model
	appState.semanticModel = semantics.NewModel()
	
	// Initialize the AST processor
	appState.astProcessor = ast.NewProcessor(appState.semanticModel)
	
	// Initialize the intent processor
	appState.intentProcessor = intent.NewProcessor(appState.astProcessor, appState.semanticModel)
	
	// Initialize LLM client if API key is available
	if appState.apiKey != "" {
		// Check connectivity to OpenRouter
		connErr := checkOpenRouterConnectivity()
		if connErr != nil {
			log.Printf("Warning: Cannot connect to OpenRouter API: %v", connErr)
			fmt.Println("Warning: Cannot connect to OpenRouter API - check your internet connection")
		}
		
		client, err := llm.NewClient()
		if err == nil {
			appState.llmClient = client
			appState.intentProcessor.SetLLMClient(client)
			fmt.Println("OpenRouter API key found - AI code generation is enabled")
		} else {
			log.Printf("Error initializing LLM client: %v", err)
			fmt.Println("Error: Failed to initialize LLM client - check your API key")
		}
	} else {
		fmt.Println("Note: OpenRouter API key not found - AI code generation requires an API key")
		fmt.Println("Set the OPENROUTER_API_KEY environment variable to enable AI code generation")
	}
	
	// Create Fyne app
	a := app.New()
	
	// Custom dark theme for a code-focused environment
	a.Settings().SetTheme(newCodeTheme())
	
	// Set app metadata
	a.SetIcon(resourceAppIconPng)
	
	// Create main window
	w := a.NewWindow("AI-Native Development Environment")
	w.Resize(fyne.NewSize(1200, 800))
	
	// Setup main menu
	setupMainMenu(w, appState)
	
	// Setup keyboard shortcuts if we're on desktop
	setupKeyboardShortcuts(w, appState)
	
	// Create UI
	appUI := createUI(w, appState)
	
	// Set window content
	w.SetContent(appUI)
	
	// Start the app
	w.ShowAndRun()
}

// setupMainMenu creates the application menu
func setupMainMenu(w fyne.Window, state *AppState) {
	// File menu
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("New Project", func() {
			createNewProject(w, state)
		}),
		fyne.NewMenuItem("Open Project", func() {
			openProject(w, state)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Save Output", func() {
			saveOutput(w, state)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Exit", func() {
			w.Close()
		}),
	)
	
	// Edit menu
	editMenu := fyne.NewMenu("Edit",
		fyne.NewMenuItem("Copy", func() {
			w.Clipboard().SetContent(getSelectedText(state))
		}),
		fyne.NewMenuItem("Paste", func() {
			// Not implemented yet
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Settings", func() {
			showSettings(w, state)
		}),
	)
	
	// View menu
	viewMenu := fyne.NewMenu("View",
		fyne.NewMenuItem("Toggle Theme", func() {
			toggleTheme(w, state)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Zoom In", func() {
			// Not implemented yet
		}),
		fyne.NewMenuItem("Zoom Out", func() {
			// Not implemented yet
		}),
		fyne.NewMenuItem("Reset Zoom", func() {
			// Not implemented yet
		}),
	)
	
	// Models menu
	modelsMenu := fyne.NewMenu("Models",
		fyne.NewMenuItem("Model Information", func() {
			showModelInfo(w, state)
		}),
		fyne.NewMenuItem("Refresh Models List", func() {
			refreshModelsList(w, state)
		}),
	)
	
	// Help menu
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", func() {
			dialog.ShowInformation("About AI-Native Development Environment", 
				"AI-Native Development Environment v1.0\n\nThis application allows for intent-based code generation and manipulation through abstract syntax trees and semantic models.", 
				w)
		}),
		fyne.NewMenuItem("Documentation", func() {
			// Open documentation (to be implemented)
		}),
	)
	
	// Set the main menu
	w.SetMainMenu(fyne.NewMainMenu(
		fileMenu,
		editMenu,
		viewMenu,
		modelsMenu,
		helpMenu,
	))
}

// setupKeyboardShortcuts sets up keyboard shortcuts for desktop platforms
func setupKeyboardShortcuts(w fyne.Window, state *AppState) {
	// Ctrl+N - New Project
	w.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: fyne.KeyModifierControl},
		func(shortcut fyne.Shortcut) {
			createNewProject(w, state)
		},
	)
	
	// Ctrl+O - Open Project
	w.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: fyne.KeyModifierControl},
		func(shortcut fyne.Shortcut) {
			openProject(w, state)
		},
	)
	
	// Ctrl+S - Save Output
	w.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl},
		func(shortcut fyne.Shortcut) {
			saveOutput(w, state)
		},
	)
	
	// Ctrl+E - Execute Intent
	w.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyE, Modifier: fyne.KeyModifierControl},
		func(shortcut fyne.Shortcut) {
			if state.ui.intentInput != nil && state.ui.intentInput.Text != "" {
				executeIntent(state.ui.intentInput.Text, state, w)
			}
		},
	)
	
	// Ctrl+T - Toggle Theme
	w.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyT, Modifier: fyne.KeyModifierControl},
		func(shortcut fyne.Shortcut) {
			toggleTheme(w, state)
		},
	)
}

// createNewProject shows a dialog to create a new project
func createNewProject(w fyne.Window, state *AppState) {
	// Create entry for project name
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Project Name")
	
	// Show dialog
	dialog.ShowForm("Create New Project", "Create", "Cancel", 
		[]*widget.FormItem{
			widget.NewFormItem("Project Name", nameEntry),
		},
		func(submit bool) {
			if submit {
				projectName := nameEntry.Text
				if projectName == "" {
					dialog.ShowError(fmt.Errorf("Project name cannot be empty"), w)
					return
				}
				
				// Create the workspace
				err := state.fileSystem.CreateWorkspace(projectName)
				if err != nil {
					dialog.ShowError(fmt.Errorf("Failed to create project: %v", err), w)
					return
				}
				
				dialog.ShowInformation("Project Created", 
					fmt.Sprintf("Project '%s' has been created at %s", 
						projectName, 
						filepath.Join(state.fileSystem.WorkingDirectory, projectName)),
					w)
				
				// Update status
				if state.ui.statusBar != nil {
					state.ui.statusBar.SetText(fmt.Sprintf("Project '%s' created", projectName))
				}
			}
		}, w)
}

// openProject shows a dialog to open an existing project
func openProject(w fyne.Window, state *AppState) {
	// Create a file dialog
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if uri == nil {
			return
		}
		
		path := uri.Path()
		err = state.fileSystem.SetWorkingDirectory(path)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to open project: %v", err), w)
			return
		}
		
		// Update status
		if state.ui.statusBar != nil {
			state.ui.statusBar.SetText(fmt.Sprintf("Project opened at %s", path))
		}
		
	}, w)
}

// saveOutput saves the generated code to a file
func saveOutput(w fyne.Window, state *AppState) {
	if state.ui.codeOutput == nil || state.ui.codeOutput.Text == "" {
		dialog.ShowInformation("No Output", "There is no generated code to save.", w)
		return
	}
	
	// Create a file dialog
	fd := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()
		
		// Write the content to the file
		_, err = writer.Write([]byte(state.ui.codeOutput.Text))
		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to save file: %v", err), w)
			return
		}
		
		// Update status
		if state.ui.statusBar != nil {
			state.ui.statusBar.SetText(fmt.Sprintf("Code saved to %s", writer.URI().Path()))
		}
	}, w)
	
	// Set default file name based on content analysis
	fd.SetFileName("generated_code.go")
	
	// Set filter for common code file types
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".go", ".py", ".js", ".java", ".cs", ".cpp", ".h"}))
	
	fd.Show()
}

// showSettings displays the settings dialog
func showSettings(w fyne.Window, state *AppState) {
	// API Configuration section with improved styling
	apiConfigLabel := widget.NewLabelWithStyle("API Configuration", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	
	// API key input with better styling
	apiKeyInput := widget.NewPasswordEntry()
	apiKeyInput.SetPlaceHolder("Enter OpenRouter API key")
	if state.apiKey != "" {
		apiKeyInput.SetText(state.apiKey)
	}
	
	// Create a field container with label
	apiKeyLabel := widget.NewLabelWithStyle("API Key:", fyne.TextAlignLeading, fyne.TextStyle{})
	apiKeyContainer := container.NewBorder(
		nil, nil, apiKeyLabel, nil,
		apiKeyInput,
	)
	
	// Save API key button with visual improvements
	saveButton := widget.NewButtonWithIcon("Save API Key", theme.ConfirmIcon(), func() {
		if apiKeyInput.Text == "" {
			dialog.ShowInformation("API Key Required", "Please enter an OpenRouter API key", w)
			return
		}
		
		// Show saving progress
		progress := dialog.NewProgress("Saving API Key", "Verifying API key...", w)
		progress.Show()
		
		// Perform the save asynchronously
		go func() {
			oldKey := state.apiKey
			state.apiKey = apiKeyInput.Text
			
			// Create a temporary client to test the key
			client := &llm.Client{
				APIKey:       state.apiKey,
				DefaultModel: state.selectedModel,
				HTTPClient:   &http.Client{},
			}
			
			// Test the connection
			if _, err := client.GetAvailableModels(); err != nil {
				// Reset to old key if there's an error
				state.apiKey = oldKey
				progress.Hide()
				dialog.ShowError(fmt.Errorf("Invalid API key: %v", err), w)
				return
			}
			
			// If successful, update the state
			state.llmClient = client
			state.intentProcessor.SetLLMClient(client)
			
			progress.Hide()
			dialog.ShowInformation("API Key Saved", "Your API key has been verified and saved. AI code generation is now enabled.", w)
			
			// Update status bar
			if state.ui != nil && state.ui.statusBar != nil {
				state.ui.statusBar.SetText("API key verified and saved")
			}
		}()
	})
	
	// Create model selector with improved appearance
	modelSelectorLabel := widget.NewLabel("Model:")
	modelSelector := createModelSelector(state)
	
	// Create a container for the model selector
	modelSelectorContainer := container.NewBorder(
		nil, nil, modelSelectorLabel, nil,
		modelSelector,
	)
	
	// Model selector info label with improved styling
	modelInfoLabel := widget.NewLabelWithStyle(
		"Models are automatically fetched from OpenRouter API",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)
	
	// Create a refresh button for the models list
	refreshModelsBtn := widget.NewButtonWithIcon("Refresh Models", theme.ViewRefreshIcon(), func() {
		refreshModelsList(w, state)
	})
	
	// Create a button container
	buttonContainer := container.NewHBox(
		saveButton, 
		layout.NewSpacer(),
		refreshModelsBtn,
	)
	
	// Create a separator for visual distinction
	separator := widget.NewSeparator()
	
	// API settings container with improved layout
	apiSettings := container.NewPadded(
		container.NewVBox(
			apiConfigLabel,
			widget.NewSeparator(),
			container.NewPadded(
				container.NewVBox(
					apiKeyContainer,
					container.NewPadded(buttonContainer),
					separator,
					container.NewPadded(modelSelectorContainer),
					container.NewPadded(modelInfoLabel),
				),
			),
		),
	)

	// Create settings dialog content
	content := container.NewVBox(
		apiSettings,
		container.NewHBox(
			widget.NewLabel("Theme:"),
			widget.NewSelect([]string{"Dark", "Light"}, func(value string) {
				// Update theme when selection changes
				newTheme := value == "Dark"
				if newTheme != state.isDarkTheme {
					state.isDarkTheme = newTheme
					applyTheme(w, state)
				}
			}),
		),
	)
	
	// Show the dialog with the content
	dialog.ShowCustom("Settings", "Close", content, w)
}

// toggleTheme switches between dark and light themes
func toggleTheme(w fyne.Window, state *AppState) {
	state.isDarkTheme = !state.isDarkTheme
	applyTheme(w, state)
}

// applyTheme applies the current theme to the window
func applyTheme(w fyne.Window, state *AppState) {
	if state.isDarkTheme {
		fyne.CurrentApp().Settings().SetTheme(newCodeTheme())
	} else {
		fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())
	}
	
	// Update status
	if state.ui != nil && state.ui.statusBar != nil {
		themeStr := "dark"
		if !state.isDarkTheme {
			themeStr = "light"
		}
		state.ui.statusBar.SetText(fmt.Sprintf("Theme switched to %s", themeStr))
	}
}

// getSelectedText returns the currently selected text (if any)
func getSelectedText(state *AppState) string {
	// This would ideally get the selected text from any focused widget
	// For now, it's a stub
	return ""
}

// createUI builds the complete user interface
func createUI(w fyne.Window, state *AppState) fyne.CanvasObject {
	// Apply custom theme settings
	fyne.CurrentApp().Settings().SetTheme(newCodeTheme())
	
	// Header with logo and title - with better styling
	header := createHeader(state)
	
	// API Configuration section with improved styling
	apiConfigLabel := widget.NewLabelWithStyle("API Configuration", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	
	// API key input with better styling
	apiKeyInput := widget.NewPasswordEntry()
	apiKeyInput.SetPlaceHolder("Enter OpenRouter API key")
	if state.apiKey != "" {
		apiKeyInput.SetText(state.apiKey)
	}
	
	// Create a field container with label
	apiKeyLabel := widget.NewLabelWithStyle("API Key:", fyne.TextAlignLeading, fyne.TextStyle{})
	apiKeyContainer := container.NewBorder(
		nil, nil, apiKeyLabel, nil,
		apiKeyInput,
	)
	
	// Save API key button with visual improvements
	saveButton := widget.NewButtonWithIcon("Save API Key", theme.ConfirmIcon(), func() {
		if apiKeyInput.Text == "" {
			dialog.ShowInformation("API Key Required", "Please enter an OpenRouter API key", w)
			return
		}
		
		// Show saving progress
		progress := dialog.NewProgress("Saving API Key", "Verifying API key...", w)
		progress.Show()
		
		// Perform the save asynchronously
		go func() {
			oldKey := state.apiKey
			state.apiKey = apiKeyInput.Text
			
			// Create a temporary client to test the key
			client := &llm.Client{
				APIKey:       state.apiKey,
				DefaultModel: state.selectedModel,
				HTTPClient:   &http.Client{},
			}
			
			// Test the connection
			if _, err := client.GetAvailableModels(); err != nil {
				// Reset to old key if there's an error
				state.apiKey = oldKey
				progress.Hide()
				dialog.ShowError(fmt.Errorf("Invalid API key: %v", err), w)
				return
			}
			
			// If successful, update the state
			state.llmClient = client
			state.intentProcessor.SetLLMClient(client)
			
			progress.Hide()
			dialog.ShowInformation("API Key Saved", "Your API key has been verified and saved. AI code generation is now enabled.", w)
			
			// Update status bar
			if state.ui != nil && state.ui.statusBar != nil {
				state.ui.statusBar.SetText("API key verified and saved")
			}
		}()
	})
	
	// Create model selector with improved appearance
	modelSelectorLabel := widget.NewLabelWithStyle("Model:", fyne.TextAlignLeading, fyne.TextStyle{})
	modelSelector := createModelSelector(state)
	
	// Create a container for the model selector
	modelSelectorContainer := container.NewBorder(
		nil, nil, modelSelectorLabel, nil,
		modelSelector,
	)
	
	// Model selector info label with improved styling
	modelInfoLabel := widget.NewLabelWithStyle(
		"Models are automatically fetched from OpenRouter API",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)
	
	// Create a refresh button for the models list
	refreshModelsBtn := widget.NewButtonWithIcon("Refresh Models", theme.ViewRefreshIcon(), func() {
		refreshModelsList(w, state)
	})
	
	// Create a button container
	buttonContainer := container.NewHBox(
		saveButton, 
		layout.NewSpacer(),
		refreshModelsBtn,
	)
	
	// Create a separator for visual distinction
	separator := widget.NewSeparator()
	
	// API settings container with improved layout
	apiSettings := container.NewPadded(
		container.NewVBox(
			apiConfigLabel,
			widget.NewSeparator(),
			container.NewPadded(
				container.NewVBox(
					apiKeyContainer,
					container.NewPadded(buttonContainer),
					separator,
					container.NewPadded(modelSelectorContainer),
					container.NewPadded(modelInfoLabel),
				),
			),
		),
	)
	
	// Create a styled background for the API settings section
	apiSettingsBackground := canvas.NewRectangle(theme.BackgroundColor())
	apiSettingsCard := container.NewMax(
		apiSettingsBackground,
		container.NewPadded(apiSettings),
	)
	
	// File explorer with improved styling
	fileExplorerLabel := widget.NewLabelWithStyle("Project Files", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	fileExplorer := NewFileExplorer(state)
	
	fileExplorerCard := container.NewBorder(
		fileExplorerLabel,
		nil, nil, nil,
		container.NewPadded(fileExplorer.Container()),
	)
	
	// File content display with improved styling
	filePathLabel := widget.NewLabel("No file selected")
	filePathLabel.Alignment = fyne.TextAlignLeading
	filePathLabel.TextStyle = fyne.TextStyle{Italic: true}
	
	fileContentDisplay := widget.NewMultiLineEntry()
	fileContentDisplay.Disable() // Read-only
	fileContentDisplay.TextStyle = fyne.TextStyle{Monospace: true}
	
	// File content container with improved styling
	fileContentBackground := canvas.NewRectangle(color.NRGBA{R: 20, G: 20, B: 20, A: 255})
	fileContentContainer := container.NewMax(
		fileContentBackground,
		container.NewBorder(
			filePathLabel,
			nil, nil, nil,
			container.NewScroll(fileContentDisplay),
		),
	)
	
	// Left panel with file explorer and content - with proper sizing
	leftPanel := container.NewVSplit(
		fileExplorerCard,
		fileContentContainer,
	)
	leftPanel.Offset = 0.35
	
	// Intent input with improved styling
	intentLabel := widget.NewLabelWithStyle("Enter your development intent:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	
	// Create a helper text label
	helperText := widget.NewLabel("Example: Create a login function that validates user credentials and returns a token")
	helperText.TextStyle = fyne.TextStyle{Italic: true}
	helperText.Wrapping = fyne.TextWrapWord
	
	// Create the intent input field with better styling
	intentInput := widget.NewMultiLineEntry()
	intentInput.SetPlaceHolder("Type your intent here...")
	intentInput.Wrapping = fyne.TextWrapWord
	intentInput.MultiLine = true
	
	// Create a stylish background for the input field
	intentInputBackground := canvas.NewRectangle(color.NRGBA{R: 25, G: 25, B: 25, A: 255})
	
	// Create a container with fixed height for the input field
	intentScrollContainer := container.NewScroll(intentInput)
	intentScrollContainer.SetMinSize(fyne.NewSize(0, 120)) 
	
	// Add a border around the input field to make it stand out
	intentBorder := container.NewMax(
		intentInputBackground,
		container.NewPadded(intentScrollContainer),
	)
	
	// Create a more professional looking execute button
	executeButton := widget.NewButtonWithIcon("Execute Intent", theme.ConfirmIcon(), func() {
		executeIntent(intentInput.Text, state, w)
	})
	executeButton.Importance = widget.HighImportance // Highlight the button
	executeButton.Resize(fyne.NewSize(150, 36))      // Make button more prominent
	
	// Create a button container with right alignment
	buttonContainer = container.NewHBox(
		layout.NewSpacer(),
		executeButton,
	)
	
	// Intent container with improved layout
	intentContainer := container.NewPadded(
		container.NewVBox(
			container.NewPadded(
				container.NewVBox(
					intentLabel,
					helperText,
				),
			),
			intentBorder,
			widget.NewSeparator(), // Add a separator for visual distinction
			container.NewPadded(buttonContainer),
		),
	)
	
	// Create a stylish background for the intent container
	intentBackground := canvas.NewRectangle(theme.BackgroundColor())
	intentCard := container.NewMax(
		intentBackground,
		intentContainer,
	)
	
	// Output tabs with improved styling
	// Code output area with improved styling
	codeOutputLabel := widget.NewLabelWithStyle("Generated Code", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	codeOutput := widget.NewMultiLineEntry()
	codeOutput.Disable() // Read-only
	codeOutput.Wrapping = fyne.TextWrapWord
	codeOutput.TextStyle = fyne.TextStyle{Monospace: true}
	
	// Create a copy button with label
	copyCodeBtn := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		if codeOutput.Text != "" {
			w.Clipboard().SetContent(codeOutput.Text)
			state.ui.statusBar.SetText("Code copied to clipboard")
		}
	})
	
	// Create a header with label and buttons
	codeOutputHeader := container.NewBorder(
		nil, nil, 
		codeOutputLabel,
		copyCodeBtn,
	)
	
	// Create a stylish background for code output
	codeOutputBackground := canvas.NewRectangle(color.NRGBA{R: 22, G: 22, B: 22, A: 255})
	
	// Create a scroll container with increased height
	codeScrollContainer := container.NewScroll(codeOutput)
	
	// Use Card container for a more professional look with background
	codeOutputContainer := container.NewMax(
		codeOutputBackground,
		container.NewBorder(
			codeOutputHeader,
			nil, nil, nil,
			container.NewPadded(codeScrollContainer),
		),
	)
	
	// AST view area with improved styling
	astOutputLabel := widget.NewLabelWithStyle("AST Representation", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	astOutput := widget.NewMultiLineEntry()
	astOutput.Disable() // Read-only
	astOutput.Wrapping = fyne.TextWrapWord
	astOutput.TextStyle = fyne.TextStyle{Monospace: true}
	
	// Create a copy button with label
	copyAstBtn := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		if astOutput.Text != "" {
			w.Clipboard().SetContent(astOutput.Text)
			state.ui.statusBar.SetText("AST copied to clipboard")
		}
	})
	
	// Create a header with label and buttons
	astOutputHeader := container.NewBorder(
		nil, nil, 
		astOutputLabel,
		copyAstBtn,
	)
	
	// Create a stylish background for AST output
	astOutputBackground := canvas.NewRectangle(color.NRGBA{R: 22, G: 22, B: 22, A: 255})
	
	// Create a scroll container with increased height
	astScrollContainer := container.NewScroll(astOutput)
	
	// Use Card container for a more professional look with background
	astOutputContainer := container.NewMax(
		astOutputBackground,
		container.NewBorder(
			astOutputHeader,
			nil, nil, nil,
			container.NewPadded(astScrollContainer),
		),
	)
	
	// Semantic model view area with improved styling
	semanticOutputLabel := widget.NewLabelWithStyle("Semantic Model", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	semanticOutput := widget.NewMultiLineEntry()
	semanticOutput.Disable() // Read-only
	semanticOutput.Wrapping = fyne.TextWrapWord
	semanticOutput.TextStyle = fyne.TextStyle{Monospace: true}
	
	// Create a copy button with label
	copySemanticBtn := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		if semanticOutput.Text != "" {
			w.Clipboard().SetContent(semanticOutput.Text)
			state.ui.statusBar.SetText("Semantic model copied to clipboard")
		}
	})
	
	// Create a header with label and buttons
	semanticOutputHeader := container.NewBorder(
		nil, nil, 
		semanticOutputLabel,
		copySemanticBtn,
	)
	
	// Create a stylish background for semantic output
	semanticOutputBackground := canvas.NewRectangle(color.NRGBA{R: 22, G: 22, B: 22, A: 255})
	
	// Create a scroll container with increased height
	semanticScrollContainer := container.NewScroll(semanticOutput)
	
	// Use Card container for a more professional look with background
	semanticOutputContainer := container.NewMax(
		semanticOutputBackground,
		container.NewBorder(
			semanticOutputHeader,
			nil, nil, nil,
			container.NewPadded(semanticScrollContainer),
		),
	)
	
	// Create tabs for different views with improved styling
	tabs := container.NewAppTabs(
		container.NewTabItem("Code", codeOutputContainer),
		container.NewTabItem("AST", astOutputContainer),
		container.NewTabItem("Semantics", semanticOutputContainer),
	)
	tabs.SetTabLocation(container.TabLocationTop) // Change to top tabs for better visibility
	
	// Add event listener to select the Code tab when content is generated
	tabs.OnSelected = func(tab *container.TabItem) {
		// This ensures tabs will display correctly when switching between them
		tab.Content.Refresh()
	}
	
	// Right panel with intent and output - give the tabs more space 
	// Using a responsive VSplit container
	rightPanel := container.NewVSplit(
		intentCard,
		tabs,
	)
	rightPanel.Offset = 0.3 // Give the output tabs more space (70% of the panel)
	
	// Create a modern, professional status bar
	statusIcon := widget.NewIcon(theme.InfoIcon())
	statusMessage := widget.NewLabelWithStyle("Ready", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	modelInfo := widget.NewLabel("Model: " + state.selectedModel)
	
	// Format current date with more detail
	currentTime := widget.NewLabelWithStyle(
		time.Now().Format("Jan 2, 2006 15:04"),
		fyne.TextAlignTrailing,
		fyne.TextStyle{},
	)
	
	// Create an improved status bar with multiple sections and better styling
	statusBackground := canvas.NewRectangle(color.NRGBA{R: 40, G: 40, B: 45, A: 255})
	statusContainer := container.NewMax(
		statusBackground,
		container.NewPadded(
			container.NewHBox(
				statusIcon,
				statusMessage,
				layout.NewSpacer(),
				modelInfo,
				layout.NewSpacer(),
				currentTime,
			),
		),
	)
	
	// Create a separator line above the status bar
	statusSeparator := canvas.NewLine(theme.ForegroundColor())
	statusSeparator.StrokeWidth = 1
	
	// Wrap the status components in a container
	statusBarWrapper := container.NewBorder(
		statusSeparator, 
		nil, nil, nil, 
		statusContainer,
	)
	
	// Main content area with Split container for left and right panels 
	// Use a responsive HSplit container
	mainContent := container.NewHSplit(
		leftPanel,
		rightPanel,
	)
	mainContent.Offset = 0.25 // Adjust split position for optimal layout
	
	// Create a background for the main content
	mainBackground := canvas.NewRectangle(theme.BackgroundColor())
	
	// Main layout with improved styling and responsiveness
	content := container.NewMax(
		mainBackground,
		container.NewBorder(
			container.NewVBox(
				header,
				apiSettingsCard,
			),
			statusBarWrapper,
			nil,
			nil,
			container.NewPadded(mainContent),
		),
	)
	
	// Store UI elements in the state for later access
	state.ui = &uiElements{
		statusBar:          statusMessage,
		codeOutput:         codeOutput,
		astOutput:          astOutput,
		semanticOutput:     semanticOutput,
		modelSelector:      modelSelector,
		intentInput:        intentInput,
		fileExplorer:       fileExplorer,
		fileContentDisplay: fileContentDisplay,
		filePathLabel:      filePathLabel,
	}
	
	return content
}

// createHeader builds the app header with logo and title
func createHeader(state *AppState) fyne.CanvasObject {
	logo := canvas.NewImageFromResource(resourceLogoJpg)
	logo.SetMinSize(fyne.NewSize(50, 50))
	
	title := widget.NewLabelWithStyle(
		"AI-Native Development Environment",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
	
	subtitle := widget.NewLabel("Direct AST and semantic model manipulation")
	
	return container.NewHBox(
		logo,
		container.NewVBox(
			title,
			subtitle,
		),
		layout.NewSpacer(),
	)
}

// createModelSelector builds the model selection dropdown
func createModelSelector(state *AppState) *widget.Select {
	// Start with default model
	modelNames := []string{state.selectedModel}
	
	// Create selector with default model
	selector := widget.NewSelect(modelNames, func(selected string) {
		state.selectedModel = selected
		if state.llmClient != nil {
			state.llmClient.SetModel(selected)
		}
	})
	
	selector.Selected = state.selectedModel
	
	// Asynchronously fetch models from API and update the selector
	go func() {
		modelIDs, err := fetchAvailableModels()
		if err != nil {
			log.Printf("Failed to fetch models: %v", err)
			return
		}
		
		// Update UI on the main thread
		if len(modelIDs) > 0 {
			// Update the selector options
			selector.Options = modelIDs
			
			// If the current selection isn't in the new options, reset to default
			found := false
			for _, name := range modelIDs {
				if name == state.selectedModel {
					found = true
					break
				}
			}
			
			if !found && len(modelIDs) > 0 {
				state.selectedModel = modelIDs[0]
				selector.Selected = modelIDs[0]
				if state.llmClient != nil {
					state.llmClient.SetModel(modelIDs[0])
				}
			}
			
			selector.Refresh()
		}
	}()
	
	return selector
}

// executeIntent processes the user's intent and updates the UI with the results
func executeIntent(intentText string, state *AppState, w fyne.Window) {
	if intentText == "" {
		dialog.ShowError(fmt.Errorf("Please enter a development intent"), w)
		return
	}
	
	if state.llmClient == nil {
		dialog.ShowInformation("API Key Required", 
			"An OpenRouter API key is required for intent processing. Please enter your API key in the settings above.", 
			w)
		
		// Show helpful information in the code output area
		if state.ui.codeOutput != nil {
			state.ui.codeOutput.SetText(`// API Key Required
//
// To process your intent, you need to provide an OpenRouter API key.
// 
// 1. Get an API key from https://openrouter.ai
// 2. Enter it in the "API Configuration" section above
// 3. Click "Save API Key"
// 4. Try again with your intent
//
// Note: This application automatically fetches available models from OpenRouter,
// so you don't need to manually fetch them.`)
		}
		
		return
	}
	
	// Show loading dialog
	progress := dialog.NewProgress("Processing Intent", "Analyzing your development intent...", w)
	progress.Show()
	
	// Update status
	state.ui.statusBar.SetText("Processing intent...")
	
	// Start asynchronous operation
	go func() {
		// Parse the intent with timeout and error handling
		var parsedIntent interface{}
		var parseErr error
		
		// Create a timeout channel
		parseTimeout := time.After(30 * time.Second)
		parseComplete := make(chan bool, 1)
		
		// Execute intent parsing in a separate goroutine
		go func() {
			parsedIntent, parseErr = state.intentProcessor.ParseIntent(intentText)
			parseComplete <- true
		}()
		
		// Wait for either completion or timeout
		select {
		case <-parseComplete:
			// Continue processing
		case <-parseTimeout:
			progress.Hide()
			dialog.ShowError(fmt.Errorf("Intent parsing timed out after 30 seconds"), w)
			state.ui.statusBar.SetText("Error: Intent parsing timed out")
			return
		}
		
		// Check for parse errors
		if parseErr != nil {
			progress.Hide()
			log.Printf("Intent parsing error: %v", parseErr)
			dialog.ShowError(fmt.Errorf("Failed to parse intent: %v", parseErr), w)
			state.ui.statusBar.SetText("Error: Failed to parse intent")
			
			// Still show something in the output areas for debugging
			state.ui.codeOutput.SetText("// Intent parsing failed. Please check the following:\n" + 
				"// 1. Your API key is valid and has not expired\n" + 
				"// 2. The selected model is available\n" + 
				"// 3. Your intent is clear and well-formed\n\n" + 
				"// Error: " + parseErr.Error())
			return
		}
		
		// Execute the intent with timeout
		var result interface{}
		var execErr error
		
		// Create a timeout channel for execution
		execTimeout := time.After(60 * time.Second)
		execComplete := make(chan bool, 1)
		
		// Execute intent in a separate goroutine
		go func() {
			// Type assertion for parsedIntent
			intentPtr, ok := parsedIntent.(*intent.Intent)
			if !ok {
				execErr = fmt.Errorf("unexpected intent type: %T", parsedIntent)
				execComplete <- true
				return
			}
			result, execErr = state.intentProcessor.ExecuteIntent(intentPtr)
			execComplete <- true
		}()
		
		// Wait for either completion or timeout
		select {
		case <-execComplete:
			// Continue processing
		case <-execTimeout:
			progress.Hide()
			dialog.ShowError(fmt.Errorf("Intent execution timed out after 60 seconds"), w)
			state.ui.statusBar.SetText("Error: Intent execution timed out")
			return
		}
		
		// Update UI after execution is complete
		progress.Hide()
		
		if execErr != nil {
			log.Printf("Intent execution error: %v", execErr)
			dialog.ShowError(fmt.Errorf("Failed to execute intent: %v", execErr), w)
			state.ui.statusBar.SetText("Error: Failed to execute intent")
			
			// Show error in output for debugging
			state.ui.codeOutput.SetText("// Intent execution failed.\n" + 
				"// Error: " + execErr.Error())
			return
		}
		
		// Handle the result
		if resultMap, ok := result.(map[string]interface{}); ok {
			// Update code output
			if code, ok := resultMap["code"].(string); ok && code != "" {
				state.ui.codeOutput.SetText(code)
			} else {
				state.ui.codeOutput.SetText("// No code was generated for this intent")
			}
			
			// Update AST output
			if ast, ok := resultMap["ast"].(string); ok && ast != "" {
				state.ui.astOutput.SetText(ast)
			} else {
				state.ui.astOutput.SetText("// No AST representation was generated")
			}
			
			// Update semantic output
			if semantics, ok := resultMap["semantics"].(string); ok && semantics != "" {
				state.ui.semanticOutput.SetText(semantics)
			} else {
				state.ui.semanticOutput.SetText("// No semantic model was generated")
			}
			
			state.ui.statusBar.SetText("Intent processed successfully")
		} else if resultMap, ok := result.(map[string]string); ok {
			// Handle string-based map (alternative response format)
			// Update code output
			if code, ok := resultMap["code"]; ok && code != "" {
				state.ui.codeOutput.SetText(code)
			} else {
				state.ui.codeOutput.SetText("// No code was generated for this intent")
			}
			
			// Update AST output
			if ast, ok := resultMap["ast"]; ok && ast != "" {
				state.ui.astOutput.SetText(ast)
			} else {
				state.ui.astOutput.SetText("// No AST representation was generated")
			}
			
			// Update semantic output
			if semantics, ok := resultMap["semantics"]; ok && semantics != "" {
				state.ui.semanticOutput.SetText(semantics)
			} else {
				state.ui.semanticOutput.SetText("// No semantic model was generated")
			}
			
			state.ui.statusBar.SetText("Intent processed successfully")
		} else {
			// Handle unexpected result format
			log.Printf("Unexpected result format: %T", result)
			state.ui.statusBar.SetText("Intent processed, but result format is unexpected")
			
			// Try to convert the result to a string-based map if possible
			if strResult, ok := convertToStringMap(result); ok {
				// Update code output
				if code, ok := strResult["code"]; ok && code != "" {
					state.ui.codeOutput.SetText(code)
				} else {
					state.ui.codeOutput.SetText("// No code was generated for this intent")
				}
				
				// Update AST output
				if ast, ok := strResult["ast"]; ok && ast != "" {
					state.ui.astOutput.SetText(ast)
				} else {
					state.ui.astOutput.SetText("// No AST representation was generated")
				}
				
				// Update semantic output
				if semantics, ok := strResult["semantics"]; ok && semantics != "" {
					state.ui.semanticOutput.SetText(semantics)
				} else {
					state.ui.semanticOutput.SetText("// No semantic model was generated")
				}
				
				state.ui.statusBar.SetText("Intent processed successfully")
			} else {
				// Last resort: try to display anything useful
				if result != nil {
					resultJSON, err := json.MarshalIndent(result, "", "  ")
					if err == nil {
						state.ui.codeOutput.SetText("// Result in unexpected format. Raw output:\n\n" + string(resultJSON))
					} else {
						state.ui.codeOutput.SetText(fmt.Sprintf("// Result in unexpected format: %v", result))
					}
				} else {
					state.ui.codeOutput.SetText("// No result was returned from the model")
				}
			}
		}
	}()
}

// convertToStringMap attempts to convert various result formats to a map[string]string
func convertToStringMap(result interface{}) (map[string]string, bool) {
	// Try to handle different output formats
	strMap := make(map[string]string)
	
	// Case 1: map[string]interface{} - convert values to strings
	if mapResult, ok := result.(map[string]interface{}); ok {
		for k, v := range mapResult {
			if strVal, ok := v.(string); ok {
				strMap[k] = strVal
			} else {
				// Try to convert to JSON
				if jsonBytes, err := json.Marshal(v); err == nil {
					strMap[k] = string(jsonBytes)
				} else {
					strMap[k] = fmt.Sprintf("%v", v)
				}
			}
		}
		return strMap, true
	}
	
	// Case 2: Already string map
	if strMapResult, ok := result.(map[string]string); ok {
		return strMapResult, true
	}
	
	// Case 3: Maybe a struct we can marshal to JSON
	if jsonBytes, err := json.Marshal(result); err == nil {
		// Try to unmarshal as a map
		var objMap map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &objMap); err == nil {
			for k, v := range objMap {
				if strVal, ok := v.(string); ok {
					strMap[k] = strVal
				} else {
					// Nested JSON
					if nestedJSON, err := json.Marshal(v); err == nil {
						strMap[k] = string(nestedJSON)
					} else {
						strMap[k] = fmt.Sprintf("%v", v)
					}
				}
			}
			return strMap, true
		}
	}
	
	return nil, false
}

// uiElements stores references to important UI elements for updating
type uiElements struct {
	statusBar          *widget.Label
	codeOutput         *widget.Entry
	astOutput          *widget.Entry
	semanticOutput     *widget.Entry
	modelSelector      *widget.Select
	intentInput        *widget.Entry
	fileExplorer       *FileExplorer
	fileContentDisplay *widget.Entry
	filePathLabel      *widget.Label
}

// codeTheme is a custom theme for the app
type codeTheme struct {
	fyne.Theme
}

func newCodeTheme() fyne.Theme {
	return &codeTheme{Theme: theme.DarkTheme()}
}

func (t *codeTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 220, G: 220, B: 220, A: 255}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 25, G: 130, B: 196, A: 255}
	default:
		return t.Theme.Color(name, variant)
	}
}

// showModelInfo displays information about the currently selected model
func showModelInfo(w fyne.Window, state *AppState) {
	// Find the selected model in cache
	var selectedModel *OpenRouterModel
	for _, model := range modelsCache.Models {
		if model.ID == state.selectedModel {
			selectedModel = &model
			break
		}
	}
	
	if selectedModel == nil {
		dialog.ShowInformation("Model Information", 
			fmt.Sprintf("Selected model: %s\n\nAdditional information not available.", state.selectedModel),
			w)
		return
	}
	
	// Display model information
	dialog.ShowInformation("Model Information", 
		fmt.Sprintf("Model: %s\nID: %s\nContext Length: %d tokens\nCreated: %s",
			selectedModel.Name,
			selectedModel.ID,
			selectedModel.ContextLength,
			time.Unix(selectedModel.Created, 0).Format("January 2, 2006")),
		w)
}

// refreshModelsList forces a refresh of the models list
func refreshModelsList(w fyne.Window, state *AppState) {
	// Clear cache timestamp to force refresh
	modelsCache.Timestamp = time.Time{}
	
	// Show progress dialog
	progress := dialog.NewProgress("Refreshing Models", "Retrieving available models from OpenRouter...", w)
	progress.Show()
	
	// Start asynchronous operation
	go func() {
		modelIDs, err := fetchAvailableModels()
		
		// Close progress dialog
		progress.Hide()
		
		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to refresh models: %v", err), w)
			return
		}
		
		// Update the selector
		if state.ui != nil && state.ui.modelSelector != nil && len(modelIDs) > 0 {
			state.ui.modelSelector.Options = modelIDs
			state.ui.modelSelector.Refresh()
			
			dialog.ShowInformation("Models Refreshed", 
				fmt.Sprintf("Successfully loaded %d models from OpenRouter API", len(modelIDs)), 
				w)
		}
	}()
}

