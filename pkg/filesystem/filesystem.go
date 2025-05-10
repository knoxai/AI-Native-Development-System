package filesystem

import (
	"errors"
	"os"
	"path/filepath"
)

// FileSystem provides access to the host file system with additional
// functionality specific to code manipulation
type FileSystem struct {
	// WorkingDirectory is the current workspace directory
	WorkingDirectory string
}

// New creates a new FileSystem instance
func New(workingDir string) (*FileSystem, error) {
	// If working directory is not provided, use current directory
	if workingDir == "" {
		var err error
		workingDir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	// Create the directory if it doesn't exist
	if _, err := os.Stat(workingDir); os.IsNotExist(err) {
		if err := os.MkdirAll(workingDir, 0755); err != nil {
			return nil, err
		}
	}

	return &FileSystem{
		WorkingDirectory: workingDir,
	}, nil
}

// SetWorkingDirectory changes the current workspace directory
func (fs *FileSystem) SetWorkingDirectory(dir string) error {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return errors.New("directory does not exist")
	}

	fs.WorkingDirectory = dir
	return nil
}

// ReadFile reads a file from the file system
func (fs *FileSystem) ReadFile(path string) ([]byte, error) {
	fullPath := fs.resolvePath(path)
	return os.ReadFile(fullPath)
}

// WriteFile writes content to a file
func (fs *FileSystem) WriteFile(path string, content []byte) error {
	fullPath := fs.resolvePath(path)
	
	// Create directories if they don't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	return os.WriteFile(fullPath, content, 0644)
}

// DeleteFile removes a file from the file system
func (fs *FileSystem) DeleteFile(path string) error {
	fullPath := fs.resolvePath(path)
	return os.Remove(fullPath)
}

// ListFiles returns a list of files in a directory
func (fs *FileSystem) ListFiles(dir string) ([]string, error) {
	fullPath := fs.resolvePath(dir)
	
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	
	var files []string
	for _, entry := range entries {
		files = append(files, entry.Name())
	}
	
	return files, nil
}

// FileExists checks if a file exists
func (fs *FileSystem) FileExists(path string) bool {
	fullPath := fs.resolvePath(path)
	_, err := os.Stat(fullPath)
	return !os.IsNotExist(err)
}

// IsDirectory checks if a path is a directory
func (fs *FileSystem) IsDirectory(path string) bool {
	fullPath := fs.resolvePath(path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// resolvePath resolves a relative path to an absolute path
func (fs *FileSystem) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(fs.WorkingDirectory, path)
}

// CreateWorkspace creates a new workspace with default structure
func (fs *FileSystem) CreateWorkspace(name string) error {
	workspacePath := filepath.Join(fs.WorkingDirectory, name)
	
	// Create main directory
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return err
	}
	
	// Create subdirectories
	dirs := []string{
		"src",
		"docs",
		"tests",
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(workspacePath, dir), 0755); err != nil {
			return err
		}
	}
	
	// Create a README.md file
	readmePath := filepath.Join(workspacePath, "README.md")
	readmeContent := []byte("# " + name + "\n\nThis workspace was created by the AI-Native Development System.\n")
	
	if err := os.WriteFile(readmePath, readmeContent, 0644); err != nil {
		return err
	}
	
	return nil
} 