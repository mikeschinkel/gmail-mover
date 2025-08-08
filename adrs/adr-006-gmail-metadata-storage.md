# ADR-006: Google Apps Script Execution API for Gmail Mover Metadata Storage

**Date:** 2025-08-06  
**Status:** Accepted  
**Supersedes:** Initial consideration of multiple metadata storage approaches

## Context

The Gmail Mover project is envisioned as a comprehensive 360-degree personal information manager, with features to include transferring email messages between Gmail accounts, email archiving with SQLite storage, AI-assisted classification, intelligent automation, and future expansion to other messaging platforms (Slack, Teams, social media), documents (Google Docs, Office files), browser data, financial transactions, and more. The system requires persistent storage of operational metadata that includes:
# ADR-006: Google Apps Script Execution API for Gmail Mover Metadata Storage

**Date:** 2025-08-06  
**Status:** Accepted  
**Supersedes:** Initial consideration of multiple metadata storage approaches

## Context

The Gmail Mover project has features to transfer email messages between Gmail accounts with advanced filtering, AI-assisted classification, and intelligent automation. The system requires persistent storage of operational metadata that includes:

### Core Requirements
- **Per-user, per-account state tracking**: Sync cursors, last processed message IDs, account-specific settings
- **Minimal operational state**: Only metadata essential for accurate future processing operations
- **Configuration management**: User preferences specific to Google account integration
- **Cross-session persistence**: State must survive application restarts and deployments
- **Account association**: Metadata must remain tied to the specific Google account it serves

### Technical Constraints
- **Invisibility requirement**: Metadata must not appear in user's Gmail/Drive interfaces
- **Concurrent access safety**: Multiple application instances may access the same user data
- **OAuth integration**: Should leverage existing Gmail API authentication where possible
- **Scalability**: Solution must work for both individual users and hosted SaaS deployment
- **Data sovereignty**: Users should retain control over their metadata

### Architectural Context
The Gmail Mover ecosystem consists of:
1. **Core CLI tool** (Go) - transfers messages between accounts
2. **Future web interface** - dashboard for email analytics and rule management
3. **SQLite-based archive** - long-term email storage with full-text search
4. **AI classification engine** - learns user patterns and generates rules
5. **Metadata storage layer** - operational state and configuration

## Decision

We will implement metadata storage using **Google Apps Script with PropertiesService and LockService**, accessed via the Apps Script Execution API from our Go application.

### Architecture Components

#### 1. Apps Script Project Structure
```javascript
// Code.gs - Main metadata management functions

/**
 * Stores JSON metadata with automatic locking and error handling
 * @param {string} key - Metadata key (e.g., 'sync_state', 'user_config')
 * @param {string} jsonData - JSON string to store
 * @returns {Object} Success/failure status with optional error details
 */
function storeMetadata(key, jsonData) {
  var lock = LockService.getUserLock();
  try {
    if (!lock.waitLock(30000)) { // 30 second timeout
      return { 
        success: false, 
        error: 'LOCK_TIMEOUT',
        message: 'Could not acquire lock within 30 seconds' 
      };
    }
    
    // Validate JSON before storing
    try {
      JSON.parse(jsonData);
    } catch (e) {
      return { 
        success: false, 
        error: 'INVALID_JSON',
        message: 'Invalid JSON provided: ' + e.message 
      };
    }
    
    PropertiesService.getUserProperties().setProperty(key, jsonData);
    return { 
      success: true, 
      timestamp: new Date().toISOString(),
      key: key,
      dataLength: jsonData.length
    };
    
  } catch (error) {
    return { 
      success: false, 
      error: 'STORAGE_ERROR',
      message: error.toString() 
    };
  } finally {
    lock.releaseLock();
  }
}

/**
 * Retrieves JSON metadata by key
 * @param {string} key - Metadata key to retrieve
 * @returns {Object} Retrieved data or error information
 */
function fetchMetadata(key) {
  try {
    var data = PropertiesService.getUserProperties().getProperty(key);
    if (data === null) {
      return { 
        success: false, 
        error: 'KEY_NOT_FOUND',
        message: 'No data found for key: ' + key 
      };
    }
    
    var parsed = JSON.parse(data);
    return { 
      success: true, 
      data: parsed,
      key: key,
      retrievedAt: new Date().toISOString()
    };
    
  } catch (error) {
    return { 
      success: false, 
      error: 'RETRIEVAL_ERROR',
      message: error.toString() 
    };
  }
}

/**
 * Lists all metadata keys for debugging and management
 * @returns {Object} Array of keys or error
 */
function listMetadataKeys() {
  try {
    var props = PropertiesService.getUserProperties();
    var keys = Object.keys(props.getProperties());
    return { 
      success: true, 
      keys: keys,
      count: keys.length 
    };
  } catch (error) {
    return { 
      success: false, 
      error: 'LIST_ERROR',
      message: error.toString() 
    };
  }
}

/**
 * Deletes metadata by key with confirmation
 * @param {string} key - Key to delete
 * @param {string} confirmation - Must match key for safety
 * @returns {Object} Deletion result
 */
function deleteMetadata(key, confirmation) {
  if (key !== confirmation) {
    return { 
      success: false, 
      error: 'CONFIRMATION_MISMATCH',
      message: 'Confirmation must match key exactly' 
    };
  }
  
  var lock = LockService.getUserLock();
  try {
    if (!lock.waitLock(10000)) {
      return { 
        success: false, 
        error: 'LOCK_TIMEOUT',
        message: 'Could not acquire lock for deletion' 
      };
    }
    
    PropertiesService.getUserProperties().deleteProperty(key);
    return { 
      success: true, 
      deletedKey: key,
      deletedAt: new Date().toISOString()
    };
    
  } catch (error) {
    return { 
      success: false, 
      error: 'DELETE_ERROR',
      message: error.toString() 
    };
  } finally {
    lock.releaseLock();
  }
}

/**
 * Atomic update operation - read, modify, write with optimistic locking
 * @param {string} key - Key to update
 * @param {string} updateFunction - Function name to call with current data
 * @param {Object} params - Parameters to pass to update function
 * @returns {Object} Update result
 */
function atomicUpdate(key, updateFunction, params) {
  var lock = LockService.getUserLock();
  try {
    if (!lock.waitLock(30000)) {
      return { 
        success: false, 
        error: 'LOCK_TIMEOUT' 
      };
    }
    
    // Read current data
    var currentData = PropertiesService.getUserProperties().getProperty(key);
    var parsed = currentData ? JSON.parse(currentData) : {};
    
    // Apply update function (must be defined in this script)
    var updated;
    switch(updateFunction) {
      case 'incrementCounter':
        updated = incrementCounter(parsed, params);
        break;
      case 'updateLastSync':
        updated = updateLastSync(parsed, params);
        break;
      case 'mergeConfig':
        updated = mergeConfig(parsed, params);
        break;
      default:
        return { 
          success: false, 
          error: 'UNKNOWN_UPDATE_FUNCTION',
          message: 'Update function not recognized: ' + updateFunction 
        };
    }
    
    // Store updated data
    PropertiesService.getUserProperties().setProperty(key, JSON.stringify(updated));
    return { 
      success: true, 
      key: key,
      updatedAt: new Date().toISOString(),
      data: updated
    };
    
  } catch (error) {
    return { 
      success: false, 
      error: 'ATOMIC_UPDATE_ERROR',
      message: error.toString() 
    };
  } finally {
    lock.releaseLock();
  }
}

// Helper functions for atomic updates
function incrementCounter(data, params) {
  data.counters = data.counters || {};
  data.counters[params.counterName] = (data.counters[params.counterName] || 0) + (params.increment || 1);
  return data;
}

function updateLastSync(data, params) {
  data.lastSync = params.timestamp;
  data.syncHistory = data.syncHistory || [];
  data.syncHistory.push({
    timestamp: params.timestamp,
    account: params.account,
    messageCount: params.messageCount
  });
  // Keep only last 50 sync records
  if (data.syncHistory.length > 50) {
    data.syncHistory = data.syncHistory.slice(-50);
  }
  return data;
}

function mergeConfig(data, params) {
  data.config = data.config || {};
  Object.keys(params.updates).forEach(function(key) {
    data.config[key] = params.updates[key];
  });
  data.config.lastModified = new Date().toISOString();
  return data;
}
```

#### 2. Go Client Implementation
```go
// metadata/client.go
package metadata

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    script "google.golang.org/api/script/v1"
)

// Client wraps Apps Script Execution API for metadata operations
type Client struct {
    service  *script.Service
    scriptID string
}

// NewClient creates a metadata client with the provided script service and ID
func NewClient(service *script.Service, scriptID string) *Client {
    return &Client{
        service:  service,
        scriptID: scriptID,
    }
}

// Response represents the standardized response from Apps Script functions
type Response struct {
    Success   bool                   `json:"success"`
    Data      map[string]interface{} `json:"data,omitempty"`
    Error     string                 `json:"error,omitempty"`
    Message   string                 `json:"message,omitempty"`
    Timestamp string                 `json:"timestamp,omitempty"`
}

// Store saves JSON data under the specified key
func (c *Client) Store(ctx context.Context, key string, data interface{}) error {
    jsonData, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("marshal data: %w", err)
    }

    req := &script.ExecutionRequest{
        Function:   "storeMetadata",
        Parameters: []interface{}{key, string(jsonData)},
    }

    resp, err := c.service.Scripts.Run(c.scriptID, req).Context(ctx).Do()
    if err != nil {
        return fmt.Errorf("execute storeMetadata: %w", err)
    }

    if resp.Error != nil {
        return fmt.Errorf("script error: %v", resp.Error)
    }

    var result Response
    if err := parseResult(resp.Response.Result, &result); err != nil {
        return fmt.Errorf("parse response: %w", err)
    }

    if !result.Success {
        return fmt.Errorf("store failed: %s - %s", result.Error, result.Message)
    }

    return nil
}

// Fetch retrieves data by key and unmarshals into the provided struct
func (c *Client) Fetch(ctx context.Context, key string, target interface{}) error {
    req := &script.ExecutionRequest{
        Function:   "fetchMetadata",
        Parameters: []interface{}{key},
    }

    resp, err := c.service.Scripts.Run(c.scriptID, req).Context(ctx).Do()
    if err != nil {
        return fmt.Errorf("execute fetchMetadata: %w", err)
    }

    if resp.Error != nil {
        return fmt.Errorf("script error: %v", resp.Error)
    }

    var result Response
    if err := parseResult(resp.Response.Result, &result); err != nil {
        return fmt.Errorf("parse response: %w", err)
    }

    if !result.Success {
        if result.Error == "KEY_NOT_FOUND" {
            return ErrKeyNotFound
        }
        return fmt.Errorf("fetch failed: %s - %s", result.Error, result.Message)
    }

    // Marshal and unmarshal to convert to target type
    dataBytes, err := json.Marshal(result.Data)
    if err != nil {
        return fmt.Errorf("marshal result data: %w", err)
    }

    if err := json.Unmarshal(dataBytes, target); err != nil {
        return fmt.Errorf("unmarshal to target: %w", err)
    }

    return nil
}

// AtomicUpdate performs an atomic read-modify-write operation
func (c *Client) AtomicUpdate(ctx context.Context, key, updateFunc string, params map[string]interface{}) error {
    req := &script.ExecutionRequest{
        Function:   "atomicUpdate",
        Parameters: []interface{}{key, updateFunc, params},
    }

    resp, err := c.service.Scripts.Run(c.scriptID, req).Context(ctx).Do()
    if err != nil {
        return fmt.Errorf("execute atomicUpdate: %w", err)
    }

    if resp.Error != nil {
        return fmt.Errorf("script error: %v", resp.Error)
    }

    var result Response
    if err := parseResult(resp.Response.Result, &result); err != nil {
        return fmt.Errorf("parse response: %w", err)
    }

    if !result.Success {
        return fmt.Errorf("atomic update failed: %s - %s", result.Error, result.Message)
    }

    return nil
}

// ListKeys returns all metadata keys (for debugging/management)
func (c *Client) ListKeys(ctx context.Context) ([]string, error) {
    req := &script.ExecutionRequest{Function: "listMetadataKeys"}

    resp, err := c.service.Scripts.Run(c.scriptID, req).Context(ctx).Do()
    if err != nil {
        return nil, fmt.Errorf("execute listMetadataKeys: %w", err)
    }

    var result struct {
        Success bool     `json:"success"`
        Keys    []string `json:"keys"`
        Error   string   `json:"error,omitempty"`
        Message string   `json:"message,omitempty"`
    }

    if err := parseResult(resp.Response.Result, &result); err != nil {
        return nil, fmt.Errorf("parse response: %w", err)
    }

    if !result.Success {
        return nil, fmt.Errorf("list keys failed: %s - %s", result.Error, result.Message)
    }

    return result.Keys, nil
}

// Delete removes a key (requires confirmation for safety)
func (c *Client) Delete(ctx context.Context, key string) error {
    req := &script.ExecutionRequest{
        Function:   "deleteMetadata",
        Parameters: []interface{}{key, key}, // confirmation parameter
    }

    resp, err := c.service.Scripts.Run(c.scriptID, req).Context(ctx).Do()
    if err != nil {
        return fmt.Errorf("execute deleteMetadata: %w", err)
    }

    var result Response
    if err := parseResult(resp.Response.Result, &result); err != nil {
        return fmt.Errorf("parse response: %w", err)
    }

    if !result.Success {
        return fmt.Errorf("delete failed: %s - %s", result.Error, result.Message)
    }

    return nil
}

// Helper function to parse script execution results
func parseResult(result interface{}, target interface{}) error {
    resultBytes, err := json.Marshal(result)
    if err != nil {
        return err
    }
    return json.Unmarshal(resultBytes, target)
}

// Common errors
var (
    ErrKeyNotFound = fmt.Errorf("metadata key not found")
)
```

#### 3. Integration with Gmail Mover
**Note: This is a hypothetical example demonstrating the API usage patterns and not an explicit architectural decision for the final implementation.**

```go
// cmd/gmail-mover/main.go
package main

import (
    "context"
    "encoding/json"
    "io/ioutil"
    "log"
    "time"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    gmail "google.golang.org/api/gmail/v1"
    script "google.golang.org/api/script/v1"
    "google.golang.org/api/option"

    "github.com/yourorg/gmail-mover/metadata"
)

// Configuration for metadata storage
const (
    MetadataScriptID = "YOUR_APPS_SCRIPT_DEPLOYMENT_ID"
    MetadataScopes = "https://www.googleapis.com/auth/script.projects," +
                    "https://www.googleapis.com/auth/script.scriptapp"
)

// SyncState represents the current synchronization state
type SyncState struct {
    LastSyncTime    time.Time              `json:"lastSyncTime"`
    ProcessedEmails map[string]bool        `json:"processedEmails"`
    ErrorCount      int                    `json:"errorCount"`
    Accounts        map[string]AccountSync `json:"accounts"`
}

type AccountSync struct {
    LastMessageID string    `json:"lastMessageId"`
    LastSyncTime  time.Time `json:"lastSyncTime"`
    TotalMessages int       `json:"totalMessages"`
}

// UserConfig represents user preferences and settings
type UserConfig struct {
    MaxMessagesPerRun   int                    `json:"maxMessagesPerRun"`
    DefaultSourceLabel  string                 `json:"defaultSourceLabel"`
    AutoDeleteAfterMove bool                   `json:"autoDeleteAfterMove"`
    AIRules            []AIRule               `json:"aiRules"`
    CustomFilters      map[string]string      `json:"customFilters"`
    LastModified       time.Time              `json:"lastModified"`
}

type AIRule struct {
    Name        string                 `json:"name"`
    Conditions  map[string]interface{} `json:"conditions"`
    Actions     []string               `json:"actions"`
    Confidence  float64                `json:"confidence"`
    LastTested  time.Time              `json:"lastTested"`
    Enabled     bool                   `json:"enabled"`
}

func main() {
    ctx := context.Background()

    // Initialize OAuth2 client with combined scopes
    client, err := initializeOAuth2Client(ctx)
    if err != nil {
        log.Fatalf("OAuth2 initialization failed: %v", err)
    }

    // Create Gmail service
    gmailService, err := gmail.NewService(ctx, option.WithHTTPClient(client))
    if err != nil {
        log.Fatalf("Gmail service creation failed: %v", err)
    }

    // Create Script service for metadata
    scriptService, err := script.NewService(ctx, option.WithHTTPClient(client))
    if err != nil {
        log.Fatalf("Script service creation failed: %v", err)
    }

    // Initialize metadata client
    metaClient := metadata.NewClient(scriptService, MetadataScriptID)

    // Load current sync state
    var syncState SyncState
    err = metaClient.Fetch(ctx, "sync_state", &syncState)
    if err != nil && err != metadata.ErrKeyNotFound {
        log.Fatalf("Failed to load sync state: %v", err)
    }

    // Load user configuration
    var userConfig UserConfig
    err = metaClient.Fetch(ctx, "user_config", &userConfig)
    if err == metadata.ErrKeyNotFound {
        // Initialize default configuration
        userConfig = UserConfig{
            MaxMessagesPerRun:   100,
            DefaultSourceLabel:  "INBOX",
            AutoDeleteAfterMove: true,
            CustomFilters:       make(map[string]string),
            LastModified:        time.Now(),
        }
        
        if err := metaClient.Store(ctx, "user_config", userConfig); err != nil {
            log.Printf("Warning: Failed to store default config: %v", err)
        }
    } else if err != nil {
        log.Fatalf("Failed to load user config: %v", err)
    }

    // Perform email operations using Gmail service...
    // (existing Gmail Mover logic here)

    // Update sync state after successful operation
    syncState.LastSyncTime = time.Now()
    syncState.ErrorCount = 0 // Reset on success
    
    if err := metaClient.Store(ctx, "sync_state", syncState); err != nil {
        log.Printf("Warning: Failed to update sync state: %v", err)
    }

    // Example of atomic counter update for statistics
    statsParams := map[string]interface{}{
        "counterName": "total_processed_emails",
        "increment":   42, // number of emails processed this run
    }
    if err := metaClient.AtomicUpdate(ctx, "statistics", "incrementCounter", statsParams); err != nil {
        log.Printf("Warning: Failed to update statistics: %v", err)
    }
}

func initializeOAuth2Client(ctx context.Context) (*http.Client, error) {
    credentialsData, err := ioutil.ReadFile("credentials.json")
    if err != nil {
        return nil, fmt.Errorf("read credentials: %w", err)
    }

    // Combine all required scopes
    allScopes := []string{
        gmail.GmailModifyScope,
        gmail.GmailLabelsScope,
        "https://www.googleapis.com/auth/script.projects",
        "https://www.googleapis.com/auth/script.scriptapp",
    }

    config, err := google.ConfigFromJSON(credentialsData, allScopes...)
    if err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }

    return getOAuthClient(ctx, config)
}

// getOAuthClient handles token loading/saving (implementation as shown in previous examples)
func getOAuthClient(ctx context.Context, config *oauth2.Config) (*http.Client, error) {
    // Implementation from previous examples...
    return nil, nil // Placeholder
}
```

## Alternatives Considered

### Storage of AI Rules and Historical Data in Apps Script
**Question raised**: Could Apps Script's PropertiesService store SQLite databases for rules and history, particularly with Workspace Add-Ons?  
**Answer**: No. Apps Script's PropertiesService is designed for simple key-value storage of text data, not binary data like SQLite databases. Workspace Add-Ons do not provide enhanced storage capabilities beyond the standard PropertiesService limits (500KB total, 9KB per value). AI rules, learning data, and comprehensive historical data should remain in the external SQLite archive databases as originally planned.

### 1. Google Drive appDataFolder
**Decision**: Rejected in favor of Apps Script  
**Pros**:
- Hidden from user's Drive UI (though still technically accessible)
- Native blob storage with no size restrictions per file
- ETags provide built-in optimistic concurrency control
- Well-documented API with extensive examples

**Cons**:
- **User visibility and deletion risk**: While marketed as "invisible," appDataFolder contents can be accessed and deleted by users through various Drive interfaces and third-party apps
- **Complexity**: Required manual ETag management for concurrency control
- **Lock implementation**: Would need custom lock-file patterns, increasing failure modes
- **File management overhead**: Listing, creating, updating, deleting files adds complexity
- **No built-in atomicity**: Multiple operations required for atomic updates

**Code complexity comparison**:
```go
// Drive appDataFolder approach (rejected)
func (d *DriveStore) AtomicUpdate(key string, updateFunc func(data []byte) []byte) error {
    // 1. Create lock file
    lockID, err := d.acquireLock()
    if err != nil { return err }
    defer d.releaseLock(lockID)
    
    // 2. List files to find target
    files, err := d.listFiles(key)
    if err != nil { return err }
    
    // 3. Download current content
    content, etag, err := d.downloadFile(files[0].Id)
    if err != nil { return err }
    
    // 4. Apply update function
    updated := updateFunc(content)
    
    // 5. Upload with ETag check
    return d.uploadWithETag(files[0].Id, etag, updated)
}

// vs Apps Script approach (accepted)
func (c *Client) AtomicUpdate(key, updateFunc string, params map[string]interface{}) error {
    req := &script.ExecutionRequest{
        Function: "atomicUpdate",
        Parameters: []interface{}{key, updateFunc, params},
    }
    resp, err := c.service.Scripts.Run(c.scriptID, req).Do()
    return handleResponse(resp, err)
}
```

### 2. Gmail Draft Messages for Metadata
**Decision**: Rejected  
**Pros**:
- No additional API enablement required (Gmail API already in use)
- Native JSON storage in message body
- Can leverage existing Gmail API authentication

**Cons**:
- **User visibility**: Drafts appear in Gmail UI, violating invisibility requirement
- **No native locking**: Gmail API lacks transaction support for drafts
- **Size limitations**: Draft size limits could constrain metadata growth
- **Semantic mismatch**: Using email drafts for application state violates principle of least surprise

### 3. Gmail Labels with Encoded Metadata
**Decision**: Rejected  
**Pros**:
- Leverages existing Gmail API access
- Labels are synced across all Gmail clients automatically
- Could potentially use label colors for simple state indication

**Cons**:
- **Severe size constraints**: Label names have strict length limits
- **User visibility**: Labels appear in Gmail interface and can be deleted by users
- **No structured data**: Would require custom encoding/decoding logic
- **No locking mechanism**: Multiple client instances could conflict
- **Alternative considered**: Creating fake emails within labels was also rejected due to user visibility and potential deletion

### 4. External Database (Cloud SQL, Firestore)
**Decision**: Rejected for initial implementation  
**Pros**:
- Unlimited storage capacity and complex query capabilities
- Full transactional support with ACID guarantees
- Horizontal scaling potential for large user bases
- Rich ecosystem of database tools and monitoring

**Cons**:
- **Infrastructure overhead**: Requires separate database provisioning and management
- **Authentication complexity**: Need to map OAuth users to database users
- **Cost implications**: Additional service costs for all users
- **Deployment complexity**: Hosted solution needs database connection management

**Future consideration**: May be reconsidered for enterprise deployment or high-scale SaaS offering.

### 5. Local File Storage
**Decision**: Rejected  
**Pros**:
- Zero network latency for read/write operations
- No external service dependencies or quotas
- Complete user control over data location and backup
- Works offline without internet connectivity

**Cons**:
- **CLI-only limitation**: Works for command-line usage but not web interface
- **No synchronization**: Multiple devices/sessions can't share state
- **Backup/recovery issues**: Users responsible for metadata backup
- **Hosted deployment impossible**: Can't work in serverless/cloud environments

### 6. Embedding in SQLite Archive Database
**Decision**: Rejected for operational metadata  
**Pros**:
- Single database to manage with unified backup/restore procedures
- Rich relational data model with complex query capabilities
- Transactional consistency between operational state and archive data
- No external service dependencies once database is local

**Cons**:
- **Multiple database risk**: Users may have multiple archive databases, causing metadata to become inconsistent
- **Database disassociation**: Archive databases may become disconnected from their originating account, causing metadata loss
- **Access patterns**: Operational metadata has different performance requirements than archive data
- **Backup/sync complexity**: Archive database may be large and infrequently synced

**Note**: Will still be used for email content metadata and AI learning data that's tightly coupled to the archive.

## Implementation Details

### GCP Project Setup
1. **Create or select GCP project**:
    - Go to [Google Cloud Console](https://console.cloud.google.com/)
    - Create a new project or select existing one (e.g., `gmail-mover-project`)

2. **Enable required APIs**:
    - Navigate to "APIs & Services" → "Library"
    - Enable the Gmail API
    - Enable the Apps Script API

3. **Create OAuth Client**:
    - Go to "APIs & Services" → "Credentials"
    - Click "Create Credentials" → "OAuth 2.0 Client IDs"
    - Choose "Desktop Application" type
    - Download the credentials JSON file as `credentials.json`

### Apps Script Deployment
1. **Create Apps Script project**:
    - Go to [script.google.com](https://script.google.com)
    - Click "New Project"
    - Replace default code with the Apps Script functions shown above

2. **Bind to GCP project**:
    - Click ⚙️ "Project Settings" on the left sidebar
    - Under "Google Cloud Platform (GCP) Project", click "Change project"
    - Select "Use existing project" and paste your GCP project number
    - Confirm the change

3. **Deploy as API Executable**:
    - Click "Deploy" button in top-right corner
    - Choose "New deployment"
    - In "Select type" dropdown, pick "API executable"
    - Add description (e.g., "Gmail Mover Metadata Storage")
    - Click "Deploy"
    - **Record the Script ID** shown in the deployment details for use in Go client

4. **Re-deploy after updates**:
    - When Apps Script code changes: Deploy → "Manage deployments" → Click pencil icon → "New version" → "Deploy"

### OAuth Scope Configuration
Required scopes for combined functionality:
- `https://www.googleapis.com/auth/gmail.modify` - Gmail read/write access
- `https://www.googleapis.com/auth/gmail.labels` - Label management (optional)
- `https://www.googleapis.com/auth/script.projects` - Apps Script project access
- `https://www.googleapis.com/auth/script.scriptapp` - Apps Script execution

### Error Handling Strategy
1. **Script-level errors**: Apps Script functions return structured error objects
2. **API-level errors**: Go client wraps HTTP/network errors appropriately
3. **Retry logic**: Implement exponential backoff for transient failures
4. **Graceful degradation**: Application continues functioning if metadata operations fail
5. **Monitoring**: Log metadata operation failures for operational visibility

### Data Organization Strategy
Metadata will be stored using simple key-value pairs appropriate for minimal operational state:
- Account-specific sync cursors and processing checkpoints
- User preferences for Google account integration
- Minimal configuration required for continued operation

### Performance Characteristics
- **Latency**: ~200-500ms per operation (network dependent)
- **Concurrency**: LockService provides reliable mutual exclusion
- **Storage limits**: 500KB total per property store per user (sufficient for operational metadata)
- **Durability**: Google-managed persistence with enterprise-grade reliability
- **Execution quotas**:
    - Consumer accounts: 6 minutes per execution, 30 simultaneous executions per user
    - Google Workspace: Same limits as consumer for script execution
    - No documented daily execution limit for manual script runs via Execution API
    - PropertiesService: 500KB total storage, 9KB per value

## Consequences

### Positive Outcomes
1. **Simplified development**: Single OAuth flow covers all functionality
2. **Zero infrastructure**: No database servers to maintain or scale
3. **Reliable locking**: Built-in mutual exclusion eliminates race conditions
4. **User privacy**: Metadata invisible in Google service UIs
5. **Cost efficiency**: No additional database costs for hosted deployments
6. **Automatic scaling**: Google Apps Script scales transparently with user load

### Trade-offs Accepted
1. **Network dependency**: Each metadata operation requires API round-trip
2. **Latency overhead**: ~200-500ms per operation vs microseconds for local storage
3. **Simple Apps Script approach**: Intentionally keeping Apps Script functions minimal to avoid debugging complexity
4. **Google ecosystem coupling**: Metadata tied to Google Apps Script platform, but this limitation is acceptable as this solution applies only to Gmail and other Google-related content (other platforms will require different storage solutions)
5. **Storage limitations**: 500KB total limit per property store per user (sufficient for minimal operational metadata)

### Risk Mitigation
1. **Google service dependency**: Migration path documented, but this limitation is acceptable since this solution is specifically for Google account integration
2. **Script project maintenance**: Version control Apps Script code, automate deployments
3. **Storage limitations**: Monitor usage, ensure data stays within operational requirements rather than comprehensive logging
4. **Data export capability**: Provide metadata export functionality for user control

## Future Enhancements

### Phase 2 Considerations
1. **Metadata synchronization**: Cross-device state synchronization for desktop app
2. **Backup/restore**: Automated metadata backup to user's Drive or local storage
3. **Performance optimization**: Batch metadata operations to reduce API calls
4. **Advanced locking**: Implement reader/writer locks for better concurrency

### Enterprise Features
1. **Team metadata sharing**: Organization-wide rule and configuration sharing
2. **Audit logging**: Enhanced logging for compliance and debugging
3. **Role-based access**: Different metadata access levels for team members
4. **Custom deployment**: Support for organization-owned Apps Script projects

This architecture decision provides a robust foundation for Gmail Mover's metadata storage needs while maintaining simplicity and user control. The choice prioritizes development velocity and operational reliability over theoretical performance optimization, aligning with the project's goals of user empowerment and data sovereignty.