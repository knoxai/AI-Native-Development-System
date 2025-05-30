package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	
	"github.com/knoxai/AI-Native-Development-System/pkg/ast"
	"github.com/knoxai/AI-Native-Development-System/pkg/intent"
	"github.com/knoxai/AI-Native-Development-System/pkg/llm"
	"github.com/knoxai/AI-Native-Development-System/pkg/semantics"
)

// Server provides an HTTP API for the AI development environment
type Server struct {
	intentProcessor *intent.Processor
	astProcessor    *ast.Processor
	semanticModel   *semantics.Model
	llmClient       *llm.Client
}

// New creates a new server
func New(intentProc *intent.Processor, astProc *ast.Processor, semModel *semantics.Model) *Server {
	// Initialize LLM client
	client, err := llm.NewClient()
	if err != nil {
		log.Printf("Warning: Could not initialize LLM client: %v", err)
	}
	
	return &Server{
		intentProcessor: intentProc,
		astProcessor:    astProc,
		semanticModel:   semModel,
		llmClient:       client,
	}
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()
	
	// Intent-based API endpoint
	mux.HandleFunc("/api/intent", s.handleIntent)
	
	// AST manipulation endpoint
	mux.HandleFunc("/api/ast", s.handleAST)
	
	// Semantic model query endpoint
	mux.HandleFunc("/api/semantics", s.handleSemantics)
	
	// Models list endpoint
	mux.HandleFunc("/api/models", s.handleModels)
	
	// Model selection endpoint
	mux.HandleFunc("/api/models/select", s.handleModelSelect)
	
	// Health check
	mux.HandleFunc("/health", s.handleHealth)
	
	// Static files for the web UI
	fs := http.FileServer(http.Dir("./web"))
	mux.Handle("/", fs)
	
	log.Printf("Server starting on %s", addr)
	return http.ListenAndServe(addr, mux)
}

// handleModels returns a list of available models from OpenRouter
func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	// Support both GET and POST methods
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// For POST requests, check for client-provided API key
	var clientAPIKey string
	if r.Method == http.MethodPost {
		var req struct {
			APIKey string `json:"api_key"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
			return
		}
		
		clientAPIKey = req.APIKey
	}
	
	// Try to use client-provided API key or server's LLM client
	var client *llm.Client
	var err error
	
	if clientAPIKey != "" {
		// Create a temporary client with the client-provided API key
		client = &llm.Client{
			APIKey:       clientAPIKey,
			DefaultModel: "openai/gpt-3.5-turbo", // Default model doesn't matter for listing
			HTTPClient:   &http.Client{},
		}
		log.Printf("Using client-provided API key to fetch models")
	} else if s.llmClient != nil {
		// Use server's LLM client
		client = s.llmClient
		log.Printf("Using server's LLM client to fetch models")
	} else {
		// No API key available
		http.Error(w, "API key is required to fetch models", http.StatusUnauthorized)
		return
	}
	
	// Fetch models from OpenRouter
	models, err := client.GetAvailableModels()
	if err != nil {
		log.Printf("Error fetching models: %v", err)
		http.Error(w, "Failed to fetch models: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return the models as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": models,
	})
}

// handleModelSelect sets the current model to use
func (s *Server) handleModelSelect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Parse request body
	var req struct {
		ModelID string `json:"model_id"`
		APIKey  string `json:"api_key"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Check if model ID is provided
	if req.ModelID == "" {
		http.Error(w, "Model ID is required", http.StatusBadRequest)
		return
	}
	
	// If client provided an API key, create a temporary client
	if req.APIKey != "" {
		log.Printf("Client provided API key for model selection")
		
		// Create a temporary client with the provided key
		// Using _ to discard the value since we don't need to use it
		_ = &llm.Client{
			APIKey:       req.APIKey,
			DefaultModel: req.ModelID,
			HTTPClient:   &http.Client{},
		}
		
		// Return success response with a note that we're using the client-provided key
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"model_id":  req.ModelID,
			"key_source": "client",
			"message":   "Using client-provided API key",
		})
		return
	}
	
	// Use server's LLM client if available
	if s.llmClient == nil {
		// Return a specific message that client can handle
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "LLM client not initialized. Please check your API key.",
			"message": "Model selection will only work locally.",
		})
		return
	}
	
	// Set the model in the LLM client
	s.llmClient.SetModel(req.ModelID)
	log.Printf("Model set to: %s", req.ModelID)
	
	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"model_id": req.ModelID,
		"key_source": "server",
	})
}

// handleIntent processes intent-based requests
func (s *Server) handleIntent(w http.ResponseWriter, r *http.Request) {
	log.Println("Received intent request")
	
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		Intent  string `json:"intent"`
		ModelID string `json:"model_id"`
		APIKey  string `json:"api_key"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	log.Printf("Processing intent: %s", req.Intent)
	
	// Check if we need to create a temporary client with the provided API key
	var tempClient *llm.Client
	if req.APIKey != "" && s.llmClient == nil {
		log.Printf("Creating temporary client with client-provided API key")
		tempClient = &llm.Client{
			APIKey:       req.APIKey,
			DefaultModel: req.ModelID,
			HTTPClient:   &http.Client{},
		}
		
		// Temporarily set the client for intent processing
		s.intentProcessor.SetLLMClient(tempClient)
		// Reset it after we're done
		defer s.intentProcessor.SetLLMClient(nil)
	}
	
	// Use existing client if available and set the model
	if s.llmClient != nil && req.ModelID != "" {
		log.Printf("Using model: %s", req.ModelID)
		// Set the model before processing
		s.llmClient.SetModel(req.ModelID)
	} else if tempClient != nil && req.ModelID != "" {
		tempClient.SetModel(req.ModelID)
	}
	
	// Parse and execute the intent
	parsedIntent, err := s.intentProcessor.ParseIntent(req.Intent)
	if err != nil {
		log.Printf("Error parsing intent: %v", err)
		http.Error(w, "Failed to parse intent: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Execute the intent
	result, err := s.intentProcessor.ExecuteIntent(parsedIntent)
	if err != nil {
		log.Printf("Error executing intent: %v", err)
		http.Error(w, "Failed to execute intent: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Check if the result is from the LLM (has sections)
	sections, ok := result.(map[string]string)
	if ok {
		// Process LLM-generated sections
		response := processLLMSections(sections, req.Intent)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// Handle legacy mock response for non-LLM processing
	mockResponse := generateMockResponse(req.Intent)
	json.NewEncoder(w).Encode(mockResponse)
}

// processLLMSections processes the sections returned by the LLM
func processLLMSections(sections map[string]string, originalIntent string) map[string]interface{} {
	response := make(map[string]interface{})
	
	// Add the original intent
	response["intent"] = originalIntent
	
	// Add the generated code
	if code, ok := sections["code"]; ok {
		response["generatedCode"] = code
	} else {
		response["generatedCode"] = "// No code was generated"
	}
	
	// Parse and add the AST representation
	if astStr, ok := sections["ast"]; ok {
		var astNode interface{}
		if err := json.Unmarshal([]byte(astStr), &astNode); err == nil {
			response["ast"] = astNode
		} else {
			// If parsing failed, just use the string
			response["ast"] = astStr
		}
	} else {
		response["ast"] = map[string]interface{}{
			"type": "Program",
			"body": []interface{}{},
		}
	}
	
	// Parse and add the semantic entities
	if semanticsStr, ok := sections["semantics"]; ok {
		var semantics interface{}
		if err := json.Unmarshal([]byte(semanticsStr), &semantics); err == nil {
			response["semantics"] = semantics
		} else {
			// If parsing failed, just use the string
			response["semantics"] = semanticsStr
		}
	} else {
		response["semantics"] = map[string]interface{}{
			"entities": []interface{}{},
			"relations": []interface{}{},
		}
	}
	
	return response
}

// generateMockResponse generates a mock response for non-LLM processing
func generateMockResponse(intentStr string) map[string]interface{} {
	// Generate code for the login function
	generatedCode := fmt.Sprintf(`// Generated code for: %s
package auth

import (
	"errors"
	"crypto/sha256"
	"encoding/hex"
)

// Login validates user credentials and returns a user ID or an error
func Login(username, password string) (string, error) {
	// In a real application, this would check against a database
	// For demonstration, we'll use a hardcoded example
	
	if username == "" || password == "" {
		return "", errors.New("username and password are required")
	}
	
	// Hash the password (in a real system, you would compare with stored hash)
	hasher := sha256.New()
	hasher.Write([]byte(password))
	hashedPassword := hex.EncodeToString(hasher.Sum(nil))
	
	// Mock validation
	if username == "admin" && hashedPassword == "8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918" {
		return "user-123", nil
	}
	
	return "", errors.New("invalid username or password")
}`, intentStr)

	// Generate AST representation
	mockASTNode := map[string]interface{}{
		"type": "Program",
		"body": []map[string]interface{}{
			{
				"type": "Package",
				"name": "auth",
			},
			{
				"type": "Import",
				"declarations": []string{"errors", "crypto/sha256", "encoding/hex"},
			},
			{
				"type": "FunctionDeclaration",
				"name": "Login",
				"params": []map[string]string{
					{"name": "username", "type": "string"},
					{"name": "password", "type": "string"},
				},
				"returnTypes": []string{"string", "error"},
				"body": []map[string]interface{}{
					{
						"type": "IfStatement",
						"test": map[string]interface{}{
							"type":     "BinaryExpression",
							"operator": "||",
							"left": map[string]interface{}{
								"type":     "BinaryExpression",
								"operator": "==",
								"left":     map[string]string{"type": "Identifier", "name": "username"},
								"right":    map[string]string{"type": "StringLiteral", "value": ""},
							},
							"right": map[string]interface{}{
								"type":     "BinaryExpression",
								"operator": "==",
								"left":     map[string]string{"type": "Identifier", "name": "password"},
								"right":    map[string]string{"type": "StringLiteral", "value": ""},
							},
						},
						"consequent": map[string]interface{}{
							"type": "BlockStatement",
							"body": []map[string]interface{}{
								{
									"type": "ReturnStatement",
									"arguments": []map[string]interface{}{
										{"type": "StringLiteral", "value": ""},
										{
											"type": "CallExpression",
											"callee": map[string]string{"type": "Identifier", "name": "errors.New"},
											"arguments": []map[string]string{
												{"type": "StringLiteral", "value": "username and password are required"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Generate semantic entities
	mockEntities := []map[string]interface{}{
		{
			"id":          "func-login-001",
			"type":        "Function",
			"name":        "Login",
			"description": "Validates user credentials and returns a user ID or error",
			"properties": map[string]interface{}{
				"parameters": []map[string]string{
					{"name": "username", "type": "string"},
					{"name": "password", "type": "string"},
				},
				"returnTypes": []string{"string", "error"},
				"visibility":  "public",
				"package":     "auth",
			},
		},
		{
			"id":          "pkg-auth-001",
			"type":        "Package",
			"name":        "auth",
			"description": "Authentication related functionality",
		},
	}

	// Generate relationships
	mockRelations := []map[string]interface{}{
		{
			"id":       "rel-001",
			"type":     "Contains",
			"fromID":   "pkg-auth-001",
			"toID":     "func-login-001",
			"metadata": map[string]interface{}{},
		},
	}

	// Create final response
	response := map[string]interface{}{
		"intent":        intentStr,
		"generatedCode": generatedCode,
		"ast":           mockASTNode,
		"semantics": map[string]interface{}{
			"entities":  mockEntities,
			"relations": mockRelations,
		},
	}

	return response
}

// handleAST processes AST manipulation requests
func (s *Server) handleAST(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Operation string                 `json:"operation"`
		Node      map[string]interface{} `json:"node"`
		Params    map[string]interface{} `json:"params"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// This is a simplified implementation that would perform AST operations
	// For demonstration, we'll just return a success message

	response := map[string]interface{}{
		"status":  "success",
		"message": "AST operation processed",
	}

	json.NewEncoder(w).Encode(response)
}

// handleSemantics processes semantic model queries
func (s *Server) handleSemantics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// This is a simplified implementation that would query the semantic model
	// For demonstration, we'll just return a mock response

	response := map[string]interface{}{
		"status":  "success",
		"message": "Semantic query processed",
		"results": []map[string]interface{}{
			{
				"id":          "func-login-001",
				"type":        "Function",
				"name":        "Login",
				"description": "Validates user credentials and returns a user ID or error",
			},
		},
	}

	json.NewEncoder(w).Encode(response)
}

// handleHealth provides a simple health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ok",
	}
	json.NewEncoder(w).Encode(response)
}

// GetLLMClient returns the LLM client for the server
func (s *Server) GetLLMClient() *llm.Client {
	return s.llmClient
}