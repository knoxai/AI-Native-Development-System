document.addEventListener('DOMContentLoaded', () => {
    // Tab switching
    const tabButtons = document.querySelectorAll('.tab-btn');
    const tabPanes = document.querySelectorAll('.tab-pane');
    
    tabButtons.forEach(button => {
        button.addEventListener('click', () => {
            // Remove active class from all buttons and panes
            tabButtons.forEach(btn => btn.classList.remove('active'));
            tabPanes.forEach(pane => pane.classList.remove('active'));
            
            // Add active class to clicked button and corresponding pane
            button.classList.add('active');
            const tabId = button.getAttribute('data-tab');
            document.getElementById(`${tabId}-tab`).classList.add('active');
        });
    });
    
    // Model selection
    const modelSelector = document.getElementById('model-selector');
    const modelInfo = document.getElementById('model-info');
    let selectedModel = null;
    let availableModels = [];
    
    // Fetch available models
    async function fetchModels() {
        try {
            const response = await fetch('/api/models');
            if (!response.ok) {
                throw new Error(`Server returned ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            availableModels = data.models || [];
            
            if (availableModels.length === 0) {
                modelSelector.innerHTML = '<option value="">No models available</option>';
                return;
            }
            
            // Sort models by name/provider
            availableModels.sort((a, b) => a.id.localeCompare(b.id));
            
            // Create option for each model
            modelSelector.innerHTML = '<option value="">Select a model</option>';
            availableModels.forEach(model => {
                const option = document.createElement('option');
                option.value = model.id;
                option.textContent = `${model.id} ${model.name ? `(${model.name})` : ''}`;
                modelSelector.appendChild(option);
            });
            
            // Select default model
            const defaultModelId = getDefaultModelId();
            if (defaultModelId && availableModels.some(m => m.id === defaultModelId)) {
                modelSelector.value = defaultModelId;
                displayModelInfo(availableModels.find(m => m.id === defaultModelId));
                selectedModel = defaultModelId;
            }
        } catch (error) {
            console.error('Error fetching models:', error);
            modelSelector.innerHTML = '<option value="">Error loading models</option>';
        }
    }
    
    // Get default model ID from local storage or use a reasonable default
    function getDefaultModelId() {
        const savedModel = localStorage.getItem('selectedModelId');
        if (savedModel) return savedModel;
        
        // Try to find a good default model like GPT-3.5-turbo
        const defaultOptions = [
            'openai/gpt-3.5-turbo',
            'openai/gpt-4',
            'anthropic/claude-3-haiku',
            'anthropic/claude-3-sonnet',
            'google/gemini-pro'
        ];
        
        for (const option of defaultOptions) {
            if (availableModels.some(m => m.id === option)) {
                return option;
            }
        }
        
        return availableModels[0]?.id || '';
    }
    
    // Display information about the selected model
    function displayModelInfo(model) {
        if (!model) {
            modelInfo.innerHTML = '<span class="warning">No model selected</span>';
            return;
        }
        
        // Calculate pricing in a readable format
        const promptPrice = parseFloat(model.pricing?.prompt || 0) * 1000;
        const completionPrice = parseFloat(model.pricing?.completion || 0) * 1000;
        
        modelInfo.innerHTML = `
            <div class="model-details">
                <div class="model-name">${model.name || model.id}</div>
                <div class="model-context">Context: ${formatTokens(model.context_length)}</div>
                <div class="model-pricing">Price: ${formatPrice(promptPrice)} / ${formatPrice(completionPrice)} per 1K tokens</div>
            </div>
        `;
    }
    
    // Format tokens with K, M suffix
    function formatTokens(tokens) {
        if (tokens >= 1000000) {
            return `${(tokens / 1000000).toFixed(1)}M`;
        } else if (tokens >= 1000) {
            return `${(tokens / 1000).toFixed(1)}K`;
        }
        return `${tokens}`;
    }
    
    // Format price with appropriate precision
    function formatPrice(price) {
        if (price < 0.01) {
            return `$${price.toFixed(5)}`;
        } else {
            return `$${price.toFixed(3)}`;
        }
    }
    
    // Handle model selection change
    modelSelector.addEventListener('change', async () => {
        const modelId = modelSelector.value;
        if (!modelId) {
            modelInfo.innerHTML = '<span class="warning">Please select a model</span>';
            selectedModel = null;
            return;
        }
        
        const model = availableModels.find(m => m.id === modelId);
        displayModelInfo(model);
        
        try {
            // Send selected model to the server
            const response = await fetch('/api/models/select', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ model_id: modelId })
            });
            
            if (!response.ok) {
                throw new Error(`Server returned ${response.status}: ${response.statusText}`);
            }
            
            // Save selection to local storage
            localStorage.setItem('selectedModelId', modelId);
            selectedModel = modelId;
            
            // Show success message
            const originalHtml = modelInfo.innerHTML;
            modelInfo.innerHTML += '<div class="success-message">Model successfully set!</div>';
            setTimeout(() => {
                const successMsg = modelInfo.querySelector('.success-message');
                if (successMsg) successMsg.remove();
            }, 3000);
        } catch (error) {
            console.error('Error setting model:', error);
            modelInfo.innerHTML = `<span class="error">Error setting model: ${error.message}</span>`;
        }
    });
    
    // Execute fetchModels on page load
    fetchModels();
    
    // Intent submission
    const intentTextarea = document.getElementById('intent-text');
    const submitButton = document.getElementById('submit-intent');
    const resultOutput = document.getElementById('result-output');
    const codeOutput = document.getElementById('code-output');
    const astOutput = document.getElementById('ast-output');
    const semanticOutput = document.getElementById('semantic-output');
    
    submitButton.addEventListener('click', async () => {
        const intent = intentTextarea.value.trim();
        if (!intent) {
            resultOutput.innerHTML = '<div class="error">Please enter an intent.</div>';
            return;
        }
        
        if (!selectedModel) {
            resultOutput.innerHTML = '<div class="error">Please select an AI model first.</div>';
            return;
        }
        
        try {
            // Show loading state
            resultOutput.innerHTML = '<div class="loading">Processing with AI...</div>';
            codeOutput.textContent = '';
            astOutput.innerHTML = '<div class="loading">Processing AST...</div>';
            semanticOutput.innerHTML = '<div class="loading">Processing semantic model...</div>';
            
            submitButton.disabled = true;
            
            // Send intent to backend
            const response = await fetch('/api/intent', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ intent })
            });
            
            const contentType = response.headers.get('Content-Type') || '';
            
            if (!response.ok) {
                let errorMessage = `Server Error (${response.status})`;
                
                if (contentType.includes('application/json')) {
                    const errorData = await response.json();
                    errorMessage = errorData.error || errorMessage;
                } else {
                    const errorText = await response.text();
                    errorMessage = errorText || errorMessage;
                }
                
                throw new Error(errorMessage);
            }
            
            // Parse response
            let data;
            if (contentType.includes('application/json')) {
                data = await response.json();
            } else {
                const textResponse = await response.text();
                resultOutput.innerHTML = `<pre>${textResponse}</pre>`;
                return;
            }
            
            // Determine if this is an LLM-generated response (has generatedCode property)
            const isLLMResponse = !!data.generatedCode;
            
            // Display the result
            resultOutput.innerHTML = `
                <h3>Intent Processed</h3>
                <div class="result-info">
                    <div class="ai-badge ${isLLMResponse ? 'ai-badge-active' : 'ai-badge-mock'}">
                        ${isLLMResponse ? 'AI-Generated' : 'Mock Data'}
                    </div>
                    <div class="model-used">
                        <strong>Model:</strong> ${selectedModel || 'N/A'}
                    </div>
                    <strong>Intent:</strong> ${data.intent || 'N/A'}<br>
                    <strong>Result:</strong> ${data.result || 'Success'}
                </div>
            `;
            
            // If there's generated code, display it
            if (data.generatedCode) {
                codeOutput.textContent = data.generatedCode;
                // Automatically switch to code tab when code is generated
                document.querySelector('[data-tab="code"]').click();
            } else if (data.code) {
                codeOutput.textContent = data.code;
                document.querySelector('[data-tab="code"]').click();
            }
            
            // If there's AST info, display it
            if (data.ast) {
                astOutput.innerHTML = `
                    <h3>Abstract Syntax Tree</h3>
                    <div class="ast-tree">
                        <pre class="json-tree">${JSON.stringify(data.ast, null, 2)}</pre>
                    </div>
                    <div class="ast-visualization">
                        <!-- Here we would render a visual representation of the AST -->
                        <div class="ast-node root">
                            <div class="node-content">Program</div>
                            <div class="node-children">
                                ${renderASTNodes(data.ast.body)}
                            </div>
                        </div>
                    </div>
                `;
            } else {
                astOutput.innerHTML = '<div class="info">No AST information available.</div>';
            }
            
            // Handle semantic entities and relations
            const semantics = data.semantics || {};
            const entities = semantics.entities || data.entities || [];
            const relations = semantics.relations || data.relations || [];
            
            if (entities.length > 0) {
                let entitiesHtml = '<h3>Semantic Entities</h3><div class="entity-list">';
                entities.forEach(entity => {
                    entitiesHtml += `
                        <div class="entity">
                            <div class="entity-header">
                                <strong class="entity-name">${entity.name}</strong>
                                <span class="entity-type">${entity.type}</span>
                                <span class="entity-id">${entity.id}</span>
                            </div>
                            <div class="entity-description">${entity.description || ''}</div>
                            ${renderEntityProperties(entity.properties)}
                        </div>
                    `;
                });
                entitiesHtml += '</div>';
                
                let semanticHtml = entitiesHtml;
                
                if (relations && relations.length > 0) {
                    let relationsHtml = '<h3>Semantic Relations</h3><div class="relation-list">';
                    relations.forEach(relation => {
                        const fromId = relation.fromID || relation.from;
                        const toId = relation.toID || relation.to;
                        
                        relationsHtml += `
                            <div class="relation">
                                <div class="relation-type">${relation.type}</div>
                                <div class="relation-entities">
                                    <div class="relation-from">${fromId}</div>
                                    <div class="relation-arrow">→</div>
                                    <div class="relation-to">${toId}</div>
                                </div>
                                ${renderRelationMetadata(relation.metadata)}
                            </div>
                        `;
                    });
                    relationsHtml += '</div>';
                    
                    semanticHtml += relationsHtml;
                }
                
                semanticOutput.innerHTML = semanticHtml;
            } else {
                semanticOutput.innerHTML = '<div class="info">No semantic information available.</div>';
            }
            
        } catch (error) {
            resultOutput.innerHTML = `<div class="error">Error: ${error.message}</div>`;
            codeOutput.textContent = '';
            astOutput.innerHTML = '<div class="error">Could not process AST due to an error.</div>';
            semanticOutput.innerHTML = '<div class="error">Could not process semantic model due to an error.</div>';
        } finally {
            submitButton.disabled = false;
        }
    });
    
    // Helper function to render AST nodes
    function renderASTNodes(nodes) {
        if (!nodes || nodes.length === 0) return '';
        
        let html = '';
        nodes.forEach(node => {
            html += `
                <div class="ast-node">
                    <div class="node-content">${node.type} ${node.name ? `(${node.name})` : ''}</div>
                    ${node.params ? `
                        <div class="node-params">
                            <div class="params-label">Params:</div>
                            ${renderParams(node.params)}
                        </div>
                    ` : ''}
                    ${node.body ? `
                        <div class="node-children">
                            ${renderASTNodes(Array.isArray(node.body) ? node.body : [node.body])}
                        </div>
                    ` : ''}
                </div>
            `;
        });
        
        return html;
    }
    
    // Helper function to render parameters
    function renderParams(params) {
        if (!params) return '';
        
        if (Array.isArray(params)) {
            let html = '<div class="params-list">';
            params.forEach(param => {
                if (typeof param === 'string') {
                    html += `<div class="param">${param}</div>`;
                } else if (typeof param === 'object') {
                    html += `<div class="param">${param.name}: ${param.type || ''}</div>`;
                }
            });
            html += '</div>';
            return html;
        }
        
        return `<div class="param">${String(params)}</div>`;
    }
    
    // Helper function to render entity properties
    function renderEntityProperties(properties) {
        if (!properties) return '';
        
        let html = '<div class="entity-properties">';
        for (const [key, value] of Object.entries(properties)) {
            html += `
                <div class="property">
                    <div class="property-key">${key}</div>
                    <div class="property-value">${renderPropertyValue(value)}</div>
                </div>
            `;
        }
        html += '</div>';
        
        return html;
    }
    
    // Helper function to render property values
    function renderPropertyValue(value) {
        if (value === null || value === undefined) return '';
        
        if (Array.isArray(value)) {
            let html = '<ul class="property-array">';
            value.forEach(item => {
                html += `<li>${renderPropertyValue(item)}</li>`;
            });
            html += '</ul>';
            return html;
        } else if (typeof value === 'object') {
            let html = '<div class="property-object">';
            for (const [k, v] of Object.entries(value)) {
                html += `<div><strong>${k}:</strong> ${renderPropertyValue(v)}</div>`;
            }
            html += '</div>';
            return html;
        }
        
        return String(value);
    }
    
    // Helper function to render relation metadata
    function renderRelationMetadata(metadata) {
        if (!metadata) return '';
        
        let html = '<div class="relation-metadata">';
        for (const [key, value] of Object.entries(metadata)) {
            html += `
                <div class="metadata-item">
                    <span class="metadata-key">${key}:</span>
                    <span class="metadata-value">${renderPropertyValue(value)}</span>
                </div>
            `;
        }
        html += '</div>';
        
        return html;
    }
    
    // Mock project tree for demonstration
    const projectTree = document.getElementById('project-tree');
    
    function createTreeItem(name, type, children) {
        const item = document.createElement('div');
        item.className = 'tree-item';
        item.innerHTML = `
            <span class="icon">${type === 'folder' ? '📁' : '📄'}</span>
            <span class="name">${name}</span>
        `;
        
        if (children && children.length) {
            const childContainer = document.createElement('div');
            childContainer.className = 'tree-children';
            childContainer.style.paddingLeft = '20px';
            
            children.forEach(child => {
                childContainer.appendChild(createTreeItem(child.name, child.type, child.children));
            });
            
            item.appendChild(childContainer);
        }
        
        return item;
    }
    
    // Mock project structure
    const projectData = [
        { 
            name: 'src', 
            type: 'folder',
            children: [
                { name: 'main.go', type: 'file' },
                { 
                    name: 'auth', 
                    type: 'folder',
                    children: [
                        { name: 'login.go', type: 'file' }
                    ]
                },
                { 
                    name: 'models', 
                    type: 'folder',
                    children: [
                        { name: 'user.go', type: 'file' },
                        { name: 'product.go', type: 'file' }
                    ]
                },
                { 
                    name: 'controllers', 
                    type: 'folder',
                    children: [
                        { name: 'user_controller.go', type: 'file' }
                    ]
                }
            ]
        },
        { name: 'go.mod', type: 'file' },
        { name: 'README.md', type: 'file' }
    ];
    
    // Render the project tree
    projectData.forEach(item => {
        projectTree.appendChild(createTreeItem(item.name, item.type, item.children));
    });
    
    // Add keyboard shortcut for submission (Ctrl+Enter)
    intentTextarea.addEventListener('keydown', (e) => {
        if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
            submitButton.click();
        }
    });
});