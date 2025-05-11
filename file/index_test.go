package file

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestVersionsIndex(t *testing.T) {
	// Create temporary directory for tests
	testDir, err := os.MkdirTemp("", "versions_index_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	testIndexPath := filepath.Join(testDir, "versions", "index.json")

	// Test creating a new index
	t.Run("create new index", func(t *testing.T) {
		index, err := ReadVersionsIndex(testIndexPath, "test")
		if err != nil {
			t.Fatalf("ReadVersionsIndex error: %v", err)
		}

		if index.ID != "test" {
			t.Errorf("Index ID = %s, want 'test'", index.ID)
		}

		if len(index.Versions) != 0 {
			t.Errorf("New index has %d versions, want 0", len(index.Versions))
		}
	})

	// Test adding a version
	t.Run("add version", func(t *testing.T) {
		index, err := ReadVersionsIndex(testIndexPath, "test")
		if err != nil {
			t.Fatalf("ReadVersionsIndex error: %v", err)
		}

		index.AddVersion("1.0.0", "linux", "amd64")
		if err := WriteVersionsIndex(testIndexPath, index); err != nil {
			t.Fatalf("WriteVersionsIndex error: %v", err)
		}

		// Read it back
		index, err = ReadVersionsIndex(testIndexPath, "test")
		if err != nil {
			t.Fatalf("ReadVersionsIndex error: %v", err)
		}

		if len(index.Versions) != 1 {
			t.Fatalf("Index has %d versions, want 1", len(index.Versions))
		}

		if index.Versions[0].Version != "1.0.0" {
			t.Errorf("Version = %s, want '1.0.0'", index.Versions[0].Version)
		}

		if len(index.Versions[0].Platforms) != 1 {
			t.Fatalf("Version has %d platforms, want 1", len(index.Versions[0].Platforms))
		}

		if index.Versions[0].Platforms[0].OS != "linux" || index.Versions[0].Platforms[0].Arch != "amd64" {
			t.Errorf("Platform = %s_%s, want 'linux_amd64'", index.Versions[0].Platforms[0].OS, index.Versions[0].Platforms[0].Arch)
		}
	})

	// Test adding the same version with a different platform
	t.Run("add platform to existing version", func(t *testing.T) {
		index, err := ReadVersionsIndex(testIndexPath, "test")
		if err != nil {
			t.Fatalf("ReadVersionsIndex error: %v", err)
		}

		index.AddVersion("1.0.0", "darwin", "arm64")
		if err := WriteVersionsIndex(testIndexPath, index); err != nil {
			t.Fatalf("WriteVersionsIndex error: %v", err)
		}

		// Read it back
		index, err = ReadVersionsIndex(testIndexPath, "test")
		if err != nil {
			t.Fatalf("ReadVersionsIndex error: %v", err)
		}

		if len(index.Versions) != 1 {
			t.Fatalf("Index has %d versions, want 1", len(index.Versions))
		}

		if index.Versions[0].Version != "1.0.0" {
			t.Errorf("Version = %s, want '1.0.0'", index.Versions[0].Version)
		}

		if len(index.Versions[0].Platforms) != 2 {
			t.Fatalf("Version has %d platforms, want 2", len(index.Versions[0].Platforms))
		}

		// Check both platforms exist
		foundLinuxAmd64 := false
		foundDarwinArm64 := false

		for _, platform := range index.Versions[0].Platforms {
			if platform.OS == "linux" && platform.Arch == "amd64" {
				foundLinuxAmd64 = true
			}
			if platform.OS == "darwin" && platform.Arch == "arm64" {
				foundDarwinArm64 = true
			}
		}

		if !foundLinuxAmd64 {
			t.Errorf("Missing platform linux_amd64")
		}
		if !foundDarwinArm64 {
			t.Errorf("Missing platform darwin_arm64")
		}
	})

	// Test adding a new version
	t.Run("add new version", func(t *testing.T) {
		index, err := ReadVersionsIndex(testIndexPath, "test")
		if err != nil {
			t.Fatalf("ReadVersionsIndex error: %v", err)
		}

		index.AddVersion("2.0.0", "windows", "amd64")
		if err := WriteVersionsIndex(testIndexPath, index); err != nil {
			t.Fatalf("WriteVersionsIndex error: %v", err)
		}

		// Read it back
		index, err = ReadVersionsIndex(testIndexPath, "test")
		if err != nil {
			t.Fatalf("ReadVersionsIndex error: %v", err)
		}

		if len(index.Versions) != 2 {
			t.Fatalf("Index has %d versions, want 2", len(index.Versions))
		}

		// Check version sorting (newest first)
		if index.Versions[0].Version != "2.0.0" || index.Versions[1].Version != "1.0.0" {
			t.Errorf("Versions not sorted correctly: got [%s, %s], want [2.0.0, 1.0.0]",
				index.Versions[0].Version, index.Versions[1].Version)
		}
	})

	// Test reading from a manually created file
	t.Run("read existing file", func(t *testing.T) {
		// Create a test file with known content
		manualIndexPath := filepath.Join(testDir, "manual", "index.json")
		manualIndex := VersionsIndex{
			ID: "manual",
			Versions: []VersionInfo{
				{
					Version:   "3.0.0",
					Protocols: []string{"6.0"},
					Platforms: []Platform{
						{OS: "linux", Arch: "amd64"},
					},
				},
			},
		}

		data, err := json.MarshalIndent(manualIndex, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal test data: %v", err)
		}

		err = os.MkdirAll(filepath.Dir(manualIndexPath), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		err = os.WriteFile(manualIndexPath, data, 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		// Read it back
		index, err := ReadVersionsIndex(manualIndexPath, "manual")
		if err != nil {
			t.Fatalf("ReadVersionsIndex error: %v", err)
		}

		if index.ID != "manual" {
			t.Errorf("Index ID = %s, want 'manual'", index.ID)
		}

		if len(index.Versions) != 1 {
			t.Fatalf("Index has %d versions, want 1", len(index.Versions))
		}

		if index.Versions[0].Version != "3.0.0" {
			t.Errorf("Version = %s, want '3.0.0'", index.Versions[0].Version)
		}
	})
}
