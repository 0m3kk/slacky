package generator

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"log"

	"github.com/0m3kk/slacky/model"
	"github.com/0m3kk/slacky/templates"
	"github.com/stoewer/go-strcase"
)

// GenerateStruct processes a SlackModal, validates it, and generates Go code.
// It now returns the generated code, the struct name, and an error.
func GenerateStruct(modal model.SlackModal, tmplName templates.Name, pkgName string) ([]byte, string, error) {
	// 1. Validate the modal's callback_id
	if modal.CallbackID == "" {
		return nil, "", errors.New("validation failed: modal 'callback_id' is missing or empty")
	}

	// 2. Prepare the data for the template
	structName := strcase.UpperCamelCase(modal.CallbackID) + "Input"
	data := model.TemplateData{
		PackageName: pkgName,
		StructName:  structName,
		Fields:      []model.FieldInfo{},
		CallbackID:  modal.CallbackID,
	}

	// 3. Iterate over blocks to find input elements and their action_ids
	for i, block := range modal.Blocks {
		if block.Type != "input" {
			continue
		}

		if block.Element == nil {
			log.Printf("WARNING: Block at index %d is of type 'input' but has no 'element' property. Skipping.", i)
			continue
		}

		if block.Element.ActionID == "" {
			return nil, "", fmt.Errorf("validation failed: input block at index %d is missing 'action_id' in its element", i)
		}

		// Determine the Go type for the struct field based on the Slack element type.
		var goType string
		switch block.Element.Type {
		case "number_input":
			if block.Element.IsDecimalAllowed {
				goType = "float64"
			} else {
				goType = "int64"
			}
		case "datetimepicker":
			goType = "int64"
		case "checkboxes", "multi_static_select", "multi_external_select", "multi_users_select", "multi_conversations_select", "multi_channels_select":
			goType = "[]string"
		default:
			// All other supported types map to a string.
			// (e.g., plain_text_input, datepicker, timepicker, selects, radio_buttons)
			goType = "string"
		}

		field := model.FieldInfo{
			Name:    strcase.UpperCamelCase(block.Element.ActionID),
			GoType:  goType,
			JSONTag: block.Element.ActionID,
		}
		data.Fields = append(data.Fields, field)
	}

	if len(data.Fields) == 0 {
		log.Printf("WARNING: No input elements found in modal with callback_id '%s'. An empty struct will be generated.", modal.CallbackID)
	}

	// 5. Get template
	tmpl, ok := templates.GetTemplate(tmplName)
	if !ok {
		return nil, "", fmt.Errorf("cannot find struct template")
	}

	// 6. Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, "", fmt.Errorf("failed to execute template: %w", err)
	}

	// 7. Format the generated Go code
	formattedCode, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("ERROR: Failed to format generated code: %v. The unformatted code will be saved.", err)
		return buf.Bytes(), structName, nil // Return unformatted code on format error
	}

	return formattedCode, structName, nil
}
