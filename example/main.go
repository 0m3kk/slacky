package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"

	"github.com/0m3kk/slacky/example/generated"
)

var logger = logrus.New()

func main() {
	logger.SetFormatter(&logrus.JSONFormatter{})

	slackClient, err := newSlackClient()
	if err != nil {
		logger.Fatalf("failed to create Slack client: %v", err)
	}

	dispatcher, err := newDispatcher()
	if err != nil {
		logger.Fatalf("failed to create dispatcher: %v", err)
	}

	view, err := loadModalView()
	if err != nil {
		logger.Fatalf("failed to load modal view: %v", err)
	}

	blocks, err := loadMessageBlocks()
	if err != nil {
		logger.Fatalf("failed to load blocks: %v", err)
	}

	http.HandleFunc("/slack/interactive", dispatcher.HandleInteraction)
	http.HandleFunc("/slack/events", func(w http.ResponseWriter, r *http.Request) {
		command, err := slack.SlashCommandParse(r)
		if err != nil {
			logger.Errorf("failed to parse slash command: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if _, err := slackClient.OpenView(command.TriggerID, view); err != nil {
			logger.Errorf("failed to open view: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if _, _, _, err := slackClient.SendMessage(command.ChannelID, slack.MsgOptionBlocks(blocks...)); err != nil {
			logger.Errorf("failed to send blocks: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

	port := ":8080"
	logger.Infof("Server listening on port %s...", port)
	logger.Infof(
		"Make sure your Slack App's Request URL is set to your public endpoint, e.g., https://your-ngrok-url.ngrok-free.app/slack/interactive",
	)
	if err := http.ListenAndServe(port, nil); err != nil {
		logger.Fatalf("failed to start server: %v", err)
	}
}

func newSlackClient() (*slack.Client, error) {
	token := os.Getenv("SLACK_AUTH_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("SLACK_AUTH_TOKEN environment variable is not set")
	}
	return slack.New(token), nil
}

func newDispatcher() (*generated.Dispatcher, error) {
	secret := os.Getenv("SLACK_SIGNING_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("SLACK_SIGNING_SECRET environment variable is not set")
	}
	handler := &Handler{}
	dispatcher := generated.NewDispatcher(secret)
	dispatcher.RegisterFullInputModalInputHandler(handler)
	dispatcher.RegisterBlockActionHandlers(handler)
	return dispatcher, nil
}

func loadModalView() (slack.ModalViewRequest, error) {
	modalJson, err := os.ReadFile("./example/modal.json")
	if err != nil {
		return slack.ModalViewRequest{}, fmt.Errorf("failed to read modal JSON: %v", err)
	}

	var view slack.ModalViewRequest
	if err := json.Unmarshal(modalJson, &view); err != nil {
		return slack.ModalViewRequest{}, fmt.Errorf("failed to unmarshal modal view: %v", err)
	}
	return view, nil
}

func loadMessageBlocks() ([]slack.Block, error) {
	messageJson, err := os.ReadFile("./example/message.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read blocks JSON: %v", err)
	}

	// Define an auxiliary struct to unmarshal the outer structure
	var message struct {
		Blocks json.RawMessage `json:"blocks"` // Unmarshal into RawMessage first
	}
	if err := json.Unmarshal(messageJson, &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message blocks raw message: %v", err)
	}

	var slackBlocks slack.Blocks
	blocksJson, err := json.Marshal(message.Blocks)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal blocks to raw: %v", err)
	}
	if err = json.Unmarshal(blocksJson, &slackBlocks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message blocks: %v", err)
	}

	return slackBlocks.BlockSet, nil
}

type Handler struct{}

func (h *Handler) HandleFullInputModal(
	ctx context.Context,
	interaction slack.InteractionCallback,
	input generated.FullInputModalInput,
) error {
	inputJson, _ := json.MarshalIndent(input, "", "\t")
	logger.Infof("Received input: %s", string(inputJson))
	return nil
}

func (h *Handler) HandleButtonDeny(
	ctx context.Context,
	interaction slack.InteractionCallback,
	action slack.BlockAction,
) error {
	fmt.Println("Deny button clicked", action.Value)
	return nil
}

func (h *Handler) HandleButtonMoreInfo(
	ctx context.Context,
	interaction slack.InteractionCallback,
	action slack.BlockAction,
) error {
	fmt.Println("More info button clicked", action.Value)
	return nil
}

func (h *Handler) HandleButtonApprove(
	ctx context.Context,
	interaction slack.InteractionCallback,
	action slack.BlockAction,
) error {
	fmt.Println("Approve button clicked", action.Value)
	return nil
}
