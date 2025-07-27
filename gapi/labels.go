package gapi

import (
	"fmt"

	"google.golang.org/api/gmail/v1"
)

// ListLabels lists all available labels for an email account
func (api *GMailAPI) ListLabels(email string) (err error) {
	var service *gmail.Service
	var labels []*gmail.Label
	var label *gmail.Label

	ensureLogger()

	service, err = api.GetGmailService(email)
	if err != nil {
		goto end
	}

	labels, err = listLabels(service, "me")
	if err != nil {
		goto end
	}

	fmt.Printf("Available labels for %s:\n", email)
	for _, label = range labels {
		fmt.Printf("  %s\n", label.Name)
	}

end:
	return err
}

// ListLabels retrieves all labels for a Gmail account
func listLabels(service *gmail.Service, userID string) (labels []*gmail.Label, err error) {
	var resp *gmail.ListLabelsResponse

	resp, err = service.Users.Labels.List(userID).Do()
	if err != nil {
		goto end
	}

	labels = resp.Labels

end:
	return labels, err
}

// applyLabels applies labels to a message
func applyLabels(service *gmail.Service, messageID string, labels []string) (err error) {
	var labelIDs []string
	var labelName string
	var labelID string
	var existingLabels []*gmail.Label
	var existingLabel *gmail.Label
	var createLabelReq *gmail.Label
	var newLabel *gmail.Label

	ensureLogger()

	if len(labels) == 0 {
		goto end
	}

	// Get existing labels to find IDs
	existingLabels, err = listLabels(service, "me")
	if err != nil {
		goto end
	}

	// Convert label names to IDs, creating labels if they don't exist
	for _, labelName = range labels {
		labelID = ""

		// Find existing label
		for _, existingLabel = range existingLabels {
			if existingLabel.Name == labelName {
				labelID = existingLabel.Id
				break
			}
		}

		// Create label if it doesn't exist
		if labelID == "" {
			logger.Info("Creating new label", "label", labelName)
			createLabelReq = &gmail.Label{
				Name:                  labelName,
				MessageListVisibility: "show",
				LabelListVisibility:   "labelShow",
			}

			newLabel, err = service.Users.Labels.Create("me", createLabelReq).Do()
			if err != nil {
				logger.Error("Failed to create label", "label", labelName, "error", err)
				goto end
			}
			labelID = newLabel.Id
		}

		labelIDs = append(labelIDs, labelID)
	}

	// Apply the labels to the message
	_, err = service.Users.Messages.Modify("me", messageID, &gmail.ModifyMessageRequest{
		AddLabelIds: labelIDs,
	}).Do()
	if err != nil {
		logger.Error("Failed to apply labels", "message_id", messageID, "labels", labels, "error", err)
		goto end
	}

	logger.Info("Applied labels to message", "message_id", messageID, "labels", labels)

end:
	return err
}
