# ADR-002: OAuth Configuration and Token Management Strategy

**Status:** Accepted  
**Date:** 2025-08-08

## Context

Gmail Mover requires OAuth 2.0 authentication to access Gmail accounts. The application must handle credentials, tokens, and multi-account access in a user-friendly and secure manner.

## Decisions

### **Single OAuth Application for Multi-Account Access**
- **Decision**: Use one OAuth 2.0 Client ID that can authenticate any Gmail account dynamically
- **Implementation**: Single `credentials.json` file supports authentication for any Gmail account
- **Rationale**: Eliminates need for users to create Google Cloud projects per account, provides better UX

### **XDG-Compliant Configuration Directory**
- **Decision**: Store all configuration and authentication data in `~/.config/gmover/`
- **Implementation**: Directory structure follows XDG Base Directory specification
- **Rationale**: Standard location users expect, separate from application code, respects user preferences

### **Configuration Directory Structure**
```
~/.config/gmover/
├── credentials.json          # Single OAuth client credentials for all accounts
├── tokens/                   # Per-account token storage
│   ├── token-user1@gmail.com.json
│   ├── token-user2@gmail.com.json
│   └── ...
└── [future config files]
```

### **Per-Account Token Isolation**
- **Decision**: Store OAuth tokens in separate files per Gmail account using format `token-{email}.json`
- **Implementation**: Token files isolated in `tokens/` subdirectory
- **Rationale**: Prevents token conflicts, enables account-specific token refresh, supports concurrent access

### **Copy-Paste OAuth Flow**
- **Decision**: Use OAuth flow that displays URL for user to copy-paste rather than callback server
- **Implementation**: Interactive authorization with device flow pattern
- **Rationale**: Simpler deployment (no callback URL), works in server environments, avoids port conflicts

### **Automatic Token Refresh**
- **Decision**: Automatically refresh OAuth tokens when expired and persist updated tokens
- **Implementation**: Token refresh handled transparently during API calls
- **Rationale**: Seamless user experience, reduces authentication friction for repeated use

### **Guided Credentials Setup**
- **Decision**: Provide interactive flow to help users set up OAuth credentials rather than manual file placement
- **Implementation**: Application guides users through credential acquisition and setup
- **Rationale**: Reduces setup complexity, provides better error messages, improves first-time user experience

## Implementation Evidence

This strategy is implemented in:
- OAuth flow in `gapi/auth.go`
- Configuration management in `gmcfg/file_store.go`
- Token handling throughout `gapi/` package
- Directory structure created by `gmcfg.NewFileStore("gmover")`

## Consequences

**Positive:**
- Simple setup process for users (one credentials file for all accounts)
- Secure token isolation prevents cross-account token conflicts
- Standard configuration directory location
- Works in various deployment scenarios (local, server, containerized)
- Automatic token refresh reduces manual intervention

**Trade-offs:**
- Users must have Google Cloud Console access to create OAuth credentials
- Copy-paste flow requires manual browser interaction
- Local token storage means tokens don't sync across devices