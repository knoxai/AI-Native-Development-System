package intent

import (
	"errors"
	"fmt"
	"log"
	"strings"
	
	"github.com/knoxai/AI-Native-Development-System/pkg/ast"
	"github.com/knoxai/AI-Native-Development-System/pkg/llm"
	"github.com/knoxai/AI-Native-Development-System/pkg/semantics"
)

// Intent represents a development intention expressed in natural language
type Intent struct {
	Raw         string
	Type        string // Create, Modify, Delete, Query, etc.
	Target      string // Function, Class, Module, etc.
	Constraints []string
	Parameters  map[string]interface{}
}

// Processor handles intent-based operations
type Processor struct {
	astProcessor  *ast.Processor
	semanticModel *semantics.Model
	llmClient     *llm.Client
}

// NewProcessor creates a new intent processor
func NewProcessor(astProcessor *ast.Processor, semanticModel *semantics.Model) *Processor {
	return &Processor{
		astProcessor:  astProcessor,
		semanticModel: semanticModel,
	}
}

// SetLLMClient sets the LLM client for the processor
func (p *Processor) SetLLMClient(client *llm.Client) {
	p.llmClient = client
}

// GetLLMClient returns the current LLM client
func (p *Processor) GetLLMClient() *llm.Client {
	return p.llmClient
}

// ParseIntent parses a natural language intent into structured form
func (p *Processor) ParseIntent(rawIntent string) (*Intent, error) {
	intent := &Intent{
		Raw:        rawIntent,
		Parameters: make(map[string]interface{}),
	}
	
	// If LLM client is available, use it to parse the intent
	if p.llmClient != nil {
		return p.parseIntentWithLLM(rawIntent)
	}
	
	// Fallback to basic parsing if LLM is not available
	// Very basic parsing for demonstration
	if strings.Contains(rawIntent, "create") || strings.Contains(rawIntent, "make") {
		intent.Type = "Create"
		if strings.Contains(rawIntent, "function") {
			intent.Target = "Function"
		} else if strings.Contains(rawIntent, "class") {
			intent.Target = "Class"
		}
	} else if strings.Contains(rawIntent, "modify") || strings.Contains(rawIntent, "change") {
		intent.Type = "Modify"
	} else if strings.Contains(rawIntent, "delete") || strings.Contains(rawIntent, "remove") {
		intent.Type = "Delete"
	} else if strings.Contains(rawIntent, "query") || strings.Contains(rawIntent, "find") {
		intent.Type = "Query"
	}
	
	return intent, nil
}

// parseIntentWithLLM uses the LLM API to parse intent
func (p *Processor) parseIntentWithLLM(rawIntent string) (*Intent, error) {
	// Prepare the prompt for the LLM
	prompt := fmt.Sprintf(`Parse the following development intent and return a JSON object with type, target, constraints, and parameters.
	
Intent: "%s"
	
Valid types are: Create, Modify, Delete, Query
Valid targets include: Function, Class, Module, Variable, Interface, etc.
	
Your response should be a valid JSON object like:
{
  "type": "Create",
  "target": "Function",
  "constraints": ["Must validate input", "Must return error on failure"],
  "parameters": {
    "name": "login",
    "returnType": "bool"
  }
}
	
JSON:`, rawIntent)
	
	// Get completion from OpenRouter
	response, err := p.llmClient.GetCompletion(prompt)
	if err != nil {
		log.Printf("Error calling LLM API: %v", err)
		// Fall back to basic parsing
		return p.ParseIntent(rawIntent)
	}
	
	// Check if we got a response
	if len(response.Choices) == 0 {
		return nil, errors.New("no response from LLM API")
	}
	
	// Parse the JSON response
	text := response.Choices[0].Text
	
	// Create the intent object
	intent := &Intent{
		Raw:        rawIntent,
		Parameters: make(map[string]interface{}),
	}
	
	// Extract type and target from the response
	// In a real implementation, we would parse the JSON properly
	if strings.Contains(text, `"type": "Create"`) {
		intent.Type = "Create"
	} else if strings.Contains(text, `"type": "Modify"`) {
		intent.Type = "Modify"
	} else if strings.Contains(text, `"type": "Delete"`) {
		intent.Type = "Delete"
	} else if strings.Contains(text, `"type": "Query"`) {
		intent.Type = "Query"
	}
	
	if strings.Contains(text, `"target": "Function"`) {
		intent.Target = "Function"
	} else if strings.Contains(text, `"target": "Class"`) {
		intent.Target = "Class"
	} else if strings.Contains(text, `"target": "Module"`) {
		intent.Target = "Module"
	}
	
	return intent, nil
}

// ExecuteIntent executes an intent and returns the result
func (p *Processor) ExecuteIntent(intent *Intent) (interface{}, error) {
	switch intent.Type {
	case "Create":
		return p.handleCreateIntent(intent)
	case "Modify":
		return p.handleModifyIntent(intent)
	case "Delete":
		return p.handleDeleteIntent(intent)
	case "Query":
		return p.handleQueryIntent(intent)
	default:
		return nil, errors.New("unknown intent type")
	}
}

// handleCreateIntent handles creation intents
func (p *Processor) handleCreateIntent(intent *Intent) (interface{}, error) {
	// If LLM client is available, use it to generate code
	if p.llmClient != nil {
		return p.generateCodeWithLLM(intent)
	}
	
	// Generate entities from the intent
	entities, err := p.semanticModel.GenerateEntitiesFromIntent(intent.Raw)
	if err != nil {
		return nil, err
	}
	
	// Simplified for demonstration
	return entities, nil
}

// generateCodeWithLLM uses the LLM API to generate code based on intent
func (p *Processor) generateCodeWithLLM(intent *Intent) (interface{}, error) {
	// Prepare the prompt for the LLM
	prompt := fmt.Sprintf(`Generate Go code based on the following intent:
	
Intent: "%s"
	
The code should be well-structured, follow best practices, and include comments.
	
Your response should include:
1. The generated code
2. An AST representation (simplified)
3. A semantic description of the entities and relationships
	
Format your response as:
===CODE===
(generated code here)
===AST===
(JSON representation of AST)
===SEMANTICS===
(JSON representation of semantic entities and relationships)`, intent.Raw)
	
	// Get completion from OpenRouter
	response, err := p.llmClient.GetCompletion(prompt)
	if err != nil {
		log.Printf("Error calling LLM API: %v", err)
		return nil, err
	}
	
	// Check if we got a response
	if len(response.Choices) == 0 {
		return nil, errors.New("no response from LLM API")
	}
	
	// Parse the response sections
	text := response.Choices[0].Text
	
	// Split the text into sections
	sections := make(map[string]string)
	
	// Extract code section
	if codeIdx := strings.Index(text, "===CODE==="); codeIdx != -1 {
		endIdx := strings.Index(text[codeIdx:], "===AST===")
		if endIdx != -1 {
			sections["code"] = text[codeIdx+len("===CODE==="):codeIdx+endIdx]
		}
	}
	
	// Extract AST section
	if astIdx := strings.Index(text, "===AST==="); astIdx != -1 {
		endIdx := strings.Index(text[astIdx:], "===SEMANTICS===")
		if endIdx != -1 {
			sections["ast"] = text[astIdx+len("===AST==="):astIdx+endIdx]
		}
	}
	
	// Extract semantics section
	if semIdx := strings.Index(text, "===SEMANTICS==="); semIdx != -1 {
		sections["semantics"] = text[semIdx+len("===SEMANTICS==="):]
	}
	
	// Return the parsed sections
	return sections, nil
}

// handleModifyIntent handles modification intents
func (p *Processor) handleModifyIntent(intent *Intent) (interface{}, error) {
	// Find the entities to modify
	entities, _ := p.semanticModel.QueryByIntent(intent.Raw)
	if len(entities) == 0 {
		return nil, errors.New("no entities found to modify")
	}
	
	// Simplified for demonstration
	return entities, nil
}

// handleDeleteIntent handles deletion intents
func (p *Processor) handleDeleteIntent(intent *Intent) (interface{}, error) {
	// Find the entities to delete
	entities, _ := p.semanticModel.QueryByIntent(intent.Raw)
	if len(entities) == 0 {
		return nil, errors.New("no entities found to delete")
	}
	
	// Simplified for demonstration
	return entities, nil
}

// handleQueryIntent handles query intents
func (p *Processor) handleQueryIntent(intent *Intent) (interface{}, error) {
	// Query the semantic model
	entities, relations := p.semanticModel.QueryByIntent(intent.Raw)
	
	// Format the results
	// In a real system, this would be much more sophisticated
	
	return map[string]interface{}{
		"entities":  entities,
		"relations": relations,
	}, nil
}