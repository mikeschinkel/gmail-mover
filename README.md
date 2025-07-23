# Gmail Mover

A fast, reliable command-line tool for transferring Gmail messages between accounts with advanced filtering and dry-run support.

## Features

- **Multi-Account Support**: Transfer messages between different Gmail accounts
- **Advanced Filtering**: Use Gmail's powerful query syntax, labels, and date ranges
- **Dry-Run Mode**: Preview what will be moved before making changes
- **Label Management**: Apply labels to moved messages and list available labels
- **Flexible Configuration**: Use command-line flags or JSON job files
- **OAuth2 Security**: Secure authentication with Google's standard OAuth2 flow
- **Interactive Setup**: Simple copy-paste authentication without complex callbacks

## Quick Start

### 1. Setup

1. **Create Google Cloud Project**:
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project or select existing one
   - Enable the Gmail API

2. **Create OAuth2 Credentials**:
   - Go to "Credentials" → "Create Credentials" → "OAuth 2.0 Client IDs"
   - Choose "Desktop Application" type
   - Download the credentials JSON file as `credentials.json`

3. **Install Gmail Mover**:
   ```bash
   # Clone and build
   git clone <repository-url>
   cd gmail-mover
   go build -o gmail-mover ./cmd/
   ```

### 2. Basic Usage

**List available labels for an account:**
```bash
./gmail-mover -list-labels -src=your-email@gmail.com
```

**Transfer messages (dry-run first!):**
```bash
./gmail-mover -src=source@gmail.com -dst=destination@gmail.com -dry-run
```

**Transfer with filtering:**
```bash
./gmail-mover -src=source@gmail.com -dst=destination@gmail.com \
  -src-label=INBOX -query="from:newsletter" -max=50 -dry-run
```

**Actual transfer (remove -dry-run when ready):**
```bash
./gmail-mover -src=source@gmail.com -dst=destination@gmail.com \
  -src-label=INBOX -query="from:newsletter" -max=50
```

## Command-Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-src` | Source Gmail address | *required* |
| `-dst` | Destination Gmail address | *required* |  
| `-src-label` | Source Gmail label to process | `INBOX` |
| `-dst-label` | Label to apply to moved messages | *(none)* |
| `-query` | Gmail search query | *(none)* |
| `-max` | Maximum messages to process | `10000` |
| `-dry-run` | Preview mode - don't actually move | `false` |
| `-delete` | Delete from source after move | `true` |
| `-job` | Load settings from JSON job file | *(none)* |
| `-list-labels` | List available labels for source account | `false` |

## Advanced Usage

### Gmail Query Examples

The `-query` flag supports Gmail's full search syntax:

```bash
# Messages from specific sender
-query="from:noreply@example.com"

# Messages with attachments
-query="has:attachment"

# Messages in date range
-query="after:2023/1/1 before:2023/12/31"

# Unread messages
-query="is:unread"

# Complex queries
-query="from:newsletter has:attachment -is:important"
```

### Job Configuration Files

For complex or repeated operations, use JSON job files:

```json
{
  "name": "Weekly Newsletter Archive",
  "src_account": {
    "email": "personal@gmail.com",
    "labels": ["INBOX"],
    "query": "from:newsletter older_than:7d",
    "max_messages": 1000
  },
  "dst_account": {
    "email": "archive@gmail.com",
    "apply_label": "newsletters-archived",
    "create_label_if_missing": true
  },
  "options": {
    "dry_run": false,
    "delete_after_move": true,
    "fail_on_error": false
  }
}
```

Run with: `./gmail-mover -job=newsletter-archive.json`

## Authentication

On first use with each email account, you'll be prompted to:

1. Visit a Google authorization URL
2. Grant permissions to Gmail Mover
3. Copy the authorization code back to the terminal

Tokens are saved in the `tokens/` directory for future use.

## Safety Features

- **Dry-run by default**: Always test with `-dry-run` first
- **Authentication per account**: Each email requires separate OAuth approval
- **Detailed logging**: See exactly what's happening during transfers
- **Error handling**: Graceful handling of API limits and network issues

## Common Use Cases

**Archive old emails:**
```bash
./gmail-mover -src=main@gmail.com -dst=archive@gmail.com \
  -query="older_than:1y" -dst-label="archived-2023" -dry-run
```

**Organize newsletters:**
```bash
./gmail-mover -src=personal@gmail.com -dst=personal@gmail.com \
  -query="from:newsletter" -dst-label="Newsletters" -delete=false -dry-run
```

**Backup important emails:**
```bash
./gmail-mover -src=work@company.com -dst=personal@gmail.com \
  -query="is:important" -dst-label="work-backup" -dry-run
```

## Troubleshooting

**"No credentials found" error:**
- Ensure `credentials.json` is in the current directory
- Verify the file contains valid OAuth2 credentials from Google Cloud Console

**"Authentication required" error:**
- Run the command and follow the OAuth flow
- Check that tokens are being saved in `tokens/` directory

**Rate limiting:**
- Gmail API has rate limits; the tool will handle retries automatically
- For large transfers, consider using smaller `-max` values

**Permission errors:**
- Ensure OAuth2 credentials have Gmail API scope enabled
- Re-authenticate if permissions were changed

## Security

- OAuth2 tokens are stored locally in `tokens/` directory
- No passwords or sensitive data are transmitted or stored
- Each account requires explicit authorization
- Tokens can be revoked from your Google Account settings

## Building from Source

```bash
# Install Go 1.21 or later
go version

# Clone repository
git clone <repository-url>
cd gmail-mover

# Install dependencies
go mod tidy

# Build
go build -o gmail-mover ./cmd/

# Run tests
go test ./test/ -v
```

## Support

This tool is designed for power users comfortable with command-line interfaces and Gmail's query syntax. For best results:

- Start with small test transfers using `-dry-run`
- Understand Gmail's label system and search operators
- Keep backups of important emails
- Monitor Gmail storage quotas on both accounts

---

**Note**: This tool moves emails between Gmail accounts. Always test thoroughly with `-dry-run` before performing actual transfers.