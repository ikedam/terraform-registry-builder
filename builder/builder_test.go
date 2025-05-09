package builder

import (
	"os"
	"path/filepath"
	"testing"
)

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
}
