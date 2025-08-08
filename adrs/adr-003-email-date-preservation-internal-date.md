# ADR-003: Email Date Preservation Using Gmail internalDate Parameter

**Status:** Accepted  
**Date:** 2025-08-08

## Context

When transferring emails between Gmail accounts, the Gmail API `users.messages.insert` method defaults to using the current timestamp rather than preserving the original email's send date. This causes transferred emails to appear at the wrong chronological position in the destination Gmail account.

## Decision

**Extract and preserve original email dates** by parsing the RFC822 `Date` header from email messages and passing it as the `internalDate` parameter to Gmail API insert operations.

## Implementation

1. **Parse Date Header**: Extract the `Date` header from the original email message using RFC2822 parsing
2. **Convert to Unix Timestamp**: Convert the parsed date to Unix milliseconds format required by Gmail API
3. **Set internalDate Parameter**: Pass the timestamp as the `internalDate` parameter when inserting messages via `users.messages.insert`

## Rationale

- **Chronological Accuracy**: Transferred emails appear at their correct historical position in Gmail's interface
- **User Experience**: Users can rely on Gmail's date-based sorting and filtering to work correctly with transferred messages
- **Data Integrity**: Preserves the original temporal context of email communications
- **Critical for Archival**: Without this, transferred emails become chronologically meaningless for long-term storage

## Implementation Evidence

This decision is implemented in the message transfer logic in `gapi/transfer.go` where date headers are extracted and passed to Gmail API calls.

## Consequences

**Positive:**
- Transferred emails maintain correct chronological order in Gmail interface
- Gmail's date-based search and filtering work correctly on transferred messages
- Preserves temporal integrity of email archives
- Essential for meaningful email archival and organization

**Trade-offs:**
- Requires parsing potentially malformed date headers from various email clients
- Adds complexity to the message transfer process
- Must handle edge cases where date headers are missing or invalid

## Alternative Considered

**Using current timestamp** was rejected because it renders transferred emails chronologically meaningless and breaks Gmail's time-based organization features.