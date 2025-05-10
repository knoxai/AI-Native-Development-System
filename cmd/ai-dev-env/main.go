package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	
	"github.com/example/ai-dev-env/pkg/ast"
	"github.com/example/ai-dev-env/pkg/intent"
	"github.com/example/ai-dev-env/pkg/semantics"
	"github.com/example/ai-dev-env/pkg/server"
)

func main() {
	// Configure logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	fmt.Println("Starting AI-oriented Software Development Environment...")
	
	// Change to the directory where the binary is located to properly load web assets
	// This ensures the ./web directory can be found
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("Warning: Could not determine executable path: %v", err)
	} else {
		exeDir := filepath.Dir(exePath)
		err = os.Chdir(exeDir)
		if err != nil {
			log.Printf("Warning: Could not change to executable directory: %v", err)
		}
	}
	
	// Check if web directory exists
	if _, err := os.Stat("web"); os.IsNotExist(err) {
		// If not found in the current directory, try one level up
		// This is useful for development
		if _, err := os.Stat("../web"); os.IsNotExist(err) {
			log.Println("Warning: Web directory not found. UI may not work correctly.")
		} else {
			err = os.Chdir("..")
			if err != nil {
				log.Printf("Warning: Could not change to parent directory: %v", err)
			}
		}
	}
	
	// Initialize the semantic model
	semanticModel := semantics.NewModel()
	
	// Initialize the AST processor
	astProcessor := ast.NewProcessor(semanticModel)
	
	// Initialize the intent processor
	intentProcessor := intent.NewProcessor(astProcessor, semanticModel)
	
	// Start the server
	srv := server.New(intentProcessor, astProcessor, semanticModel)
	
	port := ":8080"
	fmt.Printf("Server started on http://localhost%s\n", port)
	if err := srv.Start(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}