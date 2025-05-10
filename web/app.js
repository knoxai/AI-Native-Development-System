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
    
    // API Key management
    const apiKeyContainer = document.getElementById('api-key-container');
    const apiKeyInput = document.getElementById('api-key-input');
    const saveApiKeyBtn = document.getElementById('save-api-key');
    const clearApiKeyBtn = document.getElementById('clear-api-key');
    const apiKeyStatus = document.getElementById('api-key-status');
    let userApiKey = null;
    
    // Load API key from localStorage if exists
    function loadApiKey() {
        const savedKey = localStorage.getItem('openrouterApiKey');
        if (savedKey) {
            userApiKey = savedKey;
            // Mask the key for display
            apiKeyInput.value = maskApiKey(savedKey);
            apiKeyStatus.innerHTML = '<span class="success">API key is set</span>';
            apiKeyStatus.classList.add('has-key');
        } else {
            apiKeyStatus.innerHTML = '<span class="warning">No API key set - Generation features disabled</span>';
            apiKeyStatus.classList.remove('has-key');
        }
    }
    
    // Mask API key for display
    function maskApiKey(key) {
        if (!key || key.length <= 8) return '';
        return key.substring(0, 4) + '••••••••••••' + key.substring(key.length - 4);
    }
    
    // Save API key to localStorage
    saveApiKeyBtn.addEventListener('click', () => {
        const key = apiKeyInput.value.trim();
        if (!key) {
            apiKeyStatus.innerHTML = '<span class="error">Please enter an API key</span>';
            return;
        }
        
        // If key is masked, user hasn't changed it
        if (key.includes('••••')) {
            apiKeyStatus.innerHTML = '<span class="info">API key unchanged</span>';
            return;
        }
        
        // Save the key
        localStorage.setItem('openrouterApiKey', key);
        userApiKey = key;
        apiKeyInput.value = maskApiKey(key);
        apiKeyStatus.innerHTML = '<span class="success">API key saved successfully!</span>';
        apiKeyStatus.classList.add('has-key');
        
        // Attempt to fetch models with the new API key
        fetchModels();
        
        // Clear status message after delay
        setTimeout(() => {
            apiKeyStatus.innerHTML = '<span class="success">API key is set</span>';
        }, 3000);
    });
    
    // Clear API key
    clearApiKeyBtn.addEventListener('click', () => {
        localStorage.removeItem('openrouterApiKey');
        userApiKey = null;
        apiKeyInput.value = '';
        apiKeyStatus.innerHTML = '<span class="warning">API key cleared - Generation features disabled</span>';
        apiKeyStatus.classList.remove('has-key');
    });
    
    // Handle focus on API key input to clear masked version
    apiKeyInput.addEventListener('focus', function() {
        if (this.value.includes('••••')) {
            this.value = '';
        }
    });
    
    // Load the API key on page load
    loadApiKey();
    
    // Load models and show guidance based on API key status
    if (userApiKey) {
        // If we have an API key, fetch models directly
        fetchModels();
    } else {
        // If no API key, show a welcome message with instructions
        modelInfo.innerHTML = `
            <div class="welcome-message">
                <p>Welcome to the AI-Native Development Environment!</p>
                <p>To get started:</p>
                <ol>
                    <li>Enter your OpenRouter API key in the field above</li>
                    <li>Click "Save Key" to store it in your browser</li>
                    <li>Select a model from the dropdown (will load after saving your key)</li>
                    <li>Enter your development intent in the text area</li>
                    <li>Click "Execute" to generate code</li>
                </ol>
                <p><a href="https://openrouter.co" target="_blank">Get an API key from OpenRouter →</a></p>
            </div>
        `;
        
        // Add a class to the API key container to highlight it
        apiKeyContainer.classList.add('highlight-container');
        
        // Remove highlight after the user interacts with the API key input
        apiKeyInput.addEventListener('focus', function() {
            apiKeyContainer.classList.remove('highlight-container');
        }, { once: true });
    }
    
    // Model selection
    const modelSelector = document.getElementById('model-selector');
    const modelInfo = document.getElementById('model-info');
    let selectedModel = null;
    let availableModels = [];
    
    // Fetch available models
    async function fetchModels() {
        try {
            // Check if we have cached models that are still valid
            const cachedData = localStorage.getItem('modelCache');
            const cacheTimestamp = localStorage.getItem('modelCacheTimestamp');
            const now = Date.now();
            const CACHE_DURATION = 12 * 60 * 60 * 1000; // 12 hours in milliseconds
            
            // Use cached data if it exists and hasn't expired
            if (cachedData && cacheTimestamp && (now - parseInt(cacheTimestamp)) < CACHE_DURATION) {
                console.log('Using cached models data');
                availableModels = JSON.parse(cachedData);
                populateModelSelector();
                return;
            }
            
            console.log('Fetching fresh models data');
            modelInfo.innerHTML = '<span class="info">Fetching models...</span>';
            
            // API key is required for the models endpoint
            if (!userApiKey) {
                // Show specific message when no API key is available
                console.log('No API key available, cannot fetch models');
                modelInfo.innerHTML = '<span class="warning">API key required to fetch models. Please enter your OpenRouter API key above.</span>';
                return;
            }
            
            // First try fetching models from our server (it will use the client-provided API key)
            try {
                console.log('Attempting to fetch models through server proxy');
                
                const payload = { api_key: userApiKey };
                const serverResponse = await fetch('/api/models', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(payload)
                });
                
                if (serverResponse.ok) {
                    const data = await serverResponse.json();
                    if (data && data.data && Array.isArray(data.data)) {
                        console.log('Successfully fetched models through server proxy');
                        availableModels = data.data.map(model => ({
                            id: model.id,
                            name: model.name || model.id,
                            context_length: model.context_length || 0,
                            pricing: model.pricing || {}
                        }));
                        
                        // Cache the models data
                        localStorage.setItem('modelCache', JSON.stringify(availableModels));
                        localStorage.setItem('modelCacheTimestamp', now.toString());
                        
                        populateModelSelector();
                        return;
                    }
                }
                
                // If we get here, the server proxy failed but we'll try direct API call next
                console.log('Server proxy failed, falling back to direct API call');
            } catch (serverError) {
                console.error('Error with server proxy, falling back to direct API call:', serverError);
            }
            
            // Fallback: Fetch directly from OpenRouter API
            console.log('Fetching models directly from OpenRouter API');
            
            // Set up proper headers according to OpenRouter API requirements
            const headers = {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${userApiKey}`
            };
            
            // Fetch directly from OpenRouter API
            const response = await fetch('https://openrouter.co/v1/models', {
                method: 'GET',
                headers: headers
            });
            
            if (!response.ok) {
                throw new Error(`OpenRouter API returned ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            
            // Extract just the models from the response
            if (data && data.data && Array.isArray(data.data)) {
                availableModels = data.data.map(model => ({
                    id: model.id,
                    name: model.name || model.id,
                    context_length: model.context_length || 0,
                    pricing: model.pricing || {}
                }));
                
                // Cache the models data
                localStorage.setItem('modelCache', JSON.stringify(availableModels));
                localStorage.setItem('modelCacheTimestamp', now.toString());
                
                populateModelSelector();
            } else {
                throw new Error('Invalid response format from OpenRouter API');
            }
        } catch (error) {
            console.error('Error fetching models:', error);
            modelSelector.innerHTML = '<option value="">Error loading models</option>';
            modelInfo.innerHTML = `<span class="error">Error: ${error.message}</span>`;
            
            // Try using cached data as fallback if available, regardless of age
            const cachedData = localStorage.getItem('modelCache');
            if (cachedData) {
                console.log('Using cached models as fallback after error');
                availableModels = JSON.parse(cachedData);
                populateModelSelector();
            } else {
                // If no cache, use a fallback list of common models
                console.log('Using hardcoded fallback model list');
                availableModels = [
                    { id: 'openai/gpt-3.5-turbo', name: 'GPT-3.5 Turbo', context_length: 16000 },
                    { id: 'openai/gpt-4o', name: 'GPT-4o', context_length: 8000 },
                    { id: 'anthropic/claude-3-haiku', name: 'Claude 3 Haiku', context_length: 200000 },
                    { id: 'anthropic/claude-3-sonnet', name: 'Claude 3 Sonnet', context_length: 200000 },
                    { id: 'google/gemini-pro', name: 'Gemini Pro', context_length: 32000 }
                ];
                populateModelSelector();
            }
        }
    }
    
    // Populate the model selector dropdown with available models
    function populateModelSelector() {
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
            // Prepare request payload
            const payload = { model_id: modelId };
            
            // Add API key if available
            if (userApiKey) {
                payload.api_key = userApiKey;
            }
            
            // Try to send the selected model to the server
            const response = await fetch('/api/models/select', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(payload)
            });
            
            // Save selection to local storage regardless of server response
            localStorage.setItem('selectedModelId', modelId);
            selectedModel = modelId;
            
            // Show success message
            if (response.ok) {
                const data = await response.json();
                const keySource = data.key_source || 'unknown';
                const successMessage = keySource === 'client' 
                    ? 'Model set with your API key' 
                    : 'Model successfully set on server!';
                    
                modelInfo.innerHTML += `<div class="success-message">${successMessage}</div>`;
            } else {
                // If server responds with error, we still set the model locally
                modelInfo.innerHTML += '<div class="success-message">Model set locally</div>';
            }
            
            // Remove success message after a delay
            setTimeout(() => {
                const successMsg = modelInfo.querySelector('.success-message');
                if (successMsg) successMsg.remove();
            }, 3000);
        } catch (error) {
            console.error('Error setting model on server:', error);
            // Still set the model locally even if server request fails
            localStorage.setItem('selectedModelId', modelId);
            selectedModel = modelId;
            
            const currentHtml = modelInfo.innerHTML;
            modelInfo.innerHTML = currentHtml + '<div class="success-message">Model set locally</div>';
            setTimeout(() => {
                const successMsg = modelInfo.querySelector('.success-message');
                if (successMsg) successMsg.remove();
            }, 3000);
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
            
            // Prepare the request payload
            const payload = { 
                intent,
                model_id: selectedModel
            };
            
            // Add API key if user has provided one
            if (userApiKey) {
                payload.api_key = userApiKey;
            }
            
            // Send intent to backend
            const response = await fetch('/api/intent', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(payload)
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
                
                // Special handling for API key errors
                if (errorMessage.includes("API key") || errorMessage.includes("not initialized")) {
                    if (!userApiKey) {
                        errorMessage = "You need to provide an OpenRouter API key to use code generation features. Please enter your API key in the settings above.";
                    } else {
                        errorMessage = "The API key you provided appears to be invalid. Please check your API key and try again.";
                    }
                }
                
                throw new Error(errorMessage);
            }
            
            // Parse response
            let data;
            if (contentType.includes('application/json')) {
                data = await response.json();
                console.log('API Response:', data);
                
                // Log specific data formats to help with debugging
                if (data.generatedCode) {
                    console.log('Generated code format:', typeof data.generatedCode, data.generatedCode.substring(0, 100) + '...');
                }
                if (data.ast) {
                    console.log('AST format:', typeof data.ast, data.ast);
                }
                if (data.semantics) {
                    console.log('Semantics format:', typeof data.semantics, data.semantics);
                }
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
                // Clean up the generated code output - remove Markdown code blocks and unescape characters
                let cleanedCode = data.generatedCode;
                
                // Remove markdown code fences if present
                cleanedCode = cleanedCode.replace(/```[a-z]*\n/g, '').replace(/```$/g, '');
                
                // Unescape special characters (like \u003c for <)
                cleanedCode = cleanedCode
                    .replace(/\\u003c/g, '<')
                    .replace(/\\u003e/g, '>')
                    .replace(/\\u0026/g, '&')
                    .replace(/\\n/g, '\n')
                    .replace(/\\"/g, '"')
                    .replace(/\\t/g, '\t')
                    .replace(/\\r/g, '\r');
                
                codeOutput.textContent = cleanedCode;
                // Automatically switch to code tab when code is generated
                document.querySelector('[data-tab="code"]').click();
            } else if (data.code) {
                codeOutput.textContent = data.code;
                document.querySelector('[data-tab="code"]').click();
            }
            
            // If there's AST info, display it
            if (data.ast) {
                // Handle cases where ast might be a string instead of an object
                let astData = data.ast;
                if (typeof astData === 'string') {
                    try {
                        astData = JSON.parse(astData);
                    } catch (e) {
                        console.warn('Could not parse AST data:', e);
                    }
                }
                
                astOutput.innerHTML = `
                    <h3>Abstract Syntax Tree</h3>
                    <div class="ast-tree">
                        <pre class="json-tree">${JSON.stringify(astData, null, 2)}</pre>
                    </div>
                    <div class="ast-visualization">
                        <!-- Here we would render a visual representation of the AST -->
                        <div class="ast-node root">
                            <div class="node-content">Program</div>
                            <div class="node-children">
                                ${renderASTNodes(astData.body || [])}
                            </div>
                        </div>
                    </div>
                `;
            } else {
                astOutput.innerHTML = '<div class="info">No AST information available.</div>';
            }
            
            // Handle semantic entities and relations
            const semantics = data.semantics || {};
            
            // Handle cases where semantics might be a string instead of an object
            let semanticsData = semantics;
            if (typeof semanticsData === 'string') {
                try {
                    semanticsData = JSON.parse(semanticsData);
                } catch (e) {
                    console.warn('Could not parse semantics data:', e);
                    semanticsData = { entities: [], relations: [] };
                }
            }
            
            const entities = semanticsData.entities || data.entities || [];
            const relations = semanticsData.relations || data.relations || [];
            
            let semanticHtml = '<h3>Semantic Model</h3>';
            
            if (entities.length > 0) {
                semanticHtml += '<h4>Entities</h4><div class="entity-list">';
                entities.forEach(entity => {
                    semanticHtml += `
                        <div class="entity">
                            <div class="entity-header">
                                <strong class="entity-name">${entity.name || 'Unnamed'}</strong>
                                <span class="entity-type">${entity.type || 'Unknown'}</span>
                                <span class="entity-id">${entity.id || ''}</span>
                            </div>
                            <div class="entity-description">${entity.description || ''}</div>
                            ${renderEntityProperties(entity.properties || {})}
                        </div>
                    `;
                });
                semanticHtml += '</div>';
            } else {
                semanticHtml += '<div class="info">No entities available in the semantic model.</div>';
            }
            
            if (relations && relations.length > 0) {
                semanticHtml += '<h4>Relations</h4><div class="relation-list">';
                relations.forEach(relation => {
                    const fromId = relation.fromID || relation.from || '';
                    const toId = relation.toID || relation.to || '';
                    
                    semanticHtml += `
                        <div class="relation">
                            <div class="relation-type">${relation.type || 'Unknown relation'}</div>
                            <div class="relation-entities">
                                <div class="relation-from">${fromId}</div>
                                <div class="relation-arrow">→</div>
                                <div class="relation-to">${toId}</div>
                            </div>
                            ${renderRelationMetadata(relation.metadata || {})}
                        </div>
                    `;
                });
                semanticHtml += '</div>';
            } else if (entities.length > 0) {
                // Only show this message if we have entities but no relations
                semanticHtml += '<div class="info">No relations available in the semantic model.</div>';
            }
            
            semanticOutput.innerHTML = semanticHtml;
            
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