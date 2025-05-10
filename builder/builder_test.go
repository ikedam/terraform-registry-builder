package builder

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/ikedam/terraform-registry-builder/file"
)

func TestMain(m *testing.M) {
	// Setup GPG environment for all tests
	if os.Getenv("TFREGBUILDER_GPG_KEY") == "" {
		keyName := "terraform-registry-builder-test"
		email := "test@example.com"
		passphrase := "testpassphrase"

		// Initialize PGP
		pgp := crypto.PGP()

		// Create a key generation handle
		keyGenHandle := pgp.KeyGeneration().
			AddUserId(keyName, email).
			New()

		// Generate key
		key, err := keyGenHandle.GenerateKey()
		if err != nil {
			panic(err)
		}

		// Get the armored version of the key
		armored, err := key.Armor()
		if err != nil {
			panic(err)
		}

		// Get key ID from fingerprint
		fingerprint := key.GetFingerprint()
		keyID := fingerprint[len(fingerprint)-16:] // last 16 chars of fingerprint

		// Set environment variables
		os.Setenv("TFREGBUILDER_GPG_KEY", armored)
		os.Setenv("TFREGBUILDER_GPG_PASSPHRASE", passphrase)
		os.Setenv("TFREGBUILDER_GPG_ID", keyID)
	}

	// Run tests
	os.Exit(m.Run())
}

func TestBuilder(t *testing.T) {
	// Create temporary source and destination directories for tests
	srcDir, err := os.MkdirTemp("", "builder_test_src")
	if err != nil {
		t.Fatalf("Failed to create temporary source directory: %v", err)
	}
	defer os.RemoveAll(srcDir)

	dstDir, err := os.MkdirTemp("", "builder_test_dst")
	if err != nil {
		t.Fatalf("Failed to create temporary destination directory: %v", err)
	}
	defer os.RemoveAll(dstDir)

	// Test cases for provider files
	testCases := []struct {
		name     string
		fileName string
		content  string
		isZip    bool
	}{
		{
			name:     "binary provider",
			fileName: "terraform-provider-test-v1.0.0_linux_amd64",
			content:  "mock binary content",
			isZip:    false,
		},
		{
			name:     "zip provider",
			fileName: "terraform-provider-example-v2.0.0_darwin_arm64.zip",
			content:  "mock zip content",
			isZip:    true,
		},
	}

	// Create test files in source directory
	for _, tc := range testCases {
		filePath := filepath.Join(srcDir, tc.fileName)
		err = os.WriteFile(filePath, []byte(tc.content), 0755)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", tc.fileName, err)
		}
	}

	// Create a nested directory with another provider
	nestedDir := filepath.Join(srcDir, "nested", "dir")
	err = os.MkdirAll(nestedDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	nestedFile := filepath.Join(nestedDir, "terraform-provider-nested-v3.0.0_windows_386")
	err = os.WriteFile(nestedFile, []byte("nested provider content"), 0755)
	if err != nil {
		t.Fatalf("Failed to create nested test file: %v", err)
	}

	// Run the builder
	b := New(srcDir, dstDir)
	err = b.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Check expected files in destination
	expectedFiles := []string{
		// For test provider
		filepath.Join(dstDir, "test", "versions", "index.json"),
		filepath.Join(dstDir, "test", "1.0.0", "download", "linux", "amd64", "index.json"),
		filepath.Join(dstDir, "test", "1.0.0", "download", "linux", "amd64", "terraform-provider-test-v1.0.0_linux_amd64.zip"),
		filepath.Join(dstDir, "test", "1.0.0", "download", "linux", "amd64", "terraform-provider-test-v1.0.0_linux_amd64.zip_SHA256SUMS"),
		filepath.Join(dstDir, "test", "1.0.0", "download", "linux", "amd64", "terraform-provider-test-v1.0.0_linux_amd64.zip_SHA256SUMS.sig"),

		// For example provider
		filepath.Join(dstDir, "example", "versions", "index.json"),
		filepath.Join(dstDir, "example", "2.0.0", "download", "darwin", "arm64", "index.json"),
		filepath.Join(dstDir, "example", "2.0.0", "download", "darwin", "arm64", "terraform-provider-example-v2.0.0_darwin_arm64.zip"),
		filepath.Join(dstDir, "example", "2.0.0", "download", "darwin", "arm64", "terraform-provider-example-v2.0.0_darwin_arm64.zip_SHA256SUMS"),
		filepath.Join(dstDir, "example", "2.0.0", "download", "darwin", "arm64", "terraform-provider-example-v2.0.0_darwin_arm64.zip_SHA256SUMS.sig"),

		// For nested provider
		filepath.Join(dstDir, "nested", "versions", "index.json"),
		filepath.Join(dstDir, "nested", "3.0.0", "download", "windows", "386", "index.json"),
		filepath.Join(dstDir, "nested", "3.0.0", "download", "windows", "386", "terraform-provider-nested-v3.0.0_windows_386.zip"),
		filepath.Join(dstDir, "nested", "3.0.0", "download", "windows", "386", "terraform-provider-nested-v3.0.0_windows_386.zip_SHA256SUMS"),
		filepath.Join(dstDir, "nested", "3.0.0", "download", "windows", "386", "terraform-provider-nested-v3.0.0_windows_386.zip_SHA256SUMS.sig"),
	}

	for _, expectedFile := range expectedFiles {
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Errorf("Expected file not created: %s", expectedFile)
		}
	}

	// Verify versions index.json content
	testVersionsIndexContent := func(t *testing.T, providerType string, expectedVersions []string, expectedOS, expectedArch string) {
		indexPath := filepath.Join(dstDir, providerType, "versions", "index.json")
		data, err := os.ReadFile(indexPath)
		if err != nil {
			t.Fatalf("Failed to read versions index file %s: %v", indexPath, err)
		}

		var index file.VersionsIndex
		err = json.Unmarshal(data, &index)
		if err != nil {
			t.Fatalf("Failed to parse versions index file %s: %v", indexPath, err)
		}

		// Verify ID
		if index.ID != providerType {
			t.Errorf("Index ID = %s, want %s", index.ID, providerType)
		}

		// Verify versions
		if len(index.Versions) != len(expectedVersions) {
			t.Errorf("Index contains %d versions, want %d", len(index.Versions), len(expectedVersions))
		}

		for i, expected := range expectedVersions {
			if i >= len(index.Versions) {
				t.Errorf("Missing expected version %s", expected)
				continue
			}

			if index.Versions[i].Version != expected {
				t.Errorf("Version at index %d = %s, want %s", i, index.Versions[i].Version, expected)
			}

			// Verify protocols
			if len(index.Versions[i].Protocols) == 0 {
				t.Errorf("No protocols defined for version %s", expected)
			}

			// Verify platforms
			foundPlatform := false
			for _, platform := range index.Versions[i].Platforms {
				if platform.OS == expectedOS && platform.Arch == expectedArch {
					foundPlatform = true
					break
				}
			}

			if !foundPlatform {
				t.Errorf("Platform %s_%s not found for version %s", expectedOS, expectedArch, expected)
			}
		}
	}

	// Test each provider's versions index
	t.Run("test provider versions", func(t *testing.T) {
		testVersionsIndexContent(t, "test", []string{"1.0.0"}, "linux", "amd64")
	})

	t.Run("example provider versions", func(t *testing.T) {
		testVersionsIndexContent(t, "example", []string{"2.0.0"}, "darwin", "arm64")
	})

	t.Run("nested provider versions", func(t *testing.T) {
		testVersionsIndexContent(t, "nested", []string{"3.0.0"}, "windows", "386")
	})
}
