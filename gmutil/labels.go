package gmutil

import (
	"fmt"

	"google.golang.org/api/gmail/v1"
)

// ListLabels lists all available labels for an email account
func ListLabels(email string) (err error) {
	var service *gmail.Service
	var labels []*gmail.Label
	var label *gmail.Label

	ensureLogger()

	service, err = GetGmailService(email)
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

// applyLabels applies a label to a message (simplified implementation)
func applyLabels(service *gmail.Service, messageID string, labels []string) (err error) {
	panic("IMPLEMENT ME")
}
