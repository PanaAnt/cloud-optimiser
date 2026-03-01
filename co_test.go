package main

import (
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestSmoke_Help verifies the help command works
func TestSmoke_Help(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--help")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Help command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	expectedStrings := []string{
		"cloud-optimiser",
		"Usage:",
		"Available Commands:",
		"discover",
		"recommend",
		"mode",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected help output to contain '%s'", expected)
		}
	}

	t.Log("Help command works")
}

// TestSmoke_DiscoverMock verifies discover command works with mock data
func TestSmoke_DiscoverMock(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "discover", "--use-mock")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Discover command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	expectedStrings := []string{
		"MODE:",
		"Mock",
		"EC2 Instances",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected discover output to contain '%s'", expected)
		}
	}

	t.Log("Discover command works with mock data")
}

// TestSmoke_RecommendMock verifies recommend command works with mock data
func TestSmoke_RecommendMock(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "recommend", "--use-mock")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Recommend command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	expectedStrings := []string{
		"MODE:",
		"Mock",
		"ID",
		"TYPE",
		"ACTION",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected recommend output to contain '%s'", expected)
		}
	}

	t.Log("Recommend command works with mock data")
}

// TestSmoke_RecommendJSON verifies JSON output format works
func TestSmoke_RecommendJSON(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "recommend", "--use-mock", "--output", "json")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Recommend JSON command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	if !strings.Contains(outputStr, "{") || !strings.Contains(outputStr, "}") {
		t.Error("Expected JSON output with braces")
	}

	if !strings.Contains(outputStr, "instance_id") {
		t.Error("Expected JSON to contain 'instance_id' field")
	}

	t.Log("JSON output format works")
}

// TestSmoke_ModeCommands verifies mode management works
func TestSmoke_ModeCommands(t *testing.T) {
	// Set to mock mode
	cmd := exec.Command("go", "run", ".", "mode", "set", "mock")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Mode set command failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "mock") {
		t.Error("Expected confirmation of mock mode")
	}

	// Show current mode
	cmd = exec.Command("go", "run", ".", "mode", "show")
	output, err = cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Mode show command failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "mock") {
		t.Error("Expected mode show to display 'mock'")
	}

	t.Log("Mode commands work")
}

// TestSmoke_Filtering verifies filtering flags work
func TestSmoke_Filtering(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "Only downsize",
			args: []string{"recommend", "--use-mock", "--only-downsize"},
		},
		{
			name: "Only upsize",
			args: []string{"recommend", "--use-mock", "--only-upsize"},
		},
		{
			name: "State filter",
			args: []string{"recommend", "--use-mock", "--state", "running"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("go", "run", ".")
			cmd.Args = append(cmd.Args, tt.args...)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Fatalf("%s failed: %v\nOutput: %s", tt.name, err, output)
			}

			if len(output) == 0 {
				t.Error("Expected non-empty output")
			}
		})
	}

	t.Log("Filtering flags work")
}

// TestSmoke_Sorting verifies sorting flags work
func TestSmoke_Sorting(t *testing.T) {
	sortOptions := []string{"cpu", "cost", "savings"}

	for _, sortBy := range sortOptions {
		t.Run("Sort by "+sortBy, func(t *testing.T) {
			cmd := exec.Command("go", "run", ".", "recommend", "--use-mock", "--sort", sortBy)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Fatalf("Sort by %s failed: %v\nOutput: %s", sortBy, err, output)
			}

			if len(output) == 0 {
				t.Error("Expected non-empty output")
			}
		})
	}

	t.Log("Sorting flags work")
}

// TestSmoke_Performance verifies commands complete in reasonable time
func TestSmoke_Performance(t *testing.T) {
	start := time.Now()
	cmd := exec.Command("go", "run", ".", "recommend", "--use-mock")
	_, err := cmd.CombinedOutput()
	duration := time.Since(start)

	if err != nil {
		t.Logf("Command completed with error: %v (acceptable for smoke test)", err)
	}

	// Should complete in under 30 seconds (generous for go run)
	if duration > 30*time.Second {
		t.Errorf("Command took too long: %v (expected < 30s)", duration)
	}

	t.Logf("Performance acceptable (completed in %v)", duration)
}

// TestSmoke_DebugFlag verifies debug flag doesn't break commands
func TestSmoke_DebugFlag(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--debug", "discover", "--use-mock")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Debug flag failed: %v\nOutput: %s", err, output)
	}

	if len(output) == 0 {
		t.Error("Expected non-empty output with debug flag")
	}

	t.Log("Debug flag works")
}

// TestSmoke_InvalidCommand verifies error handling
func TestSmoke_InvalidCommand(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "nonexistent-command")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("Expected error for invalid command")
	}

	outputStr := strings.ToLower(string(output))
	if !strings.Contains(outputStr, "unknown") && !strings.Contains(outputStr, "error") {
		t.Error("Expected error message for invalid command")
	}

	t.Log("Invalid command handling works")
}

// TestSmoke_QuickRun runs all basic commands quickly
func TestSmoke_QuickRun(t *testing.T) {
	commands := []struct {
		name string
		args []string
	}{
		{"Help", []string{"--help"}},
		{"Discover", []string{"discover", "--use-mock"}},
		{"Recommend", []string{"recommend", "--use-mock"}},
		{"Mode Show", []string{"mode", "show"}},
	}

	for _, cmd := range commands {
		t.Run(cmd.name, func(t *testing.T) {
			start := time.Now()
			execCmd := exec.Command("go", "run", ".")
			execCmd.Args = append(execCmd.Args, cmd.args...)
			
			output, err := execCmd.CombinedOutput()
			duration := time.Since(start)

			if err != nil && !strings.Contains(cmd.name, "Help") {
				t.Logf("%s: completed with status (may be expected): %v", cmd.name, err)
			}

			if len(output) == 0 {
				t.Errorf("%s: expected non-empty output", cmd.name)
			}

			t.Logf("%s completed in %v", cmd.name, duration)
		})
	}
}