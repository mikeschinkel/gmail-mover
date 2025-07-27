package gapi

import (
	"context"
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"

	"google.golang.org/api/gmail/v1"
)

var maxMessages = 10000

//goland:noinspection GoUnusedExportedFunction
func SetMaxMessages(max int) {
	maxMessages = max
}

type TransferOpts struct {
	Labels          []string
	LabelsToApply   []string
	Before          string
	After           string
	SearchQuery     string
	MaxMessages     int
	DeleteAfterMove bool
	DryRun          bool
	FailOnError     bool
	ApprovalFunc    // Optional - if nil, auto-approve all messages
}

// TransferMessages handles the core message transfer logic
func (api *GMailAPI) TransferMessages(ctx context.Context, srcEmail, dstEmail string, opts TransferOpts) (err error) {
	var messageCount int
	var label string
	var messages []*gmail.Message
	var message *gmail.Message
	var src, dst *gmail.Service

	ensureLogger()

	src, err = api.GetGmailService(srcEmail)
	if err != nil {
		goto end
	}

	dst, err = api.GetGmailService(dstEmail)
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
			// Check for cancellation
			select {
			case <-ctx.Done():
				logger.Info("Operation cancelled by user")
				err = ctx.Err()
				goto end
			default:
			}

			if messageCount >= opts.MaxMessages {
				logger.Error("Reached max message limit", "message_limit", opts.MaxMessages)
				goto end
			}

			err = api.transferMessage(ctx, src, dst, message, opts)
			if err != nil {
				// Check if this is a terminal/input error that should abort everything
				if IsTerminalError(err) {
					logger.Error("Terminal error - cannot continue", "error", err)
					goto end
				}

				if errors.Is(err, context.Canceled) {
					goto end
				}

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
func (api *GMailAPI) transferMessage(ctx context.Context, src, dst *gmail.Service, msg *gmail.Message, opts TransferOpts) (err error) {
	var fullMessage *gmail.Message
	var insertedMessage *gmail.Message
	var messageInfo MessageInfo
	var approved, approveAll bool

	// Get message details for approval (always needed for logging)
	messageInfo, err = api.getMessageInfo(src, msg)
	if err != nil {
		goto end
	}

	logger.Info("Transferring message", "message", messageInfo.String())

	// Check approval if ApprovalFunc is set
	if opts.ApprovalFunc == nil {
		err = fmt.Errorf("no approval func specified")
		goto end
	}
	approved, approveAll, err = opts.ApprovalFunc(ctx, messageInfo.String())
	if err != nil {
		goto end
	}
	if !approved {
		logger.Info("Message skipped by user", "subject", messageInfo.Subject, "id", msg.Id)
		goto end
	}
	if approveAll {
		// Disable further prompts
		opts.ApprovalFunc = nil
		logger.Info("Auto-approving remaining messages")
	}

	if opts.DryRun {
		logger.Info("DRY RUN: Would move message", "src_id", msg.Id, "subject", messageInfo.Subject)
		goto end
	}

	// Get full msg content
	fullMessage, err = src.Users.Messages.Get("me", msg.Id).Format("raw").Do()
	if err != nil {
		goto end
	}

	//panic("ADD INFO LOGGING SO WE CAN SEE WHAT IS HAPPENING")

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

// getMessageInfo extracts message details for approval decisions
func (api *GMailAPI) getMessageInfo(service *gmail.Service, msg *gmail.Message) (info MessageInfo, err error) {
	var fullMessage *gmail.Message
	var header *gmail.MessagePartHeader

	// Get message with headers
	fullMessage, err = service.Users.Messages.Get("me", msg.Id).Format("metadata").Do()
	if err != nil {
		goto end
	}

	info.Id = msg.Id

	// Extract headers
	for _, header = range fullMessage.Payload.Headers {
		switch header.Name {
		case "Subject":
			info.Subject = header.Value
		case "From":
			info.From = header.Value
		case "To":
			info.To = header.Value
		case "Date":
			// Parse RFC2822 date format
			info.Date, _ = time.Parse(time.RFC1123Z, header.Value)
			if info.Date.IsZero() {
				// Try alternative format
				info.Date, _ = time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", header.Value)
			}
		}
	}

	// Fallback values
	if info.Subject == "" {
		info.Subject = "(no subject)"
	}
	if info.From == "" {
		info.From = "(unknown sender)"
	}
	if info.To == "" {
		info.To = "(unknown recipient)"
	}

end:
	return info, err
}

// IsTerminalError checks if an error is related to terminal/input operations
// These errors should abort the entire operation rather than continue
func IsTerminalError(err error) (isTermErr bool) {
	var pathErr *os.PathError
	var errno syscall.Errno
	var ok bool

	if err == nil {
		goto end
	}

	switch {
	case errors.As(err, &pathErr):
		// Check for syscall errors related to terminal operations
		errno, ok = pathErr.Err.(syscall.Errno)
		if ok && errno == syscall.ENOTTY { // ENOTTY = "inappropriate ioctl for device"
			isTermErr = true
			goto end
		}
	case errors.As(err, &errno):
		// Check for direct syscall errors
		if errno == syscall.ENOTTY { // ENOTTY = "inappropriate ioctl for device"
			isTermErr = true
			goto end
		}
	}

end:
	return isTermErr
}
