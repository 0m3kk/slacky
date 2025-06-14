// // main_test.go
package main

// import (
// 	"context"
// 	"encoding/json"
// 	"os"
// 	"path/filepath"
// 	"strings"
// 	"testing"

// 	"github.com/slack-go/slack"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// // --- Test Setup & Teardown ---

// // setupTestEnv creates temporary directories and files needed for a test run.
// func setupTestEnv(t *testing.T) (testDir, outputDir string, cleanup func()) {
// 	// Create a temporary directory for the entire test
// 	testDir, err := os.MkdirTemp("", "codegen_test_*")
// 	require.NoError(t, err)

// 	// Define the output directory path within the test directory
// 	outputDir = filepath.Join(testDir, "generated_structs")

// 	// Create dummy template files
// 	structTmplContent := `// Struct Template
// package {{.PackageName}}
// import "strings"
// // {{.StructName | replace "Input" "" | toLower}}
// type {{.StructName}} struct {
// {{- range .Fields}}
//     {{.Name}} string ` + "`json:\"{{.JSONTag}}\"`" + `
// {{- end}}
// }`
// 	dispatcherTmplContent := `// Dispatcher Template
// package {{.PackageName}}
// import "context"
// import "github.com/slack-go/slack"
// // {{range .Structs}}
// // Handler for {{.TypeName}}
// // {{end}}
// `
// 	err = os.WriteFile(filepath.Join(testDir, "template.tmpl"), []byte(structTmplContent), 0644)
// 	require.NoError(t, err)

// 	err = os.WriteFile(filepath.Join(testDir, "dispatcher.tmpl"), []byte(dispatcherTmplContent), 0644)
// 	require.NoError(t, err)

// 	// Set the working directory to the test directory
// 	originalWd, err := os.Getwd()
// 	require.NoError(t, err)
// 	err = os.Chdir(testDir)
// 	require.NoError(t, err)

// 	// The cleanup function restores the working directory and removes the temp dir
// 	cleanup = func() {
// 		err := os.Chdir(originalWd)
// 		if err != nil {
// 			// Log the error but don't fail the test at this point
// 			t.Logf("Warning: failed to change back to original directory: %v", err)
// 		}
// 		os.RemoveAll(testDir)
// 	}

// 	return testDir, outputDir, cleanup
// }

// // --- End-to-End Tests ---

// func TestE2E_SuccessPath(t *testing.T) {
// 	testDir, outputDir, cleanup := setupTestEnv(t)
// 	defer cleanup()

// 	// 1. Create a valid input JSON file
// 	jsonContent := `{
//         "type": "modal",
//         "callback_id": "my-cool-modal",
//         "title": {"type": "plain_text", "text": "Test Modal"},
//         "blocks": [
//             {
//                 "type": "input",
//                 "element": {"type": "plain_text_input", "action_id": "first_name"}
//             },
//             {
//                 "type": "input",
//                 "element": {"type": "static_select", "action_id": "user_role"}
//             },
//             {
//                 "type": "actions",
//                 "elements": [{"type": "button", "action_id": "a_button"}]
//             }
//         ]
//     }`
// 	jsonPath := filepath.Join(testDir, "modal.json")
// 	err := os.WriteFile(jsonPath, []byte(jsonContent), 0644)
// 	require.NoError(t, err)

// 	// 2. Mock os.Args and run the main logic via a testable function
// 	os.Args = []string{"codegen", jsonPath}
// 	runMainLogic(t)

// 	// 3. Assert that the struct file was generated correctly
// 	generatedStructPath := filepath.Join(outputDir, "modal.go")
// 	assert.FileExists(t, generatedStructPath, "Expected struct file to be created")

// 	generatedStructContent, err := os.ReadFile(generatedStructPath)
// 	require.NoError(t, err)
// 	expectedStructContent := `// Struct Template
// package generated
// import "strings"
// // mycoolmodal
// type MyCoolModalInput struct {
// 	FirstName string ` + "`json:\"first_name\"`" + `
// 	UserRole  string ` + "`json:\"user_role\"`" + `
// }
// `
// 	assert.Equal(t, normalizeNewlines(expectedStructContent), normalizeNewlines(string(generatedStructContent)))

// 	// 4. Assert that the dispatcher file was generated
// 	generatedDispatcherPath := filepath.Join(outputDir, "dispatcher.go")
// 	assert.FileExists(t, generatedDispatcherPath, "Expected dispatcher file to be created")

// 	generatedDispatcherContent, err := os.ReadFile(generatedDispatcherPath)
// 	require.NoError(t, err)
// 	assert.Contains(t, string(generatedDispatcherContent), "// Handler for MyCoolModalInput")
// }

// // runMainLogic encapsulates the main function's logic for testing, avoiding os.Exit
// func runMainLogic(t *testing.T) {
// 	// Re-implement main's logic but use t.Fatalf on error instead of log.Fatalf
// 	if len(os.Args) < 2 {
// 		t.Fatalf("Usage: %s <file1.json> ...", os.Args[0])
// 	}

// 	outputDir := "generated_structs"
// 	structTemplateFile := "template.tmpl"
// 	dispatcherTemplateFile := "dispatcher.tmpl"

// 	if err := os.MkdirAll(outputDir, 0755); err != nil {
// 		t.Fatalf("Failed to create output directory '%s': %v", outputDir, err)
// 	}

// 	var generatedStructs []StructInfo
// 	for _, jsonFilePath := range os.Args[1:] {
// 		jsonData, err := os.ReadFile(jsonFilePath)
// 		require.NoError(t, err)
// 		var modal SlackModal
// 		err = json.Unmarshal(jsonData, &modal)
// 		require.NoError(t, err)
// 		generatedCode, structName, err := GenerateStruct(modal, structTemplateFile)
// 		require.NoError(t, err)

// 		baseName := strings.TrimSuffix(filepath.Base(jsonFilePath), filepath.Ext(jsonFilePath))
// 		outputFilePath := filepath.Join(outputDir, baseName+".go")
// 		err = os.WriteFile(outputFilePath, generatedCode, 0644)
// 		require.NoError(t, err)

// 		generatedStructs = append(generatedStructs, StructInfo{
// 			CallbackID: modal.CallbackID,
// 			TypeName:   structName,
// 		})
// 	}

// 	err := GenerateDispatcher(generatedStructs, dispatcherTemplateFile, outputDir)
// 	require.NoError(t, err)
// }

// // --- Unit Tests ---

// func TestToCamelCase(t *testing.T) {
// 	testCases := []struct {
// 		name     string
// 		input    string
// 		expected string
// 	}{
// 		{"Snake Case", "hello_world", "HelloWorld"},
// 		{"Kebab Case", "my-first-modal", "MyFirstModal"},
// 		{"Already CamelCase", "AlreadyCamel", "AlreadyCamel"},
// 		{"Single Word", "submit", "Submit"},
// 		{"With Numbers", "action_123_go", "Action123Go"},
// 		{"Empty String", "", ""},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			assert.Equal(t, tc.expected, toCamelCase(tc.input))
// 		})
// 	}
// }

// func TestGenerateStruct_Logic(t *testing.T) {
// 	_, _, cleanup := setupTestEnv(t)
// 	defer cleanup()

// 	t.Run("Modal with no input blocks", func(t *testing.T) {
// 		modal := SlackModal{
// 			CallbackID: "no-inputs-modal",
// 			Blocks: []SlackBlock{
// 				{Type: "section", Label: &SlackText{Text: "Just some text"}},
// 			},
// 		}
// 		code, name, err := GenerateStruct(modal, "template.tmpl")
// 		require.NoError(t, err)
// 		assert.Equal(t, "NoInputsModalInput", name)
// 		assert.NotContains(t, string(code), "`json:") // Should be an empty struct
// 	})

// 	t.Run("Modal with unhandled input type", func(t *testing.T) {
// 		// This should still succeed, just with a warning (which we can't test directly here)
// 		modal := SlackModal{
// 			CallbackID: "unhandled-type-modal",
// 			Blocks: []SlackBlock{
// 				{Type: "input", Element: &SlackElement{Type: "some_new_fancy_input", ActionID: "fancy_action"}},
// 			},
// 		}
// 		code, name, err := GenerateStruct(modal, "template.tmpl")
// 		require.NoError(t, err)
// 		assert.Equal(t, "UnhandledTypeModalInput", name)
// 		assert.Contains(t, string(code), "FancyAction string `json:\"fancy_action\"`")
// 	})
// }

// func TestGenerateStruct_ValidationAndErrors(t *testing.T) {
// 	_, _, cleanup := setupTestEnv(t)
// 	defer cleanup()

// 	t.Run("Missing CallbackID", func(t *testing.T) {
// 		modal := SlackModal{CallbackID: "", Blocks: []SlackBlock{}}
// 		_, _, err := GenerateStruct(modal, "template.tmpl")
// 		require.Error(t, err)
// 		assert.Contains(t, err.Error(), "modal 'callback_id' is missing or empty")
// 	})

// 	t.Run("Missing ActionID in Input Block", func(t *testing.T) {
// 		modal := SlackModal{
// 			CallbackID: "valid-id",
// 			Blocks: []SlackBlock{
// 				{Type: "input", Element: &SlackElement{Type: "plain_text_input", ActionID: ""}},
// 			},
// 		}
// 		_, _, err := GenerateStruct(modal, "template.tmpl")
// 		require.Error(t, err)
// 		assert.Contains(t, err.Error(), "is missing 'action_id' in its element")
// 	})

// 	t.Run("Missing template file", func(t *testing.T) {
// 		modal := SlackModal{CallbackID: "valid-id"}
// 		_, _, err := GenerateStruct(modal, "nonexistent_template.tmpl")
// 		require.Error(t, err)
// 		assert.Contains(t, err.Error(), "failed to parse template file")
// 	})
// }

// func TestGenerateDispatcher_Logic(t *testing.T) {
// 	testDir, outputDir, cleanup := setupTestEnv(t)
// 	defer cleanup()

// 	t.Run("No structs provided", func(t *testing.T) {
// 		err := GenerateDispatcher([]StructInfo{}, "dispatcher.tmpl", outputDir)
// 		require.NoError(t, err, "Should not error when no structs are provided")
// 		// Check that the dispatcher file was NOT created
// 		assert.NoFileExists(t, filepath.Join(outputDir, "dispatcher.go"))
// 	})

// 	t.Run("Missing dispatcher template file", func(t *testing.T) {
// 		structs := []StructInfo{{CallbackID: "id", TypeName: "TypeName"}}
// 		err := GenerateDispatcher(structs, "nonexistent.tmpl", outputDir)
// 		require.Error(t, err)
// 		assert.Contains(t, err.Error(), "failed to parse dispatcher template file")
// 	})
// }

// func TestUnmarshalStateValues(t *testing.T) {
// 	stateValues := map[string]map[string]slack.BlockAction{
// 		"text_block": {
// 			"text_action": {Type: "plain_text_input", Value: "Hello World"},
// 		},
// 		"select_block": {
// 			"select_action": {Type: "static_select", SelectedOption: slack.Option{Value: "option_1"}},
// 		},
// 		"date_block": {
// 			"date_action": {Type: "datepicker", SelectedDate: "2025-06-15"},
// 		},
// 		"multi_select_block": {
// 			"multi_select_action": {
// 				Type: "multi_static_select",
// 				SelectedOptions: []slack.Option{
// 					{Value: "choice_a"},
// 					{Value: "choice_c"},
// 				},
// 			},
// 		},
// 		"checkbox_block": {
// 			"checkbox_action": {
// 				Type: "checkboxes",
// 				SelectedOptions: []slack.Option{
// 					{Value: "check_1"},
// 				},
// 			},
// 		},
// 		"radio_block": {
// 			"radio_action": {Type: "radio_buttons", SelectedOption: slack.Option{Value: "radio_yes"}},
// 		},
// 		"empty_text_block": {
// 			"empty_text_action": {Type: "plain_text_input", Value: ""},
// 		},
// 	}

// 	type TestInput struct {
// 		TextAction        string `json:"text_action"`
// 		SelectAction      string `json:"select_action"`
// 		DateAction        string `json:"date_action"`
// 		MultiSelectAction string `json:"multi_select_action"`
// 		CheckboxAction    string `json:"checkbox_action"`
// 		RadioAction       string `json:"radio_action"`
// 		EmptyTextAction   string `json:"empty_text_action"`
// 		UnrelatedField    string `json:"-"` // This should not be populated
// 	}

// 	var target TestInput
// 	err := unmarshalStateValues(stateValues, &target)
// 	require.NoError(t, err)

// 	expectedMultiSelectJSON := `["choice_a","choice_c"]`
// 	expectedCheckboxJSON := `["check_1"]`

// 	assert.Equal(t, "Hello World", target.TextAction)
// 	assert.Equal(t, "option_1", target.SelectAction)
// 	assert.Equal(t, "2025-06-15", target.DateAction)
// 	assert.JSONEq(t, expectedMultiSelectJSON, target.MultiSelectAction)
// 	assert.JSONEq(t, expectedCheckboxJSON, target.CheckboxAction)
// 	assert.Equal(t, "radio_yes", target.RadioAction)
// 	assert.Equal(t, "", target.EmptyTextAction)
// 	assert.Equal(t, "", target.UnrelatedField)
// }

// func TestDispatcher(t *testing.T) {
// 	// 1. Setup
// 	dispatcher := NewDispatcher()
// 	viewHandled := false
// 	blockActionHandled := false

// 	// 2. Create and register handlers
// 	type TestViewHandler struct{}
// 	func (h *TestViewHandler) Handle(ctx context.Context, i slack.InteractionCallback, input MyCoolModalInput) error {
// 		assert.Equal(t, "test-user", i.User.Name)
// 		assert.Equal(t, "value1", input.FirstName)
// 		viewHandled = true
// 		return nil
// 	}
// 	dispatcher.RegisterMyCoolModalInputHandler(&TestViewHandler{})

// 	type TestBlockHandler struct{}
// 	func (h *TestBlockHandler) Handle(ctx context.Context, i slack.InteractionCallback, a slack.BlockAction) error {
// 		assert.Equal(t, "button-action", a.ActionID)
// 		assert.Equal(t, "button_val", a.Value)
// 		blockActionHandled = true
// 		return nil
// 	}
// 	dispatcher.RegisterBlockActionHandler("button-action", &TestBlockHandler{})

// 	// 3. Test View Submission Dispatch
// 	viewInteraction := slack.InteractionCallback{
// 		Type: slack.InteractionTypeViewSubmission,
// 		User: slack.User{Name: "test-user"},
// 		View: slack.View{
// 			CallbackID: "my_cool_modal", // Note: The dispatcher uses the original callback_id
// 			State: &slack.State{
// 				Values: map[string]map[string]slack.BlockAction{
// 					"block1": {"first_name": {Value: "value1"}},
// 				},
// 			},
// 		},
// 	}

// 	// This is a placeholder for the generated struct to make the test compile
// 	// In a real scenario, this would come from the generated file.
// 	type MyCoolModalInput struct {
// 		FirstName string `json:"first_name"`
// 	}
// 	err := dispatcher.Dispatch(context.Background(), viewInteraction)
// 	require.NoError(t, err)
// 	assert.True(t, viewHandled, "View submission handler was not called")

// 	// 4. Test Block Action Dispatch
// 	blockInteraction := slack.InteractionCallback{
// 		Type: slack.InteractionTypeBlockActions,
// 		User: slack.User{Name: "another-user"},
// 		ActionCallback: slack.ActionCallbacks{
// 			BlockActions: []*slack.BlockAction{
// 				{ActionID: "button-action", Value: "button_val"},
// 				{ActionID: "unhandled-action"}, // This should be ignored
// 			},
// 		},
// 	}
// 	err = dispatcher.Dispatch(context.Background(), blockInteraction)
// 	require.NoError(t, err)
// 	assert.True(t, blockActionHandled, "Block action handler was not called")

// 	// 5. Test Unregistered and Other Handler Errors
// 	t.Run("Unregistered View Handler", func(t *testing.T) {
// 		unregisteredInteraction := slack.InteractionCallback{
// 			Type: slack.InteractionTypeViewSubmission,
// 			View: slack.View{CallbackID: "unknown-callback-id"},
// 		}
// 		err = dispatcher.Dispatch(context.Background(), unregisteredInteraction)
// 		require.Error(t, err)
// 		assert.Contains(t, err.Error(), "no view handler registered")
// 	})

// 	t.Run("Mismatched Handler Type", func(t *testing.T) {
// 		badDispatcher := NewDispatcher()
// 		// Register a handler with the correct key but wrong type
// 		badDispatcher.viewHandlers["my_cool_modal"] = "not a handler"
// 		err := badDispatcher.Dispatch(context.Background(), viewInteraction)
// 		require.Error(t, err)
// 		assert.Contains(t, err.Error(), "has incorrect type")
// 	})
// }

// // normalizeNewlines replaces \r\n with \n to make string comparisons OS-agnostic.
// func normalizeNewlines(s string) string {
// 	return strings.ReplaceAll(s, "\r\n", "\n")
// }

// // These types are added to allow the test file to compile, as it cannot see
// // the generated types directly. The dispatcher test manually constructs them.
// // In a real-world scenario, you would have a more complex test setup.
// type MyCoolModalInputHandler interface {
// 	Handle(ctx context.Context, i slack.InteractionCallback, input MyCoolModalInput) error
// }

// func (d *Dispatcher) RegisterMyCoolModalInputHandler(h MyCoolModalInputHandler) {
// 	d.viewHandlers["my_cool_modal"] = h
// }
