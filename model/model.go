package model

// --- Structs for parsing Slack Block Kit JSON ---
// These structs only define the fields necessary for the generator.

// SlackModal represents the top-level structure of a Slack modal view.
type SlackModal struct {
	CallbackID string       `json:"callback_id"`
	Blocks     []SlackBlock `json:"blocks"`
}

// SlackBlock represents a single block in the modal's layout.
// We are primarily interested in blocks of type "input".
type SlackBlock struct {
	Type    string        `json:"type"`
	Element *SlackElement `json:"element"` // Only relevant for "input" blocks
	Label   *SlackText    `json:"label"`   // For context, not directly used in generation
}

// SlackElement represents an interactive component within a block.
// For "input" blocks, this contains the actual input field (e.g., text input, select menu).
type SlackElement struct {
	Type             string `json:"type"`
	ActionID         string `json:"action_id"`
	IsDecimalAllowed bool   `json:"is_decimal_allowed,omitempty"` // For number_input
}

// SlackText represents a text object, used for labels and other text content.
type SlackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// FieldInfo holds the data for a single field in the generated struct.
type FieldInfo struct {
	Name    string
	GoType  string
	JSONTag string
}

// TemplateData is the data structure passed to the Go template.
type TemplateData struct {
	PackageName string
	StructName  string
	Fields      []FieldInfo
	CallbackID  string
}

// DispatcherTemplateData holds the data for the dispatcher template.
type DispatcherTemplateData struct {
	PackageName string
	Structs     []StructInfo
}

// StructInfo holds metadata about each generated struct for the dispatcher.
type StructInfo struct {
	CallbackID string // The original callback_id
	TypeName   string // The CamelCase struct type name (e.g., MyModalInput)
}
