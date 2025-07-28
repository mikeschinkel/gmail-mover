package gapi

import (
	"fmt"

	"google.golang.org/api/gmail/v1"
)

type GMailQuery struct {
	Labels      []string
	Before      string
	After       string
	Search      string
	MaxMessages int
}

// GetMessages retrieves messages from a specific label, handling pagination
func (gq GMailQuery) GetMessages(service *gmail.Service) (messages []*gmail.Message, err error) {
	var query string
	var req *gmail.UsersMessagesListCall
	var resp *gmail.ListMessagesResponse
	var pageToken string

	query = gq.BuildQueryString()
	if gq.MaxMessages == 0 {
		gq.MaxMessages = maxMessages
	}

	// Handle pagination to get all requested messages
	for {
		req = service.Users.Messages.List("me").Q(query).MaxResults(500) // Gmail API limit per page
		if pageToken != "" {
			req = req.PageToken(pageToken)
		}

		resp, err = req.Do()
		if err != nil {
			goto end
		}

		// Add messages from this page
		messages = append(messages, resp.Messages...)

		// Check if we have enough messages or if there are no more pages
		if len(messages) >= gq.MaxMessages || resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	// Trim to exact number requested
	if len(messages) > gq.MaxMessages {
		messages = messages[:gq.MaxMessages]
	}

end:
	return messages, err
}

// BuildQueryString constructs a Gmail search query from job configuration
func (gq GMailQuery) BuildQueryString() (query string) {
	for _, label := range gq.Labels {
		query = fmt.Sprintf("%s label:%s", query, label)
	}

	if gq.Before != "" {
		query = fmt.Sprintf("%s before:%s", query, gq.Before)
	}

	if gq.After != "" {
		query = fmt.Sprintf("%s after:%s", query, gq.After)
	}

	if gq.Search != "" {
		query = fmt.Sprintf("%s %s", query, gq.Search)
	}

	return query
}
