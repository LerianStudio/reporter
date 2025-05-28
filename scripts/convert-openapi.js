#!/usr/bin/env node

/**
 * Script to convert OpenAPI specification to Postman Collection
 * Usage: node convert-openapi.js <input-file> <output-file>
 */

const fs = require('fs');
const path = require('path');
const yaml = require('js-yaml'); // We'll install this dependency

// Check arguments
if (process.argv.length < 4) {
  console.error('Usage: node convert-openapi.js <input-file> <output-file>');
  process.exit(1);
}

const inputFile = process.argv[2];
const outputFile = process.argv[3];

// Check if input file exists
if (!fs.existsSync(inputFile)) {
  console.error(`Input file not found: ${inputFile}`);
  process.exit(1);
}

// Create a temporary directory for the conversion
const tempDir = path.dirname(outputFile);
if (!fs.existsSync(tempDir)) {
  fs.mkdirSync(tempDir, { recursive: true });
}

// Read and parse the input file based on extension
let openApiSpec;
const fileContent = fs.readFileSync(inputFile, 'utf8');
const fileExt = path.extname(inputFile).toLowerCase();

if (fileExt === '.yaml' || fileExt === '.yml') {
  // Parse YAML file
  try {
    openApiSpec = yaml.load(fileContent);
  } catch (error) {
    console.error(`Error parsing YAML file: ${error.message}`);
    process.exit(1);
  }
} else {
  // Parse JSON file
  try {
    openApiSpec = JSON.parse(fileContent);
  } catch (error) {
    console.error(`Error parsing JSON file: ${error.message}`);
    process.exit(1);
  }
}

// Fix path parameter formats (replace :param with {param})
function fixPathParameters(spec) {
  const paths = spec.paths || {};
  const fixedPaths = {};
  
  for (const path in paths) {
    // Replace :param with {param} in path
    const fixedPath = path
      .replace(/:name/g, '{name}')
      .replace(/:id/g, '{id}')

    
    fixedPaths[fixedPath] = paths[path];
  }
  
  spec.paths = fixedPaths;
  return spec;
}

// Fix the OpenAPI spec
const fixedSpec = fixPathParameters(openApiSpec);

// Create a simple Postman collection structure
function createPostmanCollection(spec) {
  const info = spec.info || {};
  const paths = spec.paths || {};
  
  // Create collection structure
  const collection = {
    info: {
      name: info.title || 'API Collection',
      description: info.description || '',
      schema: 'https://schema.getpostman.com/json/collection/v2.1.0/collection.json'
    },
    item: []
  };
  
  // Group by tags
  const tagGroups = {};
  
  // Process each path
  for (const path in paths) {
    const pathItem = paths[path];
    
    // Process each method (GET, POST, etc.)
    for (const method in pathItem) {
      if (['get', 'post', 'put', 'delete', 'patch', 'options', 'head'].includes(method)) {
        const operation = pathItem[method];
        const tags = operation.tags || ['default'];
        const tag = tags[0]; // Use the first tag for grouping
        
        if (!tagGroups[tag]) {
          tagGroups[tag] = [];
        }
        
        // Create request item
        const requestItem = {
          name: operation.summary || `${method.toUpperCase()} ${path}`,
          request: {
            method: method.toUpperCase(),
            header: [],
            url: {
              raw: `{{baseUrl}}${path}`,
              host: ['{{baseUrl}}'],
              path: path.split('/').filter(p => p)
            },
            description: operation.description || ''
          },
          response: []
        };
        
        // Add parameters
        if (operation.parameters) {
          // Path parameters
          const pathParams = operation.parameters.filter(p => p.in === 'path');
          if (pathParams.length > 0) {
            requestItem.request.url.variable = pathParams.map(p => ({
              key: p.name,
              value: '',
              description: p.description || ''
            }));
          }
          
          // Query parameters
          const queryParams = operation.parameters.filter(p => p.in === 'query');
          if (queryParams.length > 0) {
            requestItem.request.url.query = queryParams.map(p => ({
              key: p.name,
              value: '',
              description: p.description || '',
              disabled: !p.required
            }));
          }
          
          // Header parameters
          const headerParams = operation.parameters.filter(p => p.in === 'header');
          if (headerParams.length > 0) {
            requestItem.request.header = headerParams.map(p => ({
              key: p.name,
              value: '',
              description: p.description || '',
              disabled: !p.required
            }));
          }
        }
        
        // Add request body if present (OpenAPI 3.0 style)
        if (operation.requestBody) {
          const content = operation.requestBody.content || {};
          const jsonContent = content['application/json'];
          
          if (jsonContent) {
            let exampleData = {};
            
            // Try to get example from schema
            if (jsonContent.schema) {
              // If it's a reference, try to find the schema
              if (jsonContent.schema.$ref && jsonContent.schema.$ref.startsWith('#/components/schemas/')) {
                const schemaName = jsonContent.schema.$ref.replace('#/components/schemas/', '');
                const schema = spec.components && spec.components.schemas ? spec.components.schemas[schemaName] : null;
                
                if (schema) {
                  // If the schema has an example, use it
                  if (schema.example) {
                    exampleData = schema.example;
                  }
                  // Otherwise, build an example from the properties
                  else if (schema.properties) {
                    exampleData = Object.keys(schema.properties).reduce((acc, key) => {
                      const prop = schema.properties[key];
                      if (prop.example !== undefined) {
                        acc[key] = prop.example;
                      } else if (prop.type === 'string') {
                        acc[key] = 'string';
                      } else if (prop.type === 'integer' || prop.type === 'number') {
                        acc[key] = 0;
                      } else if (prop.type === 'boolean') {
                        acc[key] = false;
                      } else if (prop.type === 'array') {
                        acc[key] = [];
                      } else if (prop.type === 'object') {
                        acc[key] = {};
                      }
                      return acc;
                    }, {});
                  }
                }
              }
              // If it's an inline schema
              else if (jsonContent.schema.properties) {
                exampleData = Object.keys(jsonContent.schema.properties).reduce((acc, key) => {
                  const prop = jsonContent.schema.properties[key];
                  if (prop.example !== undefined) {
                    acc[key] = prop.example;
                  } else if (prop.type === 'string') {
                    acc[key] = 'string';
                  } else if (prop.type === 'integer' || prop.type === 'number') {
                    acc[key] = 0;
                  } else if (prop.type === 'boolean') {
                    acc[key] = false;
                  } else if (prop.type === 'array') {
                    acc[key] = [];
                  } else if (prop.type === 'object') {
                    acc[key] = {};
                  }
                  return acc;
                }, {});
              }
            }
            
            // Use the example from the content if available
            if (Object.keys(exampleData).length === 0 && jsonContent.example) {
              exampleData = jsonContent.example;
            }
            
            requestItem.request.body = {
              mode: 'raw',
              raw: JSON.stringify(exampleData, null, 2),
              options: {
                raw: {
                  language: 'json'
                }
              }
            };
          }
        }
        
        tagGroups[tag].push(requestItem);
      }
    }
  }
  
  // Create folders for each tag
  for (const tag in tagGroups) {
    collection.item.push({
      name: tag,
      item: tagGroups[tag]
    });
  }
  
  return collection;
}

// Convert the OpenAPI spec to a Postman collection
const postmanCollection = createPostmanCollection(fixedSpec);

// Write the Postman collection to the output file
fs.writeFileSync(outputFile, JSON.stringify(postmanCollection, null, 2));
console.log(`Successfully converted ${inputFile} to ${outputFile}`);
process.exit(0);
