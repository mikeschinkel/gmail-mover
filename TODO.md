# Gmail Mover - Implementation TODOs

## applyLabels Function Implementation

The `applyLabels` function in `gmutil/labels.go:50` currently contains `panic("IMPLEMENT ME")` and needs a proper implementation.

### Requirements:
1. Check if the specified label exists in the destination Gmail account
2. Create the label if it doesn't exist (based on job config `CreateLabelIfMissing` flag)
3. Apply the label(s) to the specified message using Gmail API
4. Handle errors gracefully (don't panic on missing labels if creation is disabled)

### Gmail API Reference:
- `service.Users.Labels.List()` - to check existing labels
- `service.Users.Labels.Create()` - to create new labels if needed  
- `service.Users.Messages.Modify()` - to apply labels to messages

### Current Usage:
Called from `gmutil/transfer.go:119` during message transfer when `opts.LabelsToApply` is not empty.

### Error Handling:
Should follow Clear Path style with `goto end` pattern and return meaningful errors for debugging.