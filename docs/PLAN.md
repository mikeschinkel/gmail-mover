# Gmail Mover CLI Tool – Project Brief and Development Plan

IMPORTANT: THIS IS ORIGINAL BRAINSTORMING: 2025-07-23 

## Overview

The Gmail Mover CLI is a standalone Go-based command-line tool that transfers email messages from one Gmail account to another. It is designed for:

* Archiving messages from a user's active Gmail accounts into a dedicated archive account
* Manual, repeatable use
* Easy future extension into larger automation and archival tooling

This tool relies on the Gmail API and OAuth 2.0 credentials to authenticate both the source and destination accounts. It supports copy/paste interactive authorization to avoid complex callback infrastructure.

---

## Functional Requirements

### ✅ Core Features (Implemented)

* Authenticate to both source and destination Gmail accounts using `credentials.json` and `token.json`
* Download emails from source account's INBOX using the Gmail API in `raw` format (base64-encoded RFC822)
* Re-upload messages into destination account using Gmail API `Users.Messages.Insert`
* Use copy-paste OAuth flow with saved token caching
* Print a console log of transferred message IDs

### 🔜 Future Enhancements (Planned, Not Yet Implemented)

* CLI flags for:

    * Label/query filters (e.g., `--label=INBOX`, `--query=from:amazon`)
    * Max messages to process (e.g., `--max=100`)
    * Date filters (e.g., `--before=2022-01-01`)
    * Destination label to apply (e.g., `--apply-label=archived-from-main`)
    * Delete message from source after copy
* Logging to a local SQLite DB to prevent duplicate moves
* Summary report of moved messages (e.g., from, subject, size)
* Support for batch operations via Gmail API batch mode

---

## Directory and File Layout

```
gmail-mover/
├── main.go                  # CLI entry point
├── auth/
│   └── auth.go              # OAuth2 loading, token management, manual authorization
├── mover/
│   └── mover.go             # Gmail message listing, downloading, and uploading
├── credentials/
│   ├── source_credentials.json
│   └── dest_credentials.json
├── token/
│   ├── token_source.json    # Auto-generated after interactive auth
│   └── token_dest.json      # Auto-generated after interactive auth
```

---

## Authentication Details

* Uses Google's OAuth2 Desktop App flow
* Tokens are persisted locally after first run to avoid repeated authorization
* Copy/paste model avoids needing an HTTP redirect server or callback URL
* Credentials must be generated via Google Cloud Console for each account:

    * Enable Gmail API
    * Create OAuth 2.0 Client ID (Desktop)

---

## Technical Stack

* **Language:** Go (latest stable)
* **APIs:** Gmail API via `google.golang.org/api/gmail/v1`
* **OAuth2:** `golang.org/x/oauth2` + `google.ConfigFromJSON`
* **Data Format:** `RAW` format of email (RFC822) via base64-encoded string

---

## Development Tasks Completed

* [x] Gmail API client setup for both source and destination accounts
* [x] Token retrieval and storage with manual auth flow
* [x] Listing messages by label ID
* [x] Downloading raw messages from source
* [x] Uploading raw messages to destination
* [x] CLI output showing progress of moved messages

---

## Tasks Not Yet Started

* [ ] Add `flag`-based CLI argument parsing
* [ ] Add filtering by Gmail query syntax
* [ ] Add date-based filtering
* [ ] Optionally delete messages after move
* [ ] Apply specific label to moved messages
* [ ] SQLite journal for deduplication
* [ ] Logging enhancements and dry-run mode

---

## Usage Summary

After downloading `credentials.json` for each Gmail account:

```sh
$ go run main.go
```

On first run, the tool will prompt the user to authorize access via a URL. After granting access, the user pastes the returned code into the terminal. This will generate `token.json` files for reuse.

Subsequent runs will reuse the saved tokens, and begin moving messages from the source INBOX to the destination account.

---

## Design Considerations

* Emphasis on **manual control**, **user transparency**, and **auditability**
* Designed for personal use, with future extensibility for automation
* Ensures full message fidelity by using raw MIME
* Avoids Gmail filters/labels from interfering by running client-side

---

## Target User

This tool is for advanced users who:

* Have multiple Gmail accounts
* Want to archive or migrate messages in bulk
* Are comfortable using the terminal and managing OAuth tokens manually
* May later want to build an automated email archival system with more features

---

## Author's Next Steps

* Integrate command-line options
* Implement SQLite logging and deduplication
* Extend tool into an ecosystem of archival and triage helpers (e.g. delayed rules, document sync)

---

This document reflects only the functionality and architecture discussed so far and avoids speculation or design beyond current scope.