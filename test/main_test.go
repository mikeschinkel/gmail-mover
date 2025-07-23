package test

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Setup code here if needed
	// For example: initialize test data, mock services, etc.
	
	// Run tests
	code := m.Run()
	
	// Cleanup code here if needed
	
	os.Exit(code)
}