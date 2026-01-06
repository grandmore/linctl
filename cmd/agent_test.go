package cmd

import (
	"testing"
)

func TestAgentCommandExists(t *testing.T) {
	if agentCmd == nil {
		t.Fatal("agentCmd should not be nil")
	}

	if agentCmd.Use != "agent [issue-id]" {
		t.Errorf("Expected Use 'agent [issue-id]', got '%s'", agentCmd.Use)
	}

	if agentCmd.Short != "View agent session for an issue" {
		t.Errorf("Expected Short description mismatch, got '%s'", agentCmd.Short)
	}
}

func TestDelegateFlagExists(t *testing.T) {
	// Check update command has delegate flag
	flag := issueUpdateCmd.Flags().Lookup("delegate")
	if flag == nil {
		t.Fatal("issueUpdateCmd should have --delegate flag")
	}
	if flag.Usage != "Delegate to agent (email, name, displayName, or 'none' to remove)" {
		t.Errorf("Unexpected delegate flag usage: %s", flag.Usage)
	}

	// Check create command has delegate flag
	flag = issueCreateCmd.Flags().Lookup("delegate")
	if flag == nil {
		t.Fatal("issueCreateCmd should have --delegate flag")
	}
}
