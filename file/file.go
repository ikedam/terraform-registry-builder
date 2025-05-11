// Package file provides utilities for file operations needed by the Terraform registry builder.
package file

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ikedam/terraform-registry-builder/internal/provider"
)

// EnsureDir ensures that a directory exists, creating it if necessary.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Ensure destination directory exists
	if err := EnsureDir(filepath.Dir(dst)); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy content
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy content: %w", err)
	}

	return nil
}

// CreateZipFromBinary creates a zip file containing a single binary with fixed mode and time.
func CreateZipFromBinary(binaryPath, zipPath string) error {
	// Create parent directory if it doesn't exist
	if err := EnsureDir(filepath.Dir(zipPath)); err != nil {
		return fmt.Errorf("failed to create directory for zip: %w", err)
	}

	// Create a new zip file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	// Create a new zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Open the binary file
	binaryFile, err := os.Open(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to open binary file: %w", err)
	}
	defer binaryFile.Close()

	// Extract provider information from binary path to create the correct inner file name
	info, err := provider.ParseProviderFileName(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to parse provider file name: %w", err)
	}

	// Create a zip header for the binary, using just TYPE and VERSION
	header := &zip.FileHeader{
		Name:   info.InnerZipFileName(),
		Method: zip.Deflate,
	}
	// Set zero time
	// nolint: staticcheck
	header.SetModTime(time.Time{})
	header.SetMode(0755)

	// Add the binary to the zip
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create zip header: %w", err)
	}

	// Copy the binary into the zip
	_, err = io.Copy(writer, binaryFile)
	if err != nil {
		return fmt.Errorf("failed to write binary to zip: %w", err)
	}

	return nil
}

// WriteEmptyFile creates an empty file at the given path.
func WriteEmptyFile(path string, comment string) error {
	// Ensure directory exists
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return fmt.Errorf("failed to create directory for file: %w", err)
	}

	// Create and write to file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if comment != "" {
		_, err = file.WriteString(comment)
		if err != nil {
			return fmt.Errorf("failed to write comment to file: %w", err)
		}
	}

	return nil
}
