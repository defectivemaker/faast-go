package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegration(t *testing.T) {
	// Find and compile the main package
	mainDir, err := findMainPackageDir()
	if err != nil {
		t.Fatalf("Failed to find main package: %v", err)
	}

	binaryPath := filepath.Join(t.TempDir(), "faast-go")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = mainDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to compile main package: %v\n%s", err, output)
	}

	// Create a temporary wordlist file
	wordlistContent := "word1\nword2\nword3"
	wordlistFile, err := os.CreateTemp("", "wordlist1.*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(wordlistFile.Name())
	if _, err := wordlistFile.Write([]byte(wordlistContent)); err != nil {
		t.Fatal(err)
	}
	wordlistFile.Close()

	// Create a temporary config file
	configContent := `
type: payload
endpoint: https://google.com
fields:
  - field1
  - field2
wordlists:
  - ` + wordlistFile.Name() + `
staticValues:
  - static1
cookies:
  - session=abc123
validateType: code
sizeDefault: 100
codeDefault: 404
rateLimit: 10
timeout: 5
shardIndex: 0
numShards: 1
`
	configFile, err := os.CreateTemp("", "config.*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(configFile.Name())
	if _, err := configFile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	configFile.Close()

	// Run the compiled binary
	cmd = exec.Command(binaryPath, configFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Program exited with error: %v\nOutput: %s", err, output)
	}

	// Log the output for debugging
	t.Logf("Program output:\n%s", string(output))

	// Add assertions based on expected behavior
	if len(output) == 0 {
		t.Error("Expected some output, but got none")
	}

	if strings.Contains(string(output), "error") || strings.Contains(string(output), "Error") {
		t.Errorf("Output contains error message: %s", string(output))
	}

	// Add more specific checks based on your program's expected behavior
}

func TestNoArgumentsBehavior(t *testing.T) {
	mainDir, err := findMainPackageDir()
	if err != nil {
		t.Fatalf("Failed to find main package: %v", err)
	}

	binaryPath := filepath.Join(t.TempDir(), "faast-go")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = mainDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to compile main package: %v\n%s", err, output)
	}

	cmd = exec.Command(binaryPath)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected an error when no config file is provided, but got none")
	}

	expectedError := "Please provide a YAML config file"
	if !strings.Contains(string(output), expectedError) {
		t.Errorf("Expected error message '%s', but got: %s", expectedError, output)
	}
}

func findMainPackageDir() (string, error) {
	// Start from the current directory and search upwards
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		cmdDir := filepath.Join(dir, "cmd")
		entries, err := os.ReadDir(cmdDir)
		if err == nil {
			for _, entry := range entries {
				if entry.Name() == "main.go" {
					return cmdDir, nil
				}
			}
		}

		// Move up one directory
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			// We've reached the root directory
			break
		}
		dir = parentDir
	}

	return "", os.ErrNotExist
}
