package commands

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestStatusCommand(t *testing.T) {
	// Create status command
	cmd := StatusCmd()

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute command
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Failed to execute status command: %v", err)
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output format matches TypeScript version
	expectedStrings := []string{
		"📊 Claude Code Router Status",
		"════════════════════════════════════════",
		"❌ Status: Not Running", // Assuming service is not running in test
		"💡 To start the service:",
		"ccproxy start",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', but it didn't", expected)
			t.Errorf("Full output:\n%s", output)
		}
	}

	// Verify exact number of separator characters (40 equal signs)
	separatorLine := "════════════════════════════════════════"
	if !strings.Contains(output, separatorLine) {
		t.Errorf("Expected separator line with exactly 40 characters, got different")
	}
}