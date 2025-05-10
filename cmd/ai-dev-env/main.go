package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	
	"github.com/knoxai/AI-Native-Development-System/pkg/ast"
	"github.com/knoxai/AI-Native-Development-System/pkg/intent"
	"github.com/knoxai/AI-Native-Development-System/pkg/semantics"
	"github.com/knoxai/AI-Native-Development-System/pkg/server"
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
	
	// Connect the server's LLM client to the intent processor
	if llmClient := srv.GetLLMClient(); llmClient != nil {
		intentProcessor.SetLLMClient(llmClient)
		fmt.Println("OpenRouter API key found - AI code generation is enabled")
	} else {
		fmt.Println("Note: OpenRouter API key not found - you can browse models but AI code generation requires an API key")
		fmt.Println("Set the OPENROUTER_API_KEY environment variable to enable AI code generation")
	}
	
	port := ":8080"
	fmt.Printf("Server started on http://localhost%s\n", port)
	if err := srv.Start(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}