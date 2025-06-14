package generator

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"log"
	"path/filepath"
	"text/template"

	"github.com/stoewer/go-strcase"

	"github.com/om3kk/slacky/model"
)

// GenerateStruct processes a SlackModal, validates it, and generates Go code.
// It now returns the generated code, the struct name, and an error.
func GenerateStruct(modal model.SlackModal, tmplPath, packageName string) ([]byte, string, error) {
	// 1. Validate the modal's callback_id
	if modal.CallbackID == "" {
		return nil, "", errors.New("validation failed: modal 'callback_id' is missing or empty")
	}

	// 2. Prepare the data for the template
	structName := strcase.UpperCamelCase(modal.CallbackID) + "Input"
	data := model.TemplateData{
		PackageName: packageName,
		StructName:  structName,
		Fields:      []model.FieldInfo{},
		CallbackID:  modal.CallbackID,
	}

	// 3. Iterate over blocks to find input elements and their action_ids
	for i, block := range modal.Blocks {
		if block.Type != "input" {
			// This is not an input block, so we skip it.
			continue
		}

		if block.Element == nil {
			log.Printf("WARNING: Block at index %d is of type 'input' but has no 'element' property. Skipping.", i)
			continue
		}

		// 4. Validate the element's action_id
		if block.Element.ActionID == "" {
			// This is a fatal validation error as per requirements.
			return nil, "", fmt.Errorf("validation failed: input block at index %d is missing 'action_id' in its element", i)
		}

		// Log a warning for unhandled element types, but proceed.
		switch block.Element.Type {
		case "plain_text_input", "static_select", "multi_static_select":
			// Known and handled types
		default:
			log.Printf("WARNING: Unhandled element type '%s' for action_id '%s'. Field will be generated as string.", block.Element.Type, block.Element.ActionID)
		}

		field := model.FieldInfo{
			Name:    strcase.UpperCamelCase(block.Element.ActionID),
			JSONTag: block.Element.ActionID,
		}
		data.Fields = append(data.Fields, field)
	}

	if len(data.Fields) == 0 {
		log.Printf("WARNING: No input elements found in modal with callback_id '%s'. An empty struct will be generated.", modal.CallbackID)
	}

	// 5. Read and parse the template file
	// Use a FuncMap to add custom functions to the template
	funcMap := template.FuncMap{
		"camel": strcase.UpperCamelCase,
		"snake": strcase.SnakeCase,
	}
	tmpl, err := template.New(filepath.Base(tmplPath)).Funcs(funcMap).ParseFiles(tmplPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse template file '%s': %w", tmplPath, err)
	}

	// 6. Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, "", fmt.Errorf("failed to execute template: %w", err)
	}

	// 7. Format the generated Go code
	formattedCode, err := format.Source(buf.Bytes())
	if err != nil {
		// As per requirements, report the error but return the unformatted code for debugging.
		log.Printf("ERROR: Failed to format generated code: %v. The unformatted code will be saved.", err)
		log.Println("--- UNFORMATTED CODE ---")
		log.Println(buf.String())
		log.Println("------------------------")
		return buf.Bytes(), structName, nil // Return unformatted code on format error
	}

	return formattedCode, structName, nil
}
