# Gmail Mover

A fast, reliable command-line tool for transferring Gmail messages between accounts with interactive approval, automatic safety labeling, and atomic operations.

## Features

- **Interactive Message Approval**: Review each message before moving with single-character responses (y/n/a/d/c)
- **Automatic Safety Labeling**: All moved messages get `[Gmoved]` label for easy tracking and recovery
- **Date Preservation**: Original email dates are preserved in Gmail interface, not transfer date
- **Multi-Account Support**: Transfer messages between different Gmail accounts
- **Command System**: Extensible command pattern with guided OAuth setup and token management
- **Terminal Handling**: Raw mode input with graceful fallbacks for IDE consoles
- **Advanced Filtering**: Use Gmail's query syntax, labels, and date ranges
- **Dry-Run Mode**: Preview what will be moved before making changes
- **Pagination Support**: Handle transfers of >500 messages efficiently
- **Atomic Operations**: Ctrl-C protection prevents orphaned messages
- **OAuth2 Security**: Secure authentication with guided setup and automatic token refresh
- **XDG-Compliant Storage**: Credentials and tokens stored in `~/.config/gmover/`

## Quick Start

### 1. Setup

1. **Create Google Cloud Project**:
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project or select existing one
   - Enable the Gmail API

2. **Create OAuth2 Credentials**:
   - Go to "Credentials" → "Create Credentials" → "OAuth 2.0 Client IDs"
   - Choose "Desktop Application" type
   - Download the credentials JSON file
   - Save it as `~/.config/gmail-mover/credentials.json`

3. **Install Gmail Mover**:
   ```bash
   # Clone and build (ALWAYS build to ./bin/ directory)
   git clone <repository-url>
   cd gmail-mover
   go build -o bin/gmover ./cmd/gmover-cli/
   ```

### 2. Basic Usage

**Get help (default behavior):**
```bash
./bin/gmover
```

**First-time OAuth setup:**
On first use, you'll be guided through the OAuth setup automatically when you try to access Gmail.

**List available labels for an account:**
```bash
./bin/gmover list --src=your-email@gmail.com
```

**Interactive transfer with approval (recommended):**
```bash
./bin/gmover move -src=source@gmail.com -dst=destination@gmail.com \
  -src-label=INBOX -query="from:newsletter" -max=50 -dry-run
```

**Transfer with automatic approval:**
```bash
./bin/gmover move -src=source@gmail.com -dst=destination@gmail.com \
  -src-label=INBOX -query="from:newsletter" -max=50 --auto-approve
```

## Command-Line Options

### Commands
Gmail Mover uses a command-based interface:

| Command | Usage | Description |
|---------|-------|-------------|
| **help** | `./bin/gmover` or `./bin/gmover help` | Shows usage information and examples |
| **list** | `./bin/gmover list --src=EMAIL` | Lists available labels for the specified account |
| **move** | `./bin/gmover move -src=EMAIL -dst=EMAIL` | Interactive email transfer with approval |
| **job** | `./bin/gmover job define` or `./bin/gmover job run` | Create or execute job files |

### Common Options

| Flag | Description | Default |
|------|-------------|---------|
| `--auto-approve` | Skip interactive approval (approve all) | `false` |
| `-src` | Source Gmail address | *required for most operations* |
| `-dst` | Destination Gmail address | *required for moves* |  
| `-src-label` | Source Gmail label to process | `INBOX` |
| `-dst-label` | Label to apply to moved messages | *(none)* |
| `-query` | Gmail search query | *(none)* |
| `-max` | Maximum messages to process | `10000` |
| `-dry-run` | Preview mode - don't actually move | `false` |
| `-delete` | Delete from source after move | `false` |
| `-job` | Load settings from JSON job file | *(none)* |

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

Run with: `./bin/gmover job run newsletter-archive.json`

## Authentication

### Automatic Setup

Gmail Mover provides guided OAuth configuration on first use. When you first run a command that requires Gmail access, it will:

1. **Prompt for Credentials**: Guide you through pasting your OAuth2 credentials JSON
2. **Account Authorization**: Automatically open browser for account authorization  
3. **Token Management**: Handle token refresh and storage automatically

### Authorization Process

On first use with each email account, you'll be prompted to:

1. Visit a Google authorization URL
2. Grant permissions to Gmail Mover
3. Copy the authorization code back to the terminal

Tokens are automatically saved in `~/.config/gmover/tokens/` for future use with automatic refresh.

## Safety Features

- **Interactive Approval**: Review each message individually before moving
- **Automatic `[Gmoved]` Labeling**: All moved messages are tagged for easy tracking
- **Atomic Operations**: Ctrl-C protection prevents partial transfers
- **Help by default**: Shows help information when run without arguments
- **Command-based Interface**: Clear separation between different operations
- **Dry-run support**: Always test with `-dry-run` first
- **Authentication per account**: Each email requires separate OAuth approval
- **Detailed logging**: See exactly what's happening during transfers
- **Terminal Handling**: Graceful fallbacks for different console environments
- **Error handling**: Graceful handling of API limits and network issues

## Common Use Cases

**Archive old emails with interactive approval:**
```bash
./bin/gmover move -src=main@gmail.com -dst=archive@gmail.com \
  -query="older_than:1y" -dst-label="archived-2023" -dry-run
```

**Organize newsletters automatically:**
```bash
./bin/gmover move -src=personal@gmail.com -dst=personal@gmail.com \
  -query="from:newsletter" -dst-label="Newsletters" -delete=false --auto-approve
```

**Backup important emails with approval:**
```bash
./bin/gmover move -src=work@company.com -dst=personal@gmail.com \
  -query="is:important" -dst-label="work-backup" -dry-run
```

**Interactive Approval Responses:**
- `y` - Yes, move this message
- `n` - No, skip this message  
- `a` - Yes to all remaining messages
- `d` - Delay 3 seconds between messages
- `c` - Cancel operation

## Troubleshooting

**"No credentials found" error:**
- Run any command that accesses Gmail to trigger guided credential setup
- Paste your OAuth2 credentials JSON when prompted
- Verify the file contains valid OAuth2 credentials from Google Cloud Console  
- The directory `~/.config/gmover/` will be created automatically

**"Authentication required" error:**
- Run the command and follow the guided OAuth flow
- Check that tokens are being saved in `~/.config/gmover/tokens/` directory

**Rate limiting:**
- Gmail API has rate limits; the tool will handle retries automatically
- For large transfers, consider using smaller `-max` values

**Permission errors:**
- Ensure OAuth2 credentials have Gmail API scope enabled
- Re-authenticate if permissions were changed

## Security

- OAuth2 tokens are stored locally in `~/.config/gmover/tokens/` directory (XDG-compliant)
- Credentials are stored in `~/.config/gmover/credentials.json` (user config area)
- Automatic token refresh prevents expired authentication
- `[Gmoved]` labels provide audit trail of all moved messages
- No passwords or sensitive data are transmitted or stored in the project
- Each account requires explicit authorization
- Tokens can be revoked from your Google Account settings
- Filesystem sandboxing restricts access to configuration directory

## Building from Source

```bash
# Install Go 1.21 or later
go version

# Clone repository
git clone <repository-url>
cd gmail-mover

# Install dependencies
go mod tidy

# Build (ALWAYS build to ./bin/ directory)
go build -o bin/gmover ./cmd/gmover-cli/

# Run tests
go test ./test/ -v

# Test with dry-run
./bin/gmover move -src=test@gmail.com -dst=test@gmail.com -dry-run
```

## Support

This tool is designed for power users comfortable with command-line interfaces and Gmail's query syntax. For best results:

- Start with small test transfers using `-dry-run`
- Understand Gmail's label system and search operators
- Keep backups of important emails
- Monitor Gmail storage quotas on both accounts

---

**Note**: This tool moves emails between Gmail accounts with interactive approval and automatic safety labeling. All moved messages receive the `[Gmoved]` label for tracking. Always test thoroughly with `-dry-run` before performing actual transfers.