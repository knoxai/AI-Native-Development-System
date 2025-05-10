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
        
        try {
            // Show loading state
            resultOutput.innerHTML = '<div class="loading">Processing...</div>';
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
            
            // Display the result
            resultOutput.innerHTML = `
                <h3>Intent Processed</h3>
                <div class="result-info">
                    <strong>Intent Type:</strong> ${data.intent?.type || 'N/A'}<br>
                    <strong>Target:</strong> ${data.intent?.target || 'N/A'}<br>
                    <strong>Result:</strong> ${data.result || 'Success'}
                </div>
            `;
            
            // If there's generated code, display it
            if (data.code) {
                codeOutput.textContent = data.code;
                // Automatically switch to code tab when code is generated
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
            
            // If there are semantic entities and relations, display them
            if (data.entities && data.entities.length > 0) {
                let entitiesHtml = '<h3>Semantic Entities</h3><div class="entity-list">';
                data.entities.forEach(entity => {
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
                
                if (data.relations && data.relations.length > 0) {
                    let relationsHtml = '<h3>Semantic Relations</h3><div class="relation-list">';
                    data.relations.forEach(relation => {
                        relationsHtml += `
                            <div class="relation">
                                <div class="relation-type">${relation.type}</div>
                                <div class="relation-entities">
                                    <div class="relation-from">${relation.from}</div>
                                    <div class="relation-arrow">→</div>
                                    <div class="relation-to">${relation.to}</div>
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