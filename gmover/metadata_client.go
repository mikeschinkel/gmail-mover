package gmover

import (
	"context"
	"fmt"
	"time"
)

// AccountSyncState represents sync state stored in Apps Script (ADR-006)
type AccountSyncState struct {
	Account       string                     `json:"account"`
	Watermark     time.Time                  `json:"watermark"` // Latest fully-synced time
	Slices        map[string]*SliceStatus    `json:"slices"`    // Active slices only
	RegisteredDBs map[string]*DBRegistration `json:"registeredDBs"`
	ImportQueue   map[string][]string        `json:"importQueue"` // Per-DB import labels to process
}

// DBRegistration represents a database registered for sync
type DBRegistration struct {
	MachineID    string    `json:"machineId"`    // MAC address or hostname
	MachineName  string    `json:"machineName"`  // User-friendly name
	DatabaseID   string    `json:"databaseId"`   // UUID from DB metadata
	DatabaseName string    `json:"databaseName"` // Config name
	LastSync     time.Time `json:"lastSync"`
	OldestEmail  time.Time `json:"oldestEmail"` // Earliest email seen
	NewestEmail  time.Time `json:"newestEmail"` // Latest email seen
}

// SliceStatus represents the status of a time slice
type SliceStatus struct {
	Status     string   `json:"status"` // "processing", "complete"
	Messages   int      `json:"messages,omitempty"`
	PendingDBs []string `json:"pendingDbs,omitempty"` // For import slices
}

// MetadataClient provides access to Apps Script metadata storage
type MetadataClient struct {
	// TODO: Add OAuth client for Apps Script Execution API
	// For now, this is a stub implementation
}

// NewMetadataClient creates a new metadata client
func NewMetadataClient() *MetadataClient {
	return &MetadataClient{}
}

// GetAccountSyncState retrieves sync state for a Gmail account
func (c *MetadataClient) GetAccountSyncState(_ context.Context, account string) (state AccountSyncState, err error) {
	// TODO: Implement actual Apps Script API call per ADR-006
	// For now, return empty state as stub

	state = AccountSyncState{
		Account:       account,
		Watermark:     time.Time{}, // Zero time means no previous sync
		Slices:        make(map[string]*SliceStatus),
		RegisteredDBs: make(map[string]*DBRegistration),
		ImportQueue:   make(map[string][]string),
	}

	// Stub implementation - would be replaced with Apps Script API call
	err = fmt.Errorf("STUB: Apps Script metadata client not yet implemented")

	return state, err
}

// SaveAccountSyncState saves sync state for a Gmail account
func (c *MetadataClient) SaveAccountSyncState(_ context.Context, state AccountSyncState) (err error) {
	// TODO: Implement actual Apps Script API call per ADR-006

	// Validate state before saving
	if state.Account == "" {
		err = fmt.Errorf("account is required")
		goto end
	}

	// Stub implementation - would be replaced with Apps Script API call
	err = fmt.Errorf("STUB: Apps Script metadata client not yet implemented")

end:
	return err
}

// RegisterDatabase registers a database for sync tracking
func (c *MetadataClient) RegisterDatabase(ctx context.Context, account string, registration DBRegistration) (err error) {
	var state AccountSyncState

	// Get current state
	state, err = c.GetAccountSyncState(ctx, account)
	if err != nil {
		// For stub, ignore the error and create new state
		state = AccountSyncState{
			Account:       account,
			Watermark:     time.Time{},
			Slices:        make(map[string]*SliceStatus),
			RegisteredDBs: make(map[string]*DBRegistration),
			ImportQueue:   make(map[string][]string),
		}
	}

	// Add registration
	state.RegisteredDBs[registration.DatabaseID] = &registration

	// Save updated state
	err = c.SaveAccountSyncState(ctx, state)

	return err
}

// UpdateSyncProgress updates the watermark when slices are complete
func (c *MetadataClient) UpdateSyncProgress(ctx context.Context, account string, watermark time.Time) (err error) {
	var state AccountSyncState

	state, err = c.GetAccountSyncState(ctx, account)
	if err != nil {
		goto end
	}

	state.Watermark = watermark
	err = c.SaveAccountSyncState(ctx, state)

end:
	return err
}

// GetMachineID returns a unique identifier for this machine
func GetMachineID() (machineID string, err error) {
	// TODO: Implement MAC address detection as shown in TODO-NEXT.md
	// For now, use hostname as fallback

	var hostname string
	hostname, err = getMachineName()
	if err != nil {
		goto end
	}

	machineID = "stub:" + hostname

end:
	return machineID, err
}

// getMachineName returns a user-friendly machine name
func getMachineName() (name string, err error) {
	// TODO: Get better machine name (hostname + user)
	name = "development-machine"

	return name, err
}
