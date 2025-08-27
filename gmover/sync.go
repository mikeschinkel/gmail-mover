package gmover

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mikeschinkel/gmover/sqlcgen"

	"github.com/mikeschinkel/gmover/gmcfg"
)

// SyncOptions contains options for the sync operation
type SyncOptions struct {
	Account   EmailAddress
	DBName    string
	Label     string // Optional specific label to sync
	Query     string // Optional Gmail search query
	Force     bool   // Force full resync
	DryRun    bool   // Preview mode
	BatchSize int    // Messages per batch (0 = auto)
}

// SyncResult contains the results of a sync operation
type SyncResult struct {
	MessagesProcessed int
	MessagesStored    int
	ErrorsEncountered int
	SyncDuration      time.Duration
}

// RunSync executes the Gmail sync operation
func RunSync(ctx context.Context, opts SyncOptions) (result SyncResult, err error) {
	var startTime time.Time
	var dbConfig gmcfg.DatabaseConfig
	var db *sql.DB
	var queries *sqlcgen.Queries
	var dbUUID string
	var metadataClient *MetadataClient

	startTime = time.Now()

	// Validate required options
	if opts.Account.IsZero() {
		err = fmt.Errorf("account email is required")
		goto end
	}

	if opts.DBName == "" {
		opts.DBName = "default"
	}

	logger.Info("Starting Gmail sync",
		"account", opts.Account.String(),
		"database", opts.DBName,
		"dry_run", opts.DryRun)

	// Load database configuration
	dbConfig, err = loadDatabaseConfig(opts.DBName)
	if err != nil {
		goto end
	}

	// Connect to database
	db, queries, err = connectToDatabase(ctx, dbConfig)
	if err != nil {
		goto end
	}
	defer func() {
		if db != nil {
			mustClose(db)
		}
	}()

	// Get or create database UUID
	dbUUID, err = getOrCreateDatabaseUUID(ctx, queries)
	if err != nil {
		goto end
	}

	logger.Info("Connected to database", "uuid", dbUUID, "type", dbConfig.Type)

	// Initialize metadata client (stub for now)
	metadataClient = NewMetadataClient()

	// Register this database if first time
	err = registerDatabaseIfNeeded(ctx, metadataClient, opts.Account.String(), dbUUID, opts.DBName)
	if err != nil {
		// Log error but continue - this is just metadata tracking
		logger.Warn("Failed to register database", "error", err)
	}

	if opts.DryRun {
		logger.Info("DRY RUN: Would sync Gmail account to database")
		result.SyncDuration = time.Since(startTime)
		goto end
	}

	// TODO: Implement actual sync logic
	// For now, just demonstrate the structure
	result.MessagesProcessed = 0
	result.MessagesStored = 0
	result.ErrorsEncountered = 0

	logger.Info("STUB: Gmail sync not yet fully implemented")
	err = fmt.Errorf("sync implementation coming soon - basic structure is ready")

end:
	result.SyncDuration = time.Since(startTime)

	if err == nil {
		logger.Info("Gmail sync completed",
			"processed", result.MessagesProcessed,
			"stored", result.MessagesStored,
			"errors", result.ErrorsEncountered,
			"duration", result.SyncDuration)
	}

	return result, err
}

// loadDatabaseConfig loads the configuration for a named database
func loadDatabaseConfig(dbName string) (config gmcfg.DatabaseConfig, err error) {
	var store *gmcfg.FileStore

	store = gmcfg.NewFileStore(AppName)
	config, err = store.GetDatabaseConfig(dbName)
	if err != nil {
		goto end
	}

	// Validate configuration
	if config.Type != "sqlite" {
		err = fmt.Errorf("only SQLite databases supported currently, got: %s", config.Type)
		goto end
	}

	if config.Path == "" {
		err = fmt.Errorf("database path is required")
		goto end
	}

end:
	return config, err
}

// connectToDatabase establishes connection to SQLite database
func connectToDatabase(ctx context.Context, config gmcfg.DatabaseConfig) (db *sql.DB, queries *sqlcgen.Queries, err error) {
	var dbDir string

	// Ensure directory exists
	dbDir = filepath.Dir(config.Path)
	err = os.MkdirAll(dbDir, 0755)
	if err != nil {
		goto end
	}

	// Connect to SQLite database
	db, err = sql.Open("sqlite3", config.Path)
	if err != nil {
		goto end
	}

	// Test connection
	err = db.PingContext(ctx)
	if err != nil {
		goto end
	}

	// TODO: Ensure schema exists - will implement with SQLC schema embedding
	// For now, assume user has run schema creation manually

	queries = sqlcgen.New(db)

end:
	return db, queries, err
}

// getOrCreateDatabaseUUID gets existing UUID or creates new one
func getOrCreateDatabaseUUID(ctx context.Context, queries *sqlcgen.Queries) (dbUUID string, err error) {
	const dbUUIDKey = "db_id"

	// Try to get existing UUID
	dbUUID, err = queries.GetMetadata(ctx, dbUUIDKey)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create new UUID
			dbUUID = uuid.New().String()
			err = queries.SetMetadata(ctx, sqlcgen.SetMetadataParams{
				Key:   dbUUIDKey,
				Value: dbUUID,
			})
			if err != nil {
				goto end
			}
			logger.Info("Created new database UUID", "uuid", dbUUID)
		} else {
			goto end
		}
	}

end:
	return dbUUID, err
}

// registerDatabaseIfNeeded registers database with metadata service if not already registered
func registerDatabaseIfNeeded(ctx context.Context, client *MetadataClient, account, dbUUID, dbName string) (err error) {
	var machineID string
	var machineName string
	var registration DBRegistration

	machineID, err = GetMachineID()
	if err != nil {
		goto end
	}

	machineName, err = getMachineName()
	if err != nil {
		goto end
	}

	registration = DBRegistration{
		MachineID:    machineID,
		MachineName:  machineName,
		DatabaseID:   dbUUID,
		DatabaseName: dbName,
		LastSync:     time.Now(),
		// TODO: Set OldestEmail and NewestEmail when we have data
	}

	err = client.RegisterDatabase(ctx, account, registration)

end:
	return err
}
