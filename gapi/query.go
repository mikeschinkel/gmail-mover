package gapi

import (
	"fmt"

	"google.golang.org/api/gmail/v1"
)

type GMailQuery struct {
	Labels      []string
	Before      string
	After       string
	Extra       string
	MaxMessages int
}

// GetMessages retrieves messages from a specific label
func (gq GMailQuery) GetMessages(service *gmail.Service) (messages []*gmail.Message, err error) {
	var query string
	var req *gmail.UsersMessagesListCall
	var resp *gmail.ListMessagesResponse

	query = gq.BuildQueryString()
	if gq.MaxMessages == 0 {
		gq.MaxMessages = maxMessages
	}
	req = service.Users.Messages.List("me").Q(query).MaxResults(int64(gq.MaxMessages))
	resp, err = req.Do()
	if err != nil {
		goto end
	}

	messages = resp.Messages

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

	if gq.Extra != "" {
		query = fmt.Sprintf("%s %s", query, gq.Extra)
	}

	return query
}
