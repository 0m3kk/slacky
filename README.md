# **Slack Modal & Interaction Codegen for Go**

This tool is a command-line code generator that automates the creation of Go structs and a dispatcher for handling Slack Block Kit modal submissions and block actions. It reads Slack's JSON modal definitions and produces type-safe Go code, significantly simplifying the backend logic required to process Slack interactions.

## **Features**

* **Struct Generation**: Automatically creates Go structs from your modal's input blocks with correct Go types (string, int64, []string, etc.).
* **Dispatcher Generation**: Generates a central Dispatcher to route both view_submission and block_actions payloads to the correct, strongly-typed handlers.
* **Secure HTTP Handler**: Includes a ready-to-use http.Handler that automatically verifies Slack's signing secret for incoming requests.
* **Type-Safe Handlers**: Generates handler interfaces for each modal and for block actions, ensuring your business logic receives a pre-parsed struct with the submitted data.
* **Comprehensive Action Support**: The dispatcher correctly unmarshals all standard Slack input element types.
* **Context-Aware Handlers**: All generated handlers include context.Context and the full slack.InteractionCallback payload.
* **Validation**: Enforces the presence of callback_id and action_id, failing fast on invalid JSON definitions.
* **Customizable Output**: Use the -output flag to specify where to save the generated files.

## **How to Use**

### **Step 1: Define Your Slack Modal JSON**

Create one or more JSON files that define your Slack modals.

**Example: feedback_modal.json**

```json
{
  "type": "modal",
  "callback_id": "user_feedback_modal",
  "title": { "type": "plain_text", "text": "Feedback" },
  "submit": { "type": "plain_text", "text": "Submit" },
  "blocks": [
    {
      "type": "input",
      "element": {
        "type": "number_input",
        "is_decimal_allowed": false,
        "action_id": "rating"
      },
      "label": { "type": "plain_text", "text": "Overall Rating (1-5)" }
    },
    {
      "type": "input",
      "element": {
        "type": "plain_text_input",
        "action_id": "comments",
        "multiline": true
      },
      "label": { "type": "plain_text", "text": "Comments" }
    }
  ]
}
```

### **Step 2: Run the Code Generator**

Execute the program from your terminal, passing the paths to your JSON files. You can use the -output flag to change the destination directory.

# Generate code in the default 'generated' directory
```sh
go run . feedback_modal.json
```

# Generate code in a custom directory
```sh
go run . -output ./internal/slack/generated feedback_modal.json
```

### **Step 3: Implement Handlers and Run the Server**

In your application's main entry point, import the generated package, implement the handler interfaces, and use the provided NewInteractionHandler to run a secure web server.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/slack-go/slack"
    // Update this path to your generated code's location
    "path/to/your/project/generated_structs"
)

// Implement the view submission handler for 'user_feedback_modal'
type FeedbackHandler struct{}

func (h *FeedbackHandler) Handle(ctx context.Context, interaction slack.InteractionCallback, input generated.UserFeedbackModalInput) error {
    log.Printf(
        "[INFO] Received feedback from %s! Rating: %d, Comments: %s\n",
        interaction.User.Name,
        input.Rating, // This is an int64!
        input.Comments,
    )
    // Your business logic here...
    return nil
}

func main() {
    // 1. Get the signing secret from environment variables
    signingSecret := os.Getenv("SLACK_SIGNING_SECRET")
    if signingSecret == "" {
        log.Fatal("[FATAL] SLACK_SIGNING_SECRET must be set")
    }

    // 2. Create a new dispatcher from the generated package
    dispatcher := generated.NewDispatcher(signingSecret)

    // 3. Register your implemented handlers
    dispatcher.RegisterUserFeedbackModalHandler(&FeedbackHandler{})

    // 4. Create and run the secure HTTP handler
    log.Println("[INFO] Starting Slack Interaction server on :8080")
    http.HandleFunc("/slack/interactions", dispatcher.HandleInteraction)
    if err := http.ListenAndServe(":8080", handler); err != nil {
        log.Fatalf("[FATAL] Server failed to start: %v", err)
    }
}
```

### **Step 4: Configure Your Slack App**

In your Slack App's configuration dashboard under "Interactivity & Shortcuts", set your Request URL to https://your-public-url.com/slack/interactions.
