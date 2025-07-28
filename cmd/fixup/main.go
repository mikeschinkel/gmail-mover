package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/mikeschinkel/gmail-mover/gapi"
	"github.com/mikeschinkel/gmail-mover/gmover"
	"google.golang.org/api/gmail/v1"
)

func main() {
	if len(os.Args) != 3 {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s <email> <label>\n", os.Args[0])
		os.Exit(1)
	}

	email := os.Args[1]
	label := os.Args[2]

	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Initialize gmover package
	err := gmover.Initialize(&gmover.Opts{
		Logger: logger,
	})
	if err != nil {
		logger.Error("Failed to initialize", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Get Gmail service
	api := gapi.NewGMailAPI("gmover", gmover.ConfigFileStore())
	service, err := api.GetGmailService(string(gmover.EmailAddress(email)))
	if err != nil {
		logger.Error("Failed to get Gmail service", "error", err)
		os.Exit(1)
	}

	// Find messages with the specified label, excluding already fixed ones
	query := fmt.Sprintf("label:%s -label:sent-date-fixed", label)
	listReq := service.Users.Messages.List("me").Q(query)

	logger.Info("Searching for messages", "query", query)

	for {
		resp, err := listReq.Do()
		if err != nil {
			logger.Error("Failed to list messages", "error", err)
			os.Exit(1)
		}
		api := gapi.NewGMailAPI(gmover.AppName, gmover.ConfigFileStore())
		for _, msg := range resp.Messages {
			info, err := api.GetMessageInfo(service, msg)
			if err != nil {
				logger.Error("Failed to get message", "service_path", service.BasePath, "error", err)
				continue
			}
			logger.Info("Fixing message", "message", info.String())
			err = fixMessage(ctx, service, msg.Id, logger)
			if err != nil {
				logger.Error("Failed to fix message", "id", msg.Id, "error", err)
				continue
			}

		}

		if resp.NextPageToken == "" {
			break
		}
		listReq.PageToken(resp.NextPageToken)
	}

	logger.Info("Fixup complete")
}

func fixMessage(_ context.Context, service *gmail.Service, msgId string, logger *slog.Logger) (err error) {
	var msg *gmail.Message
	var newMsg *gmail.Message
	var originalMsg *gmail.Message
	var fixupLabel *gmail.Label
	var labelsToAdd []string

	// Get the full message
	msg, err = service.Users.Messages.Get("me", msgId).Format("raw").Do()
	if err != nil {
		goto end
	}

	// Get the labels from the original message
	originalMsg, err = service.Users.Messages.Get("me", msgId).Do()
	if err != nil {
		goto end
	}

	// Re-insert the message with correct date handling
	newMsg, err = service.Users.Messages.Insert("me", &gmail.Message{
		Raw: msg.Raw,
	}).InternalDateSource("dateHeader").Do()
	if err != nil {
		goto end
	}

	// Apply the same labels to the new message plus our marker label
	labelsToAdd = originalMsg.LabelIds

	// Create or get the fixup marker label
	fixupLabel, err = ensureLabel(service, "sent-date-fixed")
	if err != nil {
		goto end
	}
	labelsToAdd = append(labelsToAdd, fixupLabel.Id)

	if len(labelsToAdd) > 0 {
		_, err = service.Users.Messages.Modify("me", newMsg.Id, &gmail.ModifyMessageRequest{
			AddLabelIds: labelsToAdd,
		}).Do()
		if err != nil {
			goto end
		}
	}

	// Delete the original message
	err = service.Users.Messages.Delete("me", msgId).Do()
	if err != nil {
		goto end
	}

	logger.Info("Fixed message", "old_id", msgId, "new_id", newMsg.Id)

end:
	return err
}

// ensureLabel creates a label if it doesn't exist, returns existing label if it does
func ensureLabel(service *gmail.Service, labelName string) (label *gmail.Label, err error) {
	// Try to find existing label
	labelsResp, err := service.Users.Labels.List("me").Do()
	if err != nil {
		goto end
	}

	for _, existingLabel := range labelsResp.Labels {
		if existingLabel.Name == labelName {
			label = existingLabel
			goto end
		}
	}

	// Create the label if it doesn't exist
	label, err = service.Users.Labels.Create("me", &gmail.Label{
		Name:                  labelName,
		LabelListVisibility:   "labelShow",
		MessageListVisibility: "show",
	}).Do()

end:
	return label, err
}
