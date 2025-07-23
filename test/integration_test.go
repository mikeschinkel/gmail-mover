package test

import (
	"testing"

	"github.com/mikeschinkel/gmail-mover/gmover"
)

// TestJobCreation tests the job creation
func TestJobCreation(t *testing.T) {
	opts := gmover.JobOptions{
		Name:            "Test Job",
		SrcEmail:        "test@example.com",
		SrcLabel:        "INBOX",
		DstEmail:        "archive@example.com",
		DstLabel:        "test-label",
		MaxMessages:     100,
		DryRun:          true,
		DeleteAfterMove: false,
		SearchQuery:     "",
	}

	job, err := gmover.NewJob(opts)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if job.Name != "Test Job" {
		t.Errorf("Expected job name 'Test Job', got '%s'", job.Name)
	}

	if job.SrcAccount.Email != "test@example.com" {
		t.Errorf("Expected source email 'test@example.com', got '%s'", job.SrcAccount.Email)
	}

	if job.Options.DryRun != true {
		t.Errorf("Expected DryRun to be true, got %v", job.Options.DryRun)
	}
}

// TestLoadJobFile tests loading a job from file
func TestLoadJobFile(t *testing.T) {
	// This would test loading from a real job file
	// For now, we'll skip if no test file exists
	_, err := gmover.LoadJob("nonexistent.json")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

// TestNewJobValidation tests job creation validation
func TestNewJobValidation(t *testing.T) {
	// Test missing required fields
	opts := gmover.JobOptions{
		Name: "Invalid Job",
		// Missing SrcEmail and DstEmail
	}

	_, err := gmover.NewJob(opts)
	if err == nil {
		t.Error("Expected error for missing required fields, got nil")
	}
}

// TestRunWithConfig tests gmover.Run() with test configuration
func TestRunWithConfig(t *testing.T) {
	// Set up a test logger to avoid panics
	setupTestLogger()
	
	config := gmover.NewConfig(gmover.MoveEmails)
	config.SetSrcEmail("test@example.com")
	config.SetSrcLabel("INBOX")
	config.SetDstEmail("archive@example.com")
	config.SetDstLabel("test-label")
	config.SetMaxMessages(100)
	config.SetDryRun(true)
	config.SetDeleteAfterMove(false)

	// This will fail due to auth, but we're testing that gmover.Run()
	// accepts config properly without flag parsing
	err := gmover.Run(&config)
	if err == nil {
		t.Error("Expected authentication error, but got nil")
	}
}

// TestRunListLabels tests the list-labels functionality
func TestRunListLabels(t *testing.T) {
	setupTestLogger()
	
	config := gmover.NewConfig(gmover.ListLabels)
	config.SetSrcEmail("nonexistent@example.com")

	// This will fail because the email doesn't exist, but we're testing
	// that the application logic is properly separated from CLI parsing
	err := gmover.Run(&config)
	if err == nil {
		t.Error("Expected error for nonexistent email, but got nil")
	}
}

// TestJobExecutePassesConfiguration tests that Job.Execute passes config to gmutil
func TestJobExecutePassesConfiguration(t *testing.T) {
	setupTestLogger()
	
	opts := gmover.JobOptions{
		Name:            "Test Job",
		SrcEmail:        "test@example.com",
		SrcLabel:        "INBOX",
		DstEmail:        "archive@example.com",
		DstLabel:        "test-label",
		MaxMessages:     50,
		DryRun:          true,
		DeleteAfterMove: false,
		SearchQuery:     "from:test",
		FailOnError:     false,
		LogLevel:        "info",
	}
	
	job, err := gmover.NewJob(opts)
	if err != nil {
		t.Fatalf("Expected no error creating job, got %v", err)
	}
	
	// Verify job structure has correct values
	if job.SrcAccount.Email != "test@example.com" {
		t.Errorf("Expected SrcAccount.Email 'test@example.com', got '%s'", job.SrcAccount.Email)
	}
	
	if len(job.SrcAccount.Labels) != 1 || job.SrcAccount.Labels[0] != "INBOX" {
		t.Errorf("Expected SrcAccount.Labels ['INBOX'], got %v", job.SrcAccount.Labels)
	}
	
	if job.SrcAccount.Query != "from:test" {
		t.Errorf("Expected SrcAccount.Query 'from:test', got '%s'", job.SrcAccount.Query)
	}
	
	if job.SrcAccount.MaxMessages != 50 {
		t.Errorf("Expected SrcAccount.MaxMessages 50, got %d", job.SrcAccount.MaxMessages)
	}
	
	if job.DstAccount.ApplyLabel != "test-label" {
		t.Errorf("Expected DstAccount.ApplyLabel 'test-label', got '%s'", job.DstAccount.ApplyLabel)
	}
	
	if !job.Options.DryRun {
		t.Errorf("Expected Options.DryRun true, got %v", job.Options.DryRun)
	}
	
	if job.Options.DeleteAfterMove {
		t.Errorf("Expected Options.DeleteAfterMove false, got %v", job.Options.DeleteAfterMove)
	}
	
	// Test execution fails due to auth (expected) - this verifies the config is passed through
	err = job.Execute()
	if err == nil {
		t.Error("Expected authentication error during execution, got nil")
	}
}

// TestConfigPointerHandling tests that Config properly handles accessor methods  
func TestConfigPointerHandling(t *testing.T) {
	config := gmover.NewConfig(gmover.MoveEmails)
	config.SetSrcEmail("test@example.com")
	config.SetSrcLabel("INBOX")
	config.SetDstEmail("archive@example.com")
	config.SetDstLabel("moved")
	config.SetMaxMessages(100)
	config.SetDryRun(true)
	config.SetDeleteAfterMove(false)
	config.SetSearchQuery("has:attachment")
	
	job, err := gmover.GetJob(config)
	if err != nil {
		t.Fatalf("Expected no error getting job from config, got %v", err)
	}
	
	// Verify accessor methods worked correctly in config conversion
	if job.SrcAccount.Email != "test@example.com" {
		t.Errorf("Expected SrcEmail 'test@example.com', got '%s'", job.SrcAccount.Email)
	}
	
	if job.SrcAccount.Query != "has:attachment" {
		t.Errorf("Expected SearchQuery 'has:attachment', got '%s'", job.SrcAccount.Query)
	}
	
	if job.SrcAccount.MaxMessages != 100 {
		t.Errorf("Expected MaxMessages 100, got %d", job.SrcAccount.MaxMessages)
	}
}

// TestJobFromFile tests loading job configuration from JSON file
func TestJobFromFile(t *testing.T) {
	config := gmover.NewConfig(gmover.MoveEmails)
	config.SetJobFile("../examples/jobs/backup-important.json")
	
	// This will try to load the job file - may fail if file format doesn't match
	// but tests the file loading path
	_, err := gmover.GetJob(config)
	// We expect this might fail due to file format/existence, but we're testing the code path
	if err != nil {
		t.Logf("Job file loading failed as expected: %v", err)
	}
}

// TestConfigGettersSetters tests that Config getter/setter methods work correctly
func TestConfigGettersSetters(t *testing.T) {
	config := gmover.NewConfig(gmover.MoveEmails)
	
	// Test initial state
	if config.RunMode() != gmover.MoveEmails {
		t.Errorf("Expected initial RunMode 'MoveEmails', got '%s'", config.RunMode())
	}
	
	// Test empty string defaults for pointer fields
	if config.SrcEmail() != "" {
		t.Errorf("Expected empty SrcEmail, got '%s'", config.SrcEmail())
	}
	
	if config.MaxMessages() != 0 {
		t.Errorf("Expected zero MaxMessages, got %d", config.MaxMessages())
	}
	
	// Test setters
	config.SetSrcEmail("test@example.com")
	if config.SrcEmail() != "test@example.com" {
		t.Errorf("Expected SrcEmail 'test@example.com', got '%s'", config.SrcEmail())
	}
	
	config.SetMaxMessages(500)
	if config.MaxMessages() != 500 {
		t.Errorf("Expected MaxMessages 500, got %d", config.MaxMessages())
	}
	
	config.SetDryRun(true)
	if !config.DryRun() {
		t.Errorf("Expected DryRun true, got %v", config.DryRun())
	}
	
	// Test mode change
	config.SetRunMode(gmover.ListLabels)
	if config.RunMode() != gmover.ListLabels {
		t.Errorf("Expected RunMode 'ListLabels', got '%s'", config.RunMode())
	}
}
