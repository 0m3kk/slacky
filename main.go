// main.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/0m3kk/slacky/generator"
	"github.com/0m3kk/slacky/model"
	"github.com/0m3kk/slacky/templates"
)

func main() {
	// 1. Define and parse command-line flags for the output directory.
	outputDir := flag.String("output", "generated", "The directory to save generated Go files.")
	outPkgName := flag.String("pkg", "generated", "The name of output package")
	flag.Parse()

	// The remaining non-flag arguments are treated as the input JSON file paths.
	jsonFiles := flag.Args()

	// 2. Check that at least one input file has been provided.
	if len(jsonFiles) == 0 {
		fmt.Println("Error: No input JSON files provided.")
		fmt.Println("Usage: slacky [-output <dir>] <file1.json> [<file2.json> ...]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// 3. Create the output directory if it doesn't exist.
	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		log.Fatalf("Failed to create output directory '%s': %v", *outputDir, err)
	}
	fmt.Printf("Output will be saved in './%s/'\n", *outputDir)

	// A slice to hold information about each successfully generated struct.
	var generatedStructs []model.StructInfo
	// Use a map to collect unique action IDs for simple handlers.
	simpleActionIDs := make(map[string]struct{})

	// 4. Process each file to generate individual struct files.
	for _, jsonFilePath := range jsonFiles {
		fmt.Printf("Processing file: %s\n", jsonFilePath)

		// Read the input JSON file
		jsonData, err := os.ReadFile(jsonFilePath)
		if err != nil {
			log.Printf("ERROR: Failed to read file '%s': %v. Skipping file.", jsonFilePath, err)
			continue
		}

		// Parse the JSON into our SlackModal struct
		var modal model.SlackModal
		if err := json.Unmarshal(jsonData, &modal); err != nil {
			log.Printf("ERROR: Failed to parse JSON from file '%s': %v. Skipping file.", jsonFilePath, err)
			continue
		}

		// Find simple action IDs within the blocks.
		for _, block := range modal.Blocks {
			if block.Type == "actions" {
				for _, el := range block.Elements {
					if el.ActionID != "" {
						simpleActionIDs[el.ActionID] = struct{}{}
					}
				}
			}
		}

		if modal.Type == "modal" {
			// Generate the Go struct source code from the parsed data
			generatedCode, structName, err := generator.GenerateStruct(modal, templates.StructTmpl, *outPkgName)
			if err != nil {
				log.Fatalf("Could not generate struct for %s: %v", jsonFilePath, err)
			}

			// Write the generated code to an output file in the specified directory
			baseName := strings.TrimSuffix(filepath.Base(jsonFilePath), filepath.Ext(jsonFilePath))
			outputFileName := fmt.Sprintf("%s.go", baseName)
			outputFilePath := filepath.Join(*outputDir, outputFileName)

			if err := os.WriteFile(outputFilePath, generatedCode, 0o644); err != nil {
				log.Fatalf("FATAL: Failed to write generated code to file '%s': %v", outputFilePath, err)
			}

			fmt.Printf("Successfully generated %s\n", outputFilePath)

			// Add info to our slice for the dispatcher
			generatedStructs = append(generatedStructs, model.StructInfo{
				CallbackID: modal.CallbackID,
				TypeName:   structName,
			})
		}
	}

	// Convert the map of unique action IDs to a slice for the template.
	simpleActionIDSlice := make([]string, 0, len(simpleActionIDs))
	for id := range simpleActionIDs {
		simpleActionIDSlice = append(simpleActionIDSlice, id)
	}

	// 5. Generate the dispatcher file in the specified directory.
	if err := generator.GenerateDispatcher(generatedStructs, simpleActionIDSlice, templates.DispatcherTmpl, *outputDir, *outPkgName); err != nil {
		log.Fatalf("FATAL: Failed to generate dispatcher: %v", err)
	}

	fmt.Println("Code generation complete.")
}
