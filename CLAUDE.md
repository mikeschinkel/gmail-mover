# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Status:** Active development - Core functionality implemented

Gmail Mover is a Go CLI tool that transfers email messages between Gmail accounts using OAuth2 authentication and the Gmail API. It supports flexible job configuration, CLI flags, label filtering, Gmail queries, dry-run mode, and advanced message processing features.

## Architecture

The codebase follows a clean 3-package structure with unique, branded package names:

- **cmd package**: CLI entry point with flag parsing, CLI-friendly slog logger, and application orchestration
- **gmover package**: Core business logic, job management, configuration handling with private fields and getter/setter methods
- **gapi package**: Gmail API operations including authentication, message transfer, label management, and query building

### Key Design Patterns

- **Clear Path Style**: All functions use `goto end` with single exit point, named return variables
- **Unique Package Naming**: Follows branded naming (gmover, gapi) to avoid ecosystem conflicts
- **Constructor Pattern**: All types have New*() constructors with proper defaults
- **Encapsulated Configuration**: Config struct with private fields, public getter/setter methods
- **CLI-Friendly Logging**: Custom slog.Handler in cmd package for clean user output without timestamps
- **Required Logger Pattern**: gmover and gapi packages require logger setup with panic protection
- **Job-Based Configuration**: Flexible job system supporting both CLI flags and JSON config files
- **XDG-Compliant Storage**: Credentials and tokens stored in `~/.config/gmail-mover/`
- **Email-Based Token Storage**: Tokens stored per email address in user config directory
- **Single Credentials File**: Uses one `credentials.json` for all accounts
- **Query Building**: Dynamic Gmail query construction with label, date, and custom filters
- **Dry-Run Support**: Preview mode without actually moving messages
- **Interactive Auth**: Copy-paste OAuth flow without callback servers
- **Safe Default Behavior**: Defaults to ShowHelp mode instead of destructive operations
- **Explicit Mode Switching**: Requires `-dst` or `-job` flags to enable move operations
- **Configuration Validation**: Early validation of required fields before execution

## Development Commands

**IMPORTANT: Always build executables to `./bin/` directory, never to project root!**

```bash
# Install dependencies
go mod tidy

# Build binary (ALWAYS build to ./bin/ directory)
go build -o bin/gmover ./cmd/

# Show help (default behavior)
./bin/gmover

# Run with CLI flags
./bin/gmover -src=user@gmail.com -dst=archive@gmail.com -max=50 -dry-run

# Run with job file
./bin/gmover -job=config.json

# List available labels for an account
./bin/gmover -list-labels -src=user@gmail.com

# For debugging OAuth or API issues
./bin/gmover -src=user@gmail.com -dst=archive@gmail.com -dry-run 2>&1 | tee transfer.log

# Run tests
go test ./test/ -v
```

## Authentication Setup

Before first run, you need:

1. Google Cloud Console project with Gmail API enabled
2. OAuth 2.0 Client ID (Desktop Application type)
3. Downloaded credentials placed in: `~/.config/gmail-mover/credentials.json`

Token files are auto-generated in `~/.config/gmail-mover/tokens/` directory using format `{email}_token.json` after first authorization for each email address.

## CLI Usage Patterns

### Show Help (Default)
```bash
./bin/gmover
```

### List Labels
```bash
./bin/gmover -list-labels -src=user@gmail.com
```

### Basic Transfer
```bash
./bin/gmover -src=user@gmail.com -dst=archive@gmail.com -max=100
```

### Advanced Filtering  
```bash
./bin/gmover -src=user@gmail.com -dst=archive@gmail.com \
  -src-label=INBOX -query="from:newsletter" -max=50 -dry-run
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
    "apply_label": "archived-newsletters",
    "create_label_if_missing": true
  },
  "options": {
    "dry_run": false,
    "delete_after_move": false,
    "fail_on_error": false,
    "log_level": "info"
  }
}
```

### Code Patterns

#### Clear Path Style Functions
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

#### Constructor Usage
```go
// All types have constructors
job, err := gmover.NewJob(gmover.JobOptions{
    Name:     "My Job",
    SrcEmail: "user@gmail.com",
    DstEmail: "archive@gmail.com",
    // ... other options
})
```

#### Configuration API
```go
// Config with private fields and public methods
config := gmover.NewConfig(gmover.MoveEmails)
config.SetSrcEmail("user@gmail.com")
config.SetDryRun(true)

// Access values through getters
email := config.SrcEmail()
isDryRun := config.DryRun()
```

#### Logger Setup (Required)
```go
// Must set logger before using gmover/gapi packages
handler := NewCLIHandler()  // Custom CLI-friendly handler
logger := slog.New(handler)
gmover.SetLogger(logger)
gapi.SetLogger(logger)
```

#### Job Execution Flow
1. Parse CLI flags using standard flag package  
2. Create config with gmover.NewConfig(gmover.ShowHelp) (default mode)
3. Determine run mode based on flags:
   - `-list-labels` → ListLabels mode
   - `-job FILE` → MoveEmails mode  
   - `-dst EMAIL` → MoveEmails mode
   - No special flags → ShowHelp mode (default)
4. Set config values via setters and pass config pointer to gmover.Run(&config)
5. Configuration validation occurs before execution
6. Job creation and execution handled internally based on mode

#### Error Handling
- Clear Path style with single exit point via `goto end`
- Configurable continue-on-error behavior
- Graceful handling for individual message failures
- Named return variables for consistent error propagation

## Implemented Features

- **Three Operation Modes**: ShowHelp (default), ListLabels, and MoveEmails
- **Safe Default Behavior**: Help mode prevents accidental destructive operations
- **Configuration Validation**: Early validation with helpful error messages  
- **CLI Flags**: Full support for source/dest emails, labels, queries, limits
- **Job Configuration**: JSON-based job files for complex configurations
- **Filtering**: Label-based, Gmail query syntax, date range filtering
- **Dry-Run Mode**: Preview operations without actually moving messages
- **Label Management**: Apply labels to moved messages, list available labels
- **Delete After Move**: Optional deletion from source after successful transfer
- **Error Handling**: Configurable continue-on-error behavior

## Planned Features (Future)

- SQLite logging for deduplication
- Label creation if missing
- Batch operations for improved performance
- More advanced date filtering options

## Dependencies

- `golang.org/x/oauth2`: OAuth2 authentication  
- `google.golang.org/api/gmail/v1`: Gmail API client (isolated in gapi package)
- Go standard library: `flag` for CLI parsing, `log/slog` for logging
- No external CLI frameworks - keeps dependencies minimal
- Follows single major third-party feature package per package rule

## Coding Standards

This codebase follows Clear Path style guidelines:
- **Named return variables** in all function signatures
- **Single exit point** using `goto end` pattern
- **Variable pre-declaration** before any `goto` statements
- **No early returns** - all functions exit through `end:` label
- **Unique package names** to avoid ecosystem conflicts
- **Constructor functions** for all types (New*() pattern)
- **No `else` statements** - use helper functions or `goto end` instead

## Build Standards

- **ALWAYS build executables to `./bin/` directory**: `go build -o bin/gmover ./cmd/`
- **NEVER build to project root** - keeps the root directory clean
- **Use `./bin/gmover` for all testing and examples**

## Testing Notes

Integration tests exist in `test/` directory:
- Tests use `setupTestLogger()` helper to initialize required loggers
- Tests verify Config getter/setter methods, job creation, and run logic
- Most tests fail gracefully on auth errors (expected without credentials)
- Use `-dry-run` flag for safe testing without actual Gmail operations

Manual testing requires:
- Gmail accounts with API access
- Valid OAuth2 credentials in `credentials.json` 
- Test messages in source account for specified labels

When adding tests, ensure they follow Clear Path style and call `setupTestLogger()` before using gmover/gapi packages.

## TODO List

### High Priority
- [ ] Fix Job.Execute() to pass full configuration to gapi.TransferMessagesWithOpts (architectural concern #1)
- [ ] Fix flag logic bug in cmd/main.go listLabels check
- [ ] Fix log level bug in gapi/transfer.go:133 (Error should be Info)  
- [ ] Fix MaxMessages not being set in gmover/src_account.go:30

### Medium Priority
- [ ] Add comprehensive integration tests for gmover/gapi interaction

### Implementation Details

#### applyLabels Function (gapi/labels.go:50)
The `applyLabels` function currently contains `panic("IMPLEMENT ME")` and needs proper implementation:

**Requirements:**
1. Check if the specified label exists in the destination Gmail account
2. Create the label if it doesn't exist (based on job config `CreateLabelIfMissing` flag)
3. Apply the label(s) to the specified message using Gmail API
4. Handle errors gracefully (don't panic on missing labels if creation is disabled)

**Gmail API Reference:**
- `service.Users.Labels.List()` - to check existing labels
- `service.Users.Labels.Create()` - to create new labels if needed  
- `service.Users.Messages.Modify()` - to apply labels to messages

**Current Usage:**
Called from `gapi/transfer.go:119` during message transfer when `opts.LabelsToApply` is not empty.

**Error Handling:**
Should follow Clear Path style with `goto end` pattern and return meaningful errors for debugging.

### Low Priority / Future
- [ ] Consider making MaxMessages part of TransferOpts instead of global variable
- SQLite logging for deduplication
- Label creation if missing
- Batch operations for improved performance
- More advanced date filtering options