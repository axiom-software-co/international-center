package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run bundle-openapi.go <input-spec.yaml> <output-spec.yaml>")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	// Read the main OpenAPI spec
	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	specContent := string(data)
	
	// Get directory of input file for resolving relative paths
	inputDir := filepath.Dir(inputFile)
	
	// Find and resolve external references
	refPattern := regexp.MustCompile(`\$ref:\s*['"](\./[^#]+\.yaml)#/([^'"]+)['"]`)
	
	specContent = refPattern.ReplaceAllStringFunc(specContent, func(match string) string {
		matches := refPattern.FindStringSubmatch(match)
		if len(matches) != 3 {
			return match
		}
		
		refFile := matches[1]
		refPath := matches[2]
		
		// Read the referenced file
		fullPath := filepath.Join(inputDir, refFile)
		refData, err := ioutil.ReadFile(fullPath)
		if err != nil {
			fmt.Printf("Warning: Could not read referenced file %s: %v\n", fullPath, err)
			return match
		}
		
		// Parse the referenced YAML
		var refYaml map[string]interface{}
		if err := yaml.Unmarshal(refData, &refYaml); err != nil {
			fmt.Printf("Warning: Could not parse referenced file %s: %v\n", fullPath, err)
			return match
		}
		
		// For simple parameter files, check if the reference path is a direct top-level key
		if val, exists := refYaml[refPath]; exists {
			content, err := yaml.Marshal(val)
			if err != nil {
				fmt.Printf("Warning: Could not marshal content: %v\n", err)
				return match
			}
			
			// Convert to inline YAML (indented properly)
			lines := strings.Split(strings.TrimSpace(string(content)), "\n")
			indentedLines := make([]string, len(lines))
			for i, line := range lines {
				if i == 0 {
					indentedLines[i] = line
				} else {
					indentedLines[i] = "      " + line
				}
			}
			return strings.Join(indentedLines, "\n")
		}
		
		// Navigate to the specific reference path for nested structures
		parts := strings.Split(refPath, "/")
		current := refYaml
		for _, part := range parts {
			if part == "" {
				continue
			}
			if val, ok := current[part]; ok {
				if subMap, ok := val.(map[string]interface{}); ok {
					current = subMap
				} else {
					// This is the actual content we want to inline
					content, err := yaml.Marshal(val)
					if err != nil {
						fmt.Printf("Warning: Could not marshal content: %v\n", err)
						return match
					}
					
					// Convert to inline YAML (indented properly)
					lines := strings.Split(strings.TrimSpace(string(content)), "\n")
					indentedLines := make([]string, len(lines))
					for i, line := range lines {
						if i == 0 {
							indentedLines[i] = line
						} else {
							indentedLines[i] = "      " + line
						}
					}
					return strings.Join(indentedLines, "\n")
				}
			} else {
				fmt.Printf("Warning: Could not find path %s in file %s\n", refPath, fullPath)
				return match
			}
		}
		
		// Marshal the found content
		content, err := yaml.Marshal(current)
		if err != nil {
			fmt.Printf("Warning: Could not marshal content: %v\n", err)
			return match
		}
		
		// Convert to inline YAML (indented properly)
		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		indentedLines := make([]string, len(lines))
		for i, line := range lines {
			if i == 0 {
				indentedLines[i] = line
			} else {
				indentedLines[i] = "      " + line
			}
		}
		return strings.Join(indentedLines, "\n")
	})

	// Change OpenAPI version from 3.1.0 to 3.0.3 for oapi-codegen compatibility
	specContent = strings.Replace(specContent, "openapi: 3.1.0", "openapi: 3.0.3", 1)

	// Write the bundled spec
	if err := ioutil.WriteFile(outputFile, []byte(specContent), 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully bundled OpenAPI spec to %s\n", outputFile)
}