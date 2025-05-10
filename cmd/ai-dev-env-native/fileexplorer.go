package main

import (
	"path/filepath"
	"strings"
	
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// FileExplorer is a component that displays a file tree
type FileExplorer struct {
	appState     *AppState
	container    *fyne.Container
	treeWidget   *widget.Tree
	currentFiles map[string][]string
	baseDir      string
}

// NewFileExplorer creates a new file explorer component
func NewFileExplorer(state *AppState) *FileExplorer {
	explorer := &FileExplorer{
		appState:     state,
		currentFiles: make(map[string][]string),
	}
	
	// Create the tree widget
	explorer.treeWidget = widget.NewTree(
		explorer.childUIDs,
		explorer.isBranch,
		explorer.createNode,
		func(uid string, _ bool, _ fyne.CanvasObject) {
			explorer.onSelected(uid)
		},
	)
	
	// Set the tree root to the working directory
	explorer.baseDir = state.fileSystem.WorkingDirectory
	
	// Create container
	explorer.container = container.NewBorder(
		widget.NewLabel("Project Files"),
		nil, nil, nil,
		container.NewScroll(explorer.treeWidget),
	)
	
	// Initial load of files
	explorer.Refresh()
	
	return explorer
}

// Container returns the container for the file explorer
func (e *FileExplorer) Container() fyne.CanvasObject {
	return e.container
}

// Refresh reloads the file list
func (e *FileExplorer) Refresh() {
	// Reset file cache
	e.currentFiles = make(map[string][]string)
	
	// Add the root
	e.currentFiles[""] = []string{e.baseDir}
	
	// Refresh the tree
	e.treeWidget.Refresh()
	
	// Open the root branch
	e.treeWidget.OpenBranch(e.baseDir)
}

// childUIDs returns the children of a node
func (e *FileExplorer) childUIDs(uid string) []string {
	// If we've already loaded this directory, return the cached children
	if children, ok := e.currentFiles[uid]; ok {
		return children
	}
	
	// If this is an empty ID, return the base directory
	if uid == "" {
		return []string{e.baseDir}
	}
	
	// Otherwise, load the directory contents
	path := uid
	if uid != e.baseDir {
		// For items in subdirectories
		path = filepath.Join(e.baseDir, uid)
	}
	
	// List the files in the directory
	entries, err := e.appState.fileSystem.ListFiles(path)
	if err != nil {
		// Return empty list on error
		e.currentFiles[uid] = []string{}
		return []string{}
	}
	
	// Build paths for the children
	var children []string
	for _, entry := range entries {
		fullPath := filepath.Join(path, entry)
		// Make the path relative to the base directory
		if strings.HasPrefix(fullPath, e.baseDir) {
			children = append(children, strings.TrimPrefix(fullPath, e.baseDir+"/"))
		} else {
			children = append(children, fullPath)
		}
	}
	
	// Cache the result
	e.currentFiles[uid] = children
	
	return children
}

// isBranch determines if a node should be displayed as a branch
func (e *FileExplorer) isBranch(uid string) bool {
	// The root and base dir are always branches
	if uid == "" || uid == e.baseDir {
		return true
	}
	
	// Check if this path is a directory
	path := filepath.Join(e.baseDir, uid)
	isDir := e.appState.fileSystem.IsDirectory(path)
	
	return isDir
}

// createNode creates a tree node for display
func (e *FileExplorer) createNode(branch bool) fyne.CanvasObject {
	return container.NewHBox(
		widget.NewIcon(theme.FileIcon()),
		widget.NewLabel(""),
	)
}

// onSelected handles when a tree node is selected
func (e *FileExplorer) onSelected(uid string) {
	if uid == "" || uid == e.baseDir {
		return
	}
	
	// Get the full path
	path := filepath.Join(e.baseDir, uid)
	
	// Check if it's a directory
	isDir := e.appState.fileSystem.IsDirectory(path)
	
	if isDir {
		// If it's a directory, toggle its expansion
		if e.treeWidget.IsBranchOpen(uid) {
			e.treeWidget.CloseBranch(uid)
		} else {
			e.treeWidget.OpenBranch(uid)
		}
	} else {
		// If it's a file, try to open it
		openFile(path, e.appState)
	}
}

// openFile opens a file and displays its contents
func openFile(path string, state *AppState) {
	// Read the file
	content, err := state.fileSystem.ReadFile(path)
	if err != nil {
		// Handle error
		return
	}
	
	// Display the file content
	if state.ui.fileContentDisplay != nil {
		state.ui.fileContentDisplay.SetText(string(content))
		
		// Update the file path label
		if state.ui.filePathLabel != nil {
			state.ui.filePathLabel.SetText(path)
		}
	}
} 