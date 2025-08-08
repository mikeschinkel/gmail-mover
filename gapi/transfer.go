package gapi

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/api/gmail/v1"
)

var maxMessages = 10000

//goland:noinspection GoUnusedExportedFunction
func SetMaxMessages(max int) {
	maxMessages = max
}

type approvalResponse = byte

//goland:noinspection GoUnusedConst
const (
	YesResponse    = 'y'
	NoResponse     = 'n'
	AllResponse    = 'a'
	DelayResponse  = 'd'
	CancelResponse = 'c'
)

type ApprovalFunc = func(context.Context, string) (ar approvalResponse, err error)

// MoveLogEntry represents a single email move operation for JSON logging
type MoveLogEntry struct {
	Timestamp   time.Time    `json:"timestamp"`
	MessageID   string       `json:"msg_id"`
	Subject     string       `json:"subject"`
	Date        string       `json:"date"`
	To          string       `json:"to"`
	From        string       `json:"from"`
	SrcEmail    EmailAddress `json:"src"`
	DstEmail    EmailAddress `json:"dst"`
	SrcLabels   []string     `json:"src_labels"`
	DstLabels   []string     `json:"dst_labels"`
	Moved       bool         `json:"moved"`
	Deleted     bool         `json:"deleted"`
	Labeled     bool         `json:"labeled"`
	DateParsed  bool         `json:"date_parsed"`
	Error       error        `json:"error,omitempty"`
	MessageInfo `json:"-"`
}

type MoveLogger interface {
	LogMove(MoveLogEntry) error
}

type TransferOpts struct {
	Labels          []string
	LabelsToApply   []string
	LabelsToRemove  []string
	Before          string
	After           string
	SearchQuery     string
	MaxMessages     int
	DeleteAfterMove bool
	DryRun          bool
	FailOnError     bool
	MoveLogger      MoveLogger
	ApprovalFunc    ApprovalFunc // Optional - if nil, auto-approve all messages
	ApprovalPrompt  string       // Optional - if nil, auto-approve all messages
	sameAccount     bool         // Internal flag for same-account operations
	approvalResponse
}

// TransferMessages handles the core message transfer logic
func (api *GMailAPI) TransferMessages(ctx context.Context, srcEmail, dstEmail EmailAddress, opts TransferOpts) (err error) {
	var messageCount int
	var label string
	var messages []*gmail.Message
	var message *gmail.Message
	var src, dst *gmail.Service
	var msgInfo MessageInfo

	ensureLogger()

	// Check if this is a same-account operation
	if srcEmail == dstEmail {
		opts.sameAccount = true
	}

	src, err = api.GetGmailService(srcEmail)
	if err != nil {
		goto end
	}

	if opts.sameAccount {
		dst = src // Use same service for same-account operations
	} else {
		dst, err = api.GetGmailService(dstEmail)
		if err != nil {
			goto end
		}
	}

	logger.Info("Processing messages", "labels", opts.Labels)

	for _, label = range opts.Labels {
		gq := GMailQuery{
			Labels:      []string{label},
			Before:      opts.Before,
			After:       opts.After,
			Search:      opts.SearchQuery,
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

		if opts.DryRun {
			logger.Info("DRY RUN")
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
				logger.Error("Reached max email limit", "message_limit", opts.MaxMessages)
				goto end
			}

			msgInfo, err = api.transferMessage(ctx, src, dst, message, &opts)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					goto end
				}

				if !opts.FailOnError {
					logger.Error("Error transferring email", "error", err)
					continue
				}
				goto end
			}
			if !opts.DryRun && opts.MoveLogger != nil {
				func() {
					err := opts.MoveLogger.LogMove(MoveLogEntry{
						Timestamp:  time.Now(),
						MessageID:  msgInfo.Id,
						Subject:    msgInfo.Subject,
						Date:       msgInfo.Date.Format(time.DateTime),
						To:         msgInfo.To,
						From:       msgInfo.From,
						SrcEmail:   srcEmail,
						DstEmail:   dstEmail,
						SrcLabels:  opts.Labels,
						DstLabels:  opts.LabelsToApply,
						Moved:      msgInfo.Moved,
						Deleted:    msgInfo.Deleted,
						Labeled:    msgInfo.Labeled,
						DateParsed: msgInfo.DateParsed,
						Error:      err,
					})
					if err != nil {
						logger.Warn("Error logging email message move", "error", err, "email", msgInfo)
					}
				}()
			}
			messageCount++
		}
	}

end:
	logger.Info("Messages successfully transferred", "message_count", messageCount)
	return err
}

// transferMessage handles the transfer of a single message
func (api *GMailAPI) transferMessage(ctx context.Context, src, dst *gmail.Service, message *gmail.Message, opts *TransferOpts) (msgInfo MessageInfo, err error) {
	var fullMessage *gmail.Message
	var insertedMessage *gmail.Message

	// Get message details for approval (always needed for logging)
	msgInfo, err = api.GetMessageInfo(src, message)
	if err != nil {
		goto end
	}

	if !opts.DryRun {
		logger.Info("Email to move", "email", msgInfo.String())
	}

	switch {
	case opts.DryRun:
		if opts.sameAccount {
			logger.Info("DRY RUN: Would change labels on message",
				"msg_id", msgInfo.Id,
				"subject", msgInfo.Subject,
				"sender", msgInfo.From,
				"add_labels", opts.LabelsToApply,
				"remove_labels", opts.LabelsToRemove,
			)
		} else {
			logger.Info("DRY RUN: Would move message",
				"src_id", msgInfo.Id,
				"subject", msgInfo.Subject,
				"sender", msgInfo.From,
			)
		}
		goto end

	case opts.approvalResponse == DelayResponse:
		logger.Info("Pausing 3 seconds before next email transfer. Press Ctrl-C to terminate")
		time.Sleep(time.Second * 3)

	case opts.approvalResponse == AllResponse:
		// Disable further prompts
		logger.Info("Auto-approving remaining messages")

	case opts.approvalResponse == CancelResponse:
		err = context.Canceled
		goto end

	default:
		// Check approval if ApprovalFunc is set
		if opts.ApprovalFunc == nil {
			err = fmt.Errorf("no approval func specified")
			goto end
		}
		opts.approvalResponse, err = opts.ApprovalFunc(ctx,
			fmt.Sprintf("%s %s", opts.ApprovalPrompt, msgInfo.String()),
		)
		if err != nil {
			goto end
		}
		switch opts.approvalResponse {
		case NoResponse:
			logger.Info("Message skipped by user", "subject", msgInfo.Subject, "id", msgInfo.Id)
			goto end
		case DelayResponse:
			logger.Info("Setting delay to 3 seconds between messages; press Ctrl-C to terminate")
		case AllResponse:
			logger.Info("Auto-approving remaining message moves; press Ctrl-C to terminate")
		}
	}

	// Don't hammer the Gmail server
	time.Sleep(time.Millisecond * 100)

	if opts.sameAccount {
		// Same-account operation: modify labels instead of insert/delete
		logger.Info("Changing labels...")

		err = modifyLabels(dst, msgInfo.Id, opts.LabelsToApply, opts.LabelsToRemove)
		if err != nil {
			goto end
		}
		msgInfo.Moved = true
		msgInfo.Labeled = true

		logger.Info("Changed labels on message", "msg_id", msgInfo.Id)
	} else {
		// Cross-account operation: use insert/delete approach
		logger.Info("Transferring email...")

		// Get full msgInfo content
		fullMessage, err = src.Users.Messages.Get("me", msgInfo.Id).Format("raw").Do()
		if err != nil {
			goto end
		}

		// Insert into destination with original email date preserved
		insertedMessage, err = dst.Users.Messages.Insert("me", &gmail.Message{
			Raw: fullMessage.Raw,
		}).InternalDateSource("dateHeader").Do()
		if err != nil {
			goto end
		}
		msgInfo.Moved = true

		// Apply label if specified
		if len(opts.LabelsToApply) != 0 {
			err = applyLabels(dst, insertedMessage.Id, opts.LabelsToApply)
			if err != nil {
				goto end
			}
		}
		msgInfo.Labeled = true

		// Delete from source if requested
		if opts.DeleteAfterMove {
			err = src.Users.Messages.Delete("me", msgInfo.Id).Do()
			if err != nil {
				goto end
			}
		}
		msgInfo.Deleted = true

		logger.Info("Moved message", "src_id", msgInfo.Id, "dst_id", insertedMessage.Id)
	}

end:
	return msgInfo, err
}

// GetMessageInfo extracts message details for approval decisions
func (api *GMailAPI) GetMessageInfo(service *gmail.Service, msg *gmail.Message) (msgInfo MessageInfo, err error) {
	var fullMessage *gmail.Message
	var header *gmail.MessagePartHeader

	// Get message with headers
	fullMessage, err = service.Users.Messages.Get("me", msg.Id).Format("metadata").Do()
	if err != nil {
		goto end
	}

	msgInfo.Id = msg.Id

	// Extract headers
	for _, header = range fullMessage.Payload.Headers {
		switch header.Name {
		case "Subject":
			msgInfo.Subject = header.Value
		case "From":
			msgInfo.From = header.Value
		case "To":
			msgInfo.To = header.Value
		case "Date":
			// Parse RFC2822 date format with multiple fallback formats
			msgInfo.Date, msgInfo.DateParsed = parseEmailDate(header.Value)
		}
	}

	// Fallback values
	if msgInfo.Subject == "" {
		msgInfo.Subject = "(no subject)"
	}
	if msgInfo.From == "" {
		msgInfo.From = "(unknown sender)"
	}
	if msgInfo.To == "" {
		msgInfo.To = "(unknown recipient)"
	}
	// Date fallback is handled in parseEmailDate function

end:
	return msgInfo, err
}
