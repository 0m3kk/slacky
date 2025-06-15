package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/om3kk/slacky/example/generated"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
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
	})

	port := ":8080"
	logger.Infof("Server listening on port %s...", port)
	logger.Infof("Make sure your Slack App's Request URL is set to your public endpoint, e.g., https://your-ngrok-url.ngrok-free.app/slack/interactive")
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
	dispatcher := generated.NewDispatcher(secret)
	dispatcher.RegisterFullInputModalInputHandler(&Handler{})
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

type Handler struct{}

func (h *Handler) HandleFullInputModal(ctx context.Context, interaction slack.InteractionCallback, input generated.FullInputModalInput) ([]generated.SlackErrorResp, error) {
	inputJson, _ := json.MarshalIndent(input, "", "\t")
	logger.Infof("Received input: %s", string(inputJson))
	return nil, nil
}
