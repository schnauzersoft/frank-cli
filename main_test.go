package main

import (
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	// Test that the main function doesn't panic
	// This is a basic smoke test to ensure the application can start
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Main function panicked: %v", r)
		}
	}()

	// Test with help flag to avoid actual execution
	os.Args = []string{"frank", "--help"}
	main()
}

func TestHelp(t *testing.T) {
	// Test help command
	os.Args = []string{"frank", "help"}
	main()
}
