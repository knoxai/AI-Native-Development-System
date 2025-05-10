package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	
	"github.com/knoxai/AI-Native-Development-System/pkg/ast"
	"github.com/knoxai/AI-Native-Development-System/pkg/intent"
	"github.com/knoxai/AI-Native-Development-System/pkg/semantics"
)

// Server provides an HTTP API for the AI development environment
type Server struct {
	intentProcessor *intent.Processor
	astProcessor    *ast.Processor
	semanticModel   *semantics.Model
}

// New creates a new server
func New(intentProc *intent.Processor, astProc *ast.Processor, semModel *semantics.Model) *Server {
	return &Server{
		intentProcessor: intentProc,
		astProcessor:    astProc,
		semanticModel:   semModel,
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
	
	// Health check
	mux.HandleFunc("/health", s.handleHealth)
	
	// Static files for the web UI
	fs := http.FileServer(http.Dir("./web"))
	mux.Handle("/", fs)
	
	log.Printf("Server starting on %s", addr)
	return http.ListenAndServe(addr, mux)
}

// handleIntent processes intent-based requests
func (s *Server) handleIntent(w http.ResponseWriter, r *http.Request) {
	log.Println("Received intent request")
	
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		Intent string `json:"intent"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	log.Printf("Processing intent: %s", req.Intent)
	
	// Parse and execute the intent
	parsedIntent, err := s.intentProcessor.ParseIntent(req.Intent)
	if err != nil {
		log.Printf("Error parsing intent: %v", err)
		http.Error(w, "Failed to parse intent: "+err.Error(), http.StatusBadRequest)
		return
	}
	
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
}`, req.Intent)

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
		{
			"id":          "type-error-001",
			"type":        "Type",
			"name":        "error",
			"description": "Standard error interface",
		},
	}

	// Generate semantic relations
	mockRelations := []map[string]interface{}{
		{
			"type": "Contains",
			"from": "pkg-auth-001",
			"to":   "func-login-001",
			"metadata": map[string]interface{}{
				"relationship": "package-function",
			},
		},
		{
			"type": "Uses",
			"from": "func-login-001",
			"to":   "pkg-crypto-sha256-001",
			"metadata": map[string]interface{}{
				"purpose":    "password hashing",
				"importance": "high",
			},
		},
		{
			"type": "Returns",
			"from": "func-login-001",
			"to":   "type-error-001",
			"metadata": map[string]interface{}{
				"condition": "invalid credentials",
			},
		},
	}

	// Create a response with all required data
	mockResponse := map[string]interface{}{
		"success": true,
		"intent": map[string]interface{}{
			"type":   parsedIntent.Type,
			"target": parsedIntent.Target,
		},
		"result":    "Intent processed successfully",
		"code":      generatedCode,
		"ast":       mockASTNode,
		"entities":  mockEntities,
		"relations": mockRelations,
	}
	
	// Return the result
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(mockResponse); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	log.Println("Intent processed successfully")
}

// handleAST processes AST manipulation requests
func (s *Server) handleAST(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		Code      string                 `json:"code"`
		Operation string                 `json:"operation"`
		Params    map[string]interface{} `json:"params"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// For demo purposes, just return a simplified response
	mockASTNode := map[string]interface{}{
		"type":  "Program",
		"body": []map[string]interface{}{
			{
				"type": "FunctionDeclaration",
				"name": "Login",
				"params": []string{"username", "password"},
				"returnType": []string{"string", "error"},
			},
		},
	}
	
	// Return the result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"code":    req.Code,
		"ast":     mockASTNode,
	})
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
	
	// For demo purposes, return mock entities and relations
	mockEntities := []map[string]interface{}{
		{
			"id":   "func-001",
			"type": "Function",
			"name": "Login",
			"description": "Validates user credentials",
			"properties": map[string]interface{}{
				"parameters": []string{"username", "password"},
				"returnTypes": []string{"string", "error"},
			},
		},
	}
	
	mockRelations := []map[string]interface{}{
		{
			"type": "Uses",
			"from": "func-001",
			"to":   "pkg-crypto",
			"metadata": map[string]interface{}{
				"purpose": "password hashing",
			},
		},
	}
	
	// Return the results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"entities":  mockEntities,
		"relations": mockRelations,
	})
}

// handleHealth provides a simple health check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}