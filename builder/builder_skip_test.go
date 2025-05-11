package builder

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// calculateFileHash returns the SHA256 hash of a file as a hex string
func calculateFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	hashBytes := hash.Sum(nil)
	return hex.EncodeToString(hashBytes), nil
}

// TestBuilderSkipsExistingFiles verifies that the builder skips processing files
// that are already registered in the index and doesn't modify existing output files.
func TestBuilderSkipsExistingFiles(t *testing.T) {
	// Create temporary source and destination directories for tests
	srcDir, err := os.MkdirTemp("", "builder_skip_test_src")
	if err != nil {
		t.Fatalf("Failed to create temporary source directory: %v", err)
	}
	defer os.RemoveAll(srcDir)

	dstDir, err := os.MkdirTemp("", "builder_skip_test_dst")
	if err != nil {
		t.Fatalf("Failed to create temporary destination directory: %v", err)
	}
	defer os.RemoveAll(dstDir)

	// Step 1: Create initial provider file
	initialProvider := filepath.Join(srcDir, "terraform-provider-skip_v1.0.0_linux_amd64")
	initialContent := "initial binary content"
	err = os.WriteFile(initialProvider, []byte(initialContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create initial test file: %v", err)
	}

	// Step 2: Run the builder for the first time
	b := New(srcDir, dstDir)
	err = b.Build()
	if err != nil {
		t.Fatalf("First Build() error = %v", err)
	}

	// Step 3: Collect file paths and their hashes
	var filePaths []string
	var fileHashes = make(map[string]string)

	// Paths to check
	expectedPaths := []string{
		filepath.Join(dstDir, "skip", "versions", "index.json"),
		filepath.Join(dstDir, "skip", "1.0.0", "download", "linux", "amd64", "terraform-provider-skip_v1.0.0_linux_amd64.zip"),
		filepath.Join(dstDir, "skip", "1.0.0", "download", "linux", "amd64", "terraform-provider-skip_v1.0.0_linux_amd64_SHA256SUMS"),
		filepath.Join(dstDir, "skip", "1.0.0", "download", "linux", "amd64", "terraform-provider-skip_v1.0.0_linux_amd64_SHA256SUMS.sig"),
		filepath.Join(dstDir, "skip", "1.0.0", "download", "linux", "amd64", "index.json"),
	}

	// Calculate and save hashes of all expected files
	for _, path := range expectedPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatalf("Expected file not created: %s", path)
		}

		hash, err := calculateFileHash(path)
		if err != nil {
			t.Fatalf("Failed to calculate hash for %s: %v", path, err)
		}

		filePaths = append(filePaths, path)
		fileHashes[path] = hash
	}

	// Step 4: Create a new source file with different content but same name
	// First, remove the original file
	err = os.Remove(initialProvider)
	if err != nil {
		t.Fatalf("Failed to remove initial provider file: %v", err)
	}

	// Create a new provider file with different content
	newContent := "modified binary content that should be ignored"
	err = os.WriteFile(initialProvider, []byte(newContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create modified test file: %v", err)
	}

	// Step 5: Run the builder again
	err = b.Build()
	if err != nil {
		t.Fatalf("Second Build() error = %v", err)
	}

	// Step 6: Verify that all file hashes remain unchanged
	for _, path := range filePaths {
		newHash, err := calculateFileHash(path)
		if err != nil {
			t.Fatalf("Failed to calculate second hash for %s: %v", path, err)
		}

		originalHash := fileHashes[path]
		if newHash != originalHash {
			t.Errorf("File hash changed for %s:\nOriginal: %s\nNew: %s", path, originalHash, newHash)
		}
	}
}
