# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Status:** Active development - Core functionality implemented and working

Gmail Mover is a Go CLI tool that transfers email messages between Gmail accounts using OAuth2 authentication and the Gmail API. It supports interactive message approval, automatic labeling, job configuration, CLI flags, label filtering, Gmail queries, dry-run mode, and advanced message processing features.

## Architecture

The codebase follows a clean multi-package structure with unique, branded package names:

- **cmd package**: CLI entry point with CLI-friendly slog logger and application orchestration
- **gmover package**: Core business logic with domain types and configuration handling
- **gapi package**: Gmail API operations including authentication, message transfer, label management, and query building with pagination
- **gmcmds package**: Command handlers implementing the command pattern
- **gmcfg package**: Centralized configuration file management with filesystem sandboxing
- **cliutil package**: CLI utilities, command registration, and terminal handling

### Key Design Patterns

- **Clear Path Style**: All functions use `goto end` with single exit point, named return variables
- **Unique Package Naming**: Follows branded naming (gmover, gapi, gmcfg) to avoid ecosystem conflicts
- **Constructor Pattern**: All types have New*() constructors with proper defaults
- **Domain Types**: EmailAddress, LabelName, etc. with validation and zero-value checking
- **CLI-Friendly Logging**: Custom slog.Handler in cmd package for clean user output without timestamps
- **Command Pattern**: Extensible command system with registration and validation
- **Interactive Approval**: Single-character input with terminal raw mode and graceful fallbacks
- **Automatic Labeling**: All moved messages get [Gmoved] label for safety
- **Date Preservation**: Original email dates are preserved in Gmail interface using RFC2822 parsing
- **XDG-Compliant Storage**: Credentials and tokens stored in `~/.config/gmover/`
- **Email-Based Token Storage**: Tokens stored per email address with automatic refresh
- **Single Credentials File**: Uses one `credentials.json` for all accounts with guided setup
- **Query Building**: Dynamic Gmail query construction with label, date, and custom filters
- **Pagination Support**: Handles Gmail API pagination to process >500 messages
- **Dry-Run Support**: Preview mode without actually moving messages
- **Interactive Auth**: Copy-paste OAuth flow without callback servers
- **Guided Credentials Setup**: Interactive flow for OAuth2 setup instead of manual file placement
- **Terminal Error Detection**: Proper handling of non-TTY environments (like IDE consoles)
- **OAuth Token Refresh**: Automatic token refresh with persistence
- **Atomic Message Operations**: Ctrl-C protection ensures no partial message transfers

## Development Commands

**IMPORTANT: Always build executables to `./bin/` directory, never to project root!**

```bash
# Install dependencies
go mod tidy

# Build binary (ALWAYS build to ./bin/ directory)
go build -o bin/gmover ./cmd/

# Show help (default behavior)
./bin/gmover

# Run with job file (recommended)
./bin/gmover job run my-job.json --dry-run

# List available labels for an account
./bin/gmover list labels --src=user@gmail.com

# Create a new job interactively
./bin/gmover job define move-emails

# Run tests
go test ./test/ -v
```

## Authentication Setup

The application now provides **guided setup** - no manual credential file placement required:

1. **Google Cloud Console project** with Gmail API enabled
2. **OAuth 2.0 Client ID** (Desktop Application type)
3. **Run the application** - it will guide you through credential setup
4. **Paste credentials JSON** when prompted during first run

**Important**: Now uses `gmail.MailGoogleComScope` for full Gmail access including delete permissions.

Token files are auto-generated in `~/.config/gmover/tokens/` directory using format `token-{email}.json` after first authorization for each email address. Tokens automatically refresh and persist.

## CLI Usage Patterns

### Show Help (Default)
```bash
./bin/gmover
```

### List Labels
```bash
./bin/gmover list labels --src=user@gmail.com
```

### Create and Run Job (Recommended)
```bash
# Create job interactively
./bin/gmover job define move-newsletters

# Run job with approval prompts
./bin/gmover job run move-newsletters.json

# Run job in dry-run mode
./bin/gmover job run move-newsletters.json --dry-run
```

### Direct Move (Legacy)
```bash
./bin/gmover move --src=user@gmail.com --dst=archive@gmail.com --src-label=INBOX --dst-label=archived --max=100 --dry-run
```

### Job Configuration File Example
```json
{
  "name": "Archive Newsletter",
  "src_account": {
    "email": "user@gmail.com",
    "labels": ["INBOX"],
    "query": "from:newsletter",
    "max_messages": 100
  },
  "dst_account": {
    "email": "archive@gmail.com",
    "apply_label": "archived-newsletters"
  },
  "options": {
    "dry_run": false,
    "delete_after_move": false
  }
}
```

## Interactive Features

### Message Approval
- **Single character input**: Press Y/N/A/C without Enter (in real terminals)
- **Graceful fallback**: Line input with Enter in non-TTY environments (IDE consoles)
- **Ctrl-C handling**: Proper cancellation with atomic message operations
- **Options**:
  - `Y` - Approve this message
  - `N` - Skip this message  
  - `A` - Approve all remaining messages
  - `C` - Cancel entire operation
  - `Ctrl-C` - Cancel entire operation

### Automatic Safety Features
- **[Gmoved] label**: Automatically applied to all moved messages
- **Label creation**: Missing labels are created automatically
- **Atomic operations**: Ctrl-C waits for current message to complete
- **Dry-run mode**: Preview operations safely

## Code Patterns

### Clear Path Style Functions
```go
// All functions follow this pattern
func GetGmailService(email string) (service *gmail.Service, err error) {
    var config *oauth2.Config
    var token *oauth2.Token
    
    config, err = loadCredentials()
    if err != nil {
        goto end
    }
    
    token, err = getToken(config, email)
    if err != nil {
        goto end
    }
    
    // ... more logic
    
end:
    return service, err
}
```

### Domain Types with Validation
```go
// Use domain types instead of strings
var email gmover.EmailAddress = "user@gmail.com"
var label gmover.LabelName = "INBOX"

// Zero-value checking
if email.IsZero() {
    return fmt.Errorf("email is required")
}
```

### Command Pattern
```go
// Commands implement CommandHandler interface
type MoveCmd struct {
    *cliutil.CmdBase
}

func (c *MoveCmd) Handle(ctx context.Context, config cliutil.Config, args []string) error {
    // Implementation
}

// Registration in init()
func init() {
    cliutil.RegisterCommand(&MoveCmd{...})
}
```

### Logger Setup (Required)
```go
// Must set logger before using packages
handler := NewCLIHandler()  // Custom CLI-friendly handler
logger := slog.New(handler)
gmover.Initialize(&gmover.Opts{Logger: logger})
```

### File Store Usage
```go
// Centralized config file management
store := gmcfg.NewFileStore("gmover")
err := store.Save("config.json", configData)
err = store.Load("config.json", &configData)
exists := store.Exists("config.json")
```

## Implemented Features

- **Command System**: Extensible command pattern with job management
- **Interactive Approval**: Single-character input with terminal fallbacks
- **Automatic Safety**: [Gmoved] labels and atomic operations
- **OAuth2 Flow**: Guided credential setup with automatic token refresh
- **Pagination**: Handle >500 messages via Gmail API pagination
- **Label Management**: Automatic label creation and application
- **Dry-Run Mode**: Safe preview of operations
- **Job Configuration**: JSON-based job files with validation
- **Terminal Handling**: Raw mode with graceful fallbacks for IDE consoles
- **Error Recovery**: Proper error detection and user-friendly messaging
- **Filesystem Sandboxing**: Secure config file access with `io/fs.Sub`

## Dependencies

- `golang.org/x/oauth2`: OAuth2 authentication  
- `golang.org/x/term`: Terminal control for single-character input
- `google.golang.org/api/gmail/v1`: Gmail API client (isolated in gapi package)
- Go standard library: Enhanced with domain types and clear patterns
- No external CLI frameworks - keeps dependencies minimal

## Coding Standards

This codebase follows Clear Path style guidelines:
- **Named return variables** in all function signatures
- **Single exit point** using `goto end` pattern
- **Variable pre-declaration** before any `goto` statements
- **No early returns** - all functions exit through `end:` label
- **Unique package names** to avoid ecosystem conflicts
- **Constructor functions** for all types (New*() pattern)
- **No `else` statements** - use helper functions or `goto end` instead
- **Domain types** instead of primitive strings for type safety
- **Comments describe code state, not development process** - Never reference conversations, chat sessions, or implementation changes in code comments

## Build Standards

- **ALWAYS build executables to `./bin/` directory**: `go build -o bin/gmover ./cmd/`
- **NEVER build to project root** - keeps the root directory clean
- **Use `./bin/gmover` for all testing and examples**

## Testing Notes

Integration tests in `test/` directory:
- Tests verify command system, job creation, and configuration
- Tests use proper logger setup and Clear Path style
- Most tests gracefully handle auth errors (expected without credentials)
- Use `--dry-run` flag for safe testing without actual Gmail operations

Manual testing requires:
- Gmail accounts with API access
- OAuth2 credentials (guided setup on first run)
- Test messages in source account for specified labels

## Recent Major Changes

### Completed
- ✅ **Command System**: Implemented extensible command pattern
- ✅ **Interactive Approval**: Single-character input with Ctrl-C handling
- ✅ **Automatic Labels**: [Gmoved] label added to all moved messages
- ✅ **OAuth Improvements**: Full scope access and automatic token refresh
- ✅ **Pagination**: Gmail API pagination for >500 messages
- ✅ **Label Creation**: Automatic creation of missing labels
- ✅ **Terminal Handling**: Raw mode with IDE console fallbacks
- ✅ **Guided Setup**: Interactive OAuth2 credential setup
- ✅ **File Store**: Centralized config management with security
- ✅ **Atomic Operations**: Ctrl-C protection for message integrity

### Current Bugs/TODOs
- [ ] Fix log level bug in gapi/transfer.go (Error should be Info)
- [ ] Fix MaxMessages bug in gmover/src_account.go:30 (not being set)
- [ ] Fix flag logic bug in cmd/main.go listLabels check
- [ ] Move signal handling from cmd/main.go to gmover.Initialize()
- [ ] Add comprehensive integration tests for new command system
- [ ] Consider making MaxMessages part of TransferOpts instead of global variable

## Security Notes

- **OAuth2 Scopes**: Uses `gmail.MailGoogleComScope` for full Gmail access
- **File Access**: Sandboxed to config directory using `io/fs.Sub`
- **Token Storage**: Secure per-email token files with automatic refresh
- **Input Validation**: Domain types prevent invalid email/label values
- **Error Handling**: No sensitive data in error messages