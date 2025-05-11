package file

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Fixed UTC time for zip files: 2049-01-01 00:00:00
var FixedTime = time.Date(2049, 1, 1, 0, 0, 0, 0, time.UTC)

func TestEnsureDir(t *testing.T) {
	// Create a temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "file_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testDir := filepath.Join(tmpDir, "test", "nested", "dir")

	// Test creating a nested directory
	err = EnsureDir(testDir)
	if err != nil {
		t.Fatalf("EnsureDir() error = %v", err)
	}

	// Check if directory exists
	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("Failed to stat directory: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("Created path is not a directory")
	}

	// Test idempotence - should not error when directory already exists
	err = EnsureDir(testDir)
	if err != nil {
		t.Errorf("EnsureDir() failed on existing directory: %v", err)
	}
}

func TestWriteEmptyFile(t *testing.T) {
	// Create a temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "file_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFilePath := filepath.Join(tmpDir, "nested", "dir", "test.txt")
	testComment := "This is a test comment"

	// Test writing an empty file with a comment
	err = WriteEmptyFile(testFilePath, testComment)
	if err != nil {
		t.Fatalf("WriteEmptyFile() error = %v", err)
	}

	// Check if file exists
	info, err := os.Stat(testFilePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.IsDir() {
		t.Errorf("Created path is a directory, expected a file")
	}

	// Check file content
	content, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != testComment {
		t.Errorf("File content = %q, want %q", string(content), testComment)
	}
}

func TestCopyFile(t *testing.T) {
	// Create a temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "file_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "nested", "dest.txt")
	testContent := "This is test content for copy"

	// Create source file
	err = os.WriteFile(srcPath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy file to destination
	err = CopyFile(srcPath, dstPath)
	if err != nil {
		t.Fatalf("CopyFile() error = %v", err)
	}

	// Check if destination file exists
	info, err := os.Stat(dstPath)
	if err != nil {
		t.Fatalf("Failed to stat destination file: %v", err)
	}
	if info.IsDir() {
		t.Errorf("Destination is a directory, expected a file")
	}

	// Check destination content
	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}
	if string(content) != testContent {
		t.Errorf("Destination content = %q, want %q", string(content), testContent)
	}
}

func TestCreateZipFromBinary(t *testing.T) {
	// Create a temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "file_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath := filepath.Join(tmpDir, "terraform-provider-test_v1.0.0_linux_amd64")
	zipPath := filepath.Join(tmpDir, "output", "test.zip")
	testContent := "This is test binary content"

	// Create mock binary file
	err = os.WriteFile(binaryPath, []byte(testContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create binary file: %v", err)
	}

	// Create zip from binary
	err = CreateZipFromBinary(binaryPath, zipPath)
	if err != nil {
		t.Fatalf("CreateZipFromBinary() error = %v", err)
	}

	// Check if zip file exists
	info, err := os.Stat(zipPath)
	if err != nil {
		t.Fatalf("Failed to stat zip file: %v", err)
	}
	if info.IsDir() {
		t.Errorf("Output is a directory, expected a file")
	}

	// Open and inspect the zip file
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("Failed to open zip file: %v", err)
	}
	defer zipReader.Close()

	if len(zipReader.File) != 1 {
		t.Errorf("Zip contains %d files, want 1", len(zipReader.File))
		return
	}

	zipFile := zipReader.File[0]

	// Check zip file name - should be without OS and ARCH
	expectedName := "terraform-provider-test_v1.0.0"
	if zipFile.Name != expectedName {
		t.Errorf("Zip entry name = %q, want %q", zipFile.Name, expectedName)
	}

	// Check file mode
	if zipFile.Mode() != 0755 {
		t.Errorf("Zip entry mode = %v, want %v", zipFile.Mode(), 0755)
	}

	// Check modified time
	if !zipFile.Modified.Equal(FixedTime) {
		t.Errorf("Zip entry time = %v, want %v", zipFile.Modified, FixedTime)
	}

	// Check file content
	rc, err := zipFile.Open()
	if err != nil {
		t.Fatalf("Failed to open zip entry: %v", err)
	}
	defer rc.Close()

	content := make([]byte, len(testContent))
	byteSize, err := rc.Read(content)
	if err == io.EOF {
		if byteSize < len(testContent) {
			t.Fatalf("Read %d bytes, want %d", byteSize, len(testContent))
		}
	} else if err != nil {
		t.Fatalf("Failed to read zip content: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Zip content = %q, want %q", string(content), testContent)
	}
}
