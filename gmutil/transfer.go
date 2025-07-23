package gmutil

import (
	"google.golang.org/api/gmail/v1"
)

var maxMessages = 10000

func SetMaxMessages(max int) {
	maxMessages = max
}

type TransferOpts struct {
	Labels          []string
	FailOnError     bool
	MaxMessages     int
	Before          string
	After           string
	SearchQuery     string
	LabelsToApply   []string
	DeleteAfterMove bool
	DryRun          bool
}

// TransferMessages handles the core message transfer logic
func TransferMessages(srcEmail, dstEmail string) (err error) {
	return TransferMessagesWithOpts(srcEmail, dstEmail, TransferOpts{})
}

// TransferMessagesWithOpts handles the core message transfer logic
func TransferMessagesWithOpts(srcEmail, dstEmail string, opts TransferOpts) (err error) {
	var messageCount int
	var label string
	var messages []*gmail.Message
	var message *gmail.Message
	var src, dst *gmail.Service

	ensureLogger()

	src, err = GetGmailService(srcEmail)
	if err != nil {
		goto end
	}

	dst, err = GetGmailService(dstEmail)
	if err != nil {
		goto end
	}

	logger.Info("Processing messages", "labels", opts.Labels)

	for _, label = range opts.Labels {
		gq := GMailQuery{
			Labels:      []string{label},
			Before:      opts.Before,
			After:       opts.After,
			Extra:       opts.SearchQuery,
			MaxMessages: opts.MaxMessages,
		}

		messages, err = gq.GetMessages(src)
		if err != nil {
			if !opts.FailOnError {
				logger.Error("Error getting messages", "label", label, "error", err)
				continue
			}
			goto end
		}

		for _, message = range messages {
			if messageCount >= opts.MaxMessages {
				logger.Error("Reached max message limit", "message_limit", opts.MaxMessages)
				goto end
			}

			err = transferMessage(src, dst, message, opts)
			if err != nil {
				if !opts.FailOnError {
					logger.Error("Error transferring message", "error", err)
					continue
				}
				goto end
			}

			messageCount++
		}
	}

end:
	logger.Info("Messages successfully transferred", "message_count", messageCount)
	return err
}

// transferMessage handles the transfer of a single message
func transferMessage(src, dst *gmail.Service, msg *gmail.Message, opts TransferOpts) (err error) {
	var fullMessage *gmail.Message
	var insertedMessage *gmail.Message

	if opts.DryRun {
		logger.Info("DRY RUN: Would move message", "src_id", msg.Id)
		goto end
	}

	// Get full msg content
	fullMessage, err = src.Users.Messages.Get("me", msg.Id).Format("raw").Do()
	if err != nil {
		goto end
	}

	// Insert into destination
	insertedMessage, err = dst.Users.Messages.Insert("me", &gmail.Message{
		Raw: fullMessage.Raw,
	}).Do()
	if err != nil {
		goto end
	}

	// Apply label if specified
	if len(opts.LabelsToApply) != 0 {

		err = applyLabels(dst, insertedMessage.Id, opts.LabelsToApply)
		if err != nil {
			goto end
		}
	}

	// Delete from source if requested
	if opts.DeleteAfterMove {
		err = src.Users.Messages.Delete("me", msg.Id).Do()
		if err != nil {
			goto end
		}
	}

	logger.Info("Moved message", "src_id", msg.Id, "dst_id", insertedMessage.Id)

end:
	return err
}
