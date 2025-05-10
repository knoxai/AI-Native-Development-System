package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENROUTER_API_KEY environment variable is not set")
		fmt.Println("Please set it using: export OPENROUTER_API_KEY=your_api_key_here")
		os.Exit(1)
	}
	
	fmt.Println("Testing OpenRouter API connection...")
	fmt.Println("API Key found:", maskAPIKey(apiKey))
	
	// Create HTTP request to models endpoint
	req, err := http.NewRequest("GET", "https://openrouter.co/v1/models", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	
	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		os.Exit(1)
	}
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("API Error: %s\n", resp.Status)
		fmt.Printf("Response body: %s\n", string(body))
		os.Exit(1)
	}
	
	fmt.Println("Success! Connection to OpenRouter API is working.")
	fmt.Println("Response status:", resp.Status)
	fmt.Printf("First 200 characters of response body: %s...\n", string(body)[:min(200, len(string(body)))])
}

// Mask API key for display
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****" // Too short to show safely
	}
	
	return key[:4] + "****" + key[len(key)-4:]
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
} 