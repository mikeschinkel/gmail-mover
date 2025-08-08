# CHAT HISTORY #3 — Gmail API Email Date Preservation


I just discovered something troubling. When I move an email it updates the date. If I look at its raw message form it adds a line like this to the top:Received: from 302033755609 named unknown by gmailapi.google.com with HTTPREST; Sun, 27 Jul 2025 03:34:24 -0700Is it not possible to retain its original date?EditYou've discovered a significant limitation with Gmail's API when using the insert method. When Gmail receives a message through the API, it treats it as a newly received email and adds that Received: header with the current timestamp, which affects how the message appears in the interface.
Unfortunately, there's no way to completely preserve the original received date when moving emails between Gmail accounts using the Gmail API. This is a known limitation that affects all tools that transfer emails via Gmail's API.
However, there are a few approaches that can help minimize the impact:
Potential Solutions:

Use internalDate parameter: When calling Users.Messages.Insert, you can specify the internalDate parameter to set Gmail's internal timestamp. This won't remove the Received: header but may help with sorting.
IMAP approach: Instead of Gmail API, use IMAP with the APPEND command, which sometimes preserves dates better (though Gmail may still add received headers).
Accept the limitation: Many users find that while the received date changes, the original Date: header from the email (when it was actually sent) remains intact, so the chronological order of conversations is usually preserved.

Quick Check:
In your current implementation, are you passing any date-related parameters to the Gmail API insert call? And more importantly - is the original Date: header (showing when the email was originally sent) being preserved in the message body?
Would you like me to show you how to modify your code to use the internalDate parameter, or would you prefer to explore the IMAP approach as an alternative?RetryMSI don't mind the header, what I mind is seeing the message in the gmail interface with the wrong date.EditAh, I understand - you're seeing the messages appear with today's date in the Gmail interface instead of their original dates, which makes them useless for chronological organization.
The good news is this can likely be fixed! The issue is probably that you're not setting the internalDate parameter when calling Users.Messages.Insert. Gmail uses this internal date for display purposes in the interface.
The Solution:
When inserting messages, you need to extract the original date from the email headers and pass it as the internalDate parameter. Here's what needs to happen:

Extract the original date: Parse the Date: header from the raw email message
Convert to Unix timestamp: Gmail expects internalDate as milliseconds since epoch
Pass it to the insert call: Include this timestamp in your API request

Implementation:
The key changes would be in your mover.go file where you're calling the Gmail insert method. Instead of just inserting the raw message, you'd:
go// Parse the Date header from the raw message
originalDate := extractDateFromMessage(rawMessage)
// Convert to milliseconds since epoch
internalDateMs := originalDate.UnixNano() / int64(time.Millisecond)

// Use it in the insert call
call := service.Users.Messages.Insert(userID, &gmail.Message{
Raw: rawMessage,
}).InternalDateSource("dateHeader").InternalDate(internalDateMs)
Would you like me to help you implement this fix in your existing code? I'd need to see your current message insertion logic to show you exactly where to make the changes.RetryMSI am using Claude Code to update. Can you give me instructions for Claude Code to make this change?EditHere are the instructions for Claude Code to implement the internalDate fix:
Claude Code Instructions:
Update the Gmail message insertion code to preserve original email dates in the Gmail interface.

Requirements:
1. Extract the original Date header from each email message before inserting
2. Convert the date to Unix milliseconds timestamp
3. Pass this timestamp as the internalDate parameter to the Gmail API insert call
4. Use Go's time parsing to handle RFC2822 date formats commonly found in email headers

Implementation details:
- Add a helper function to parse the Date header from raw email messages
- Modify the message insertion logic in mover.go to use the extracted date
- Handle cases where Date header might be missing or malformed (fallback to current time)
- Ensure the internalDate is passed as milliseconds since Unix epoch as required by Gmail API

The goal is to make transferred emails appear in Gmail with their original send dates rather than the current date/time.
This should give Claude Code enough context to:

Locate your current message insertion code
Add the date extraction logic
Modify the Gmail API call to include the internalDate parameter
Handle edge cases appropriately

The key technical detail is that Gmail expects the internalDate as milliseconds since Unix epoch, and email Date headers are typically in RFC2822 format which Go can parse with time.Parse(time.RFC1123Z, dateString) or similar.