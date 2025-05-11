// Package builder provides the main functionality for building a Terraform registry structure.
package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ikedam/terraform-registry-builder/file"
	"github.com/ikedam/terraform-registry-builder/internal/provider"
)

// Builder is responsible for building a Terraform registry structure.
type Builder struct {
	srcDir string
	dstDir string
}

// New creates a new Builder instance.
func New(srcDir, dstDir string) *Builder {
	return &Builder{
		srcDir: srcDir,
		dstDir: dstDir,
	}
}

// Build processes the source directory and builds the registry structure in the destination directory.
func (b *Builder) Build() error {
	// Check if source directory exists
	srcInfo, err := os.Stat(b.srcDir)
	if err != nil {
		return fmt.Errorf("source directory error: %w", err)
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("source path is not a directory")
	}

	// Ensure destination directory exists
	err = file.EnsureDir(b.dstDir)
	if err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Find and process provider files
	return b.processDirectory(b.srcDir)
}

// processDirectory walks through the directory and processes provider files.
func (b *Builder) processDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			// Recursively process subdirectories
			if err := b.processDirectory(path); err != nil {
				return err
			}
		} else {
			// Process files matching the provider pattern
			if strings.HasPrefix(entry.Name(), "terraform-provider-") {
				if err := b.processProviderFile(path); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// processProviderFile processes a single provider file.
func (b *Builder) processProviderFile(filePath string) error {
	// Parse provider information from file name
	info, err := provider.ParseProviderFileName(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse provider file name %s: %w", filePath, err)
	}

	// First, check if this version/platform already exists in the index
	versionsIndexPath := filepath.Join(b.dstDir, info.TargetVersionsIndexPath())
	versionsIndex, err := file.ReadVersionsIndex(versionsIndexPath, info.Type)
	if err != nil {
		return fmt.Errorf("failed to read versions index file: %w", err)
	}

	// Check if the version/platform already exists before adding it
	needsAdding := true
	for _, ver := range versionsIndex.Versions {
		if ver.Version == info.Version {
			for _, plat := range ver.Platforms {
				if plat.OS == info.OS && plat.Arch == info.Arch {
					needsAdding = false
					break
				}
			}
			break
		}
	}

	if !needsAdding {
		fmt.Printf("Skipped %s version %s for %s/%s (already in index)\n", info.Type, info.Version, info.OS, info.Arch)
		return nil // Skip further processing for this file
	}

	fmt.Printf("Adding %s version %s for %s/%s to index\n", info.Type, info.Version, info.OS, info.Arch)

	// Create target directories
	targetPath := filepath.Join(b.dstDir, info.TargetDownloadPath())
	if err = file.EnsureDir(targetPath); err != nil {
		return fmt.Errorf("failed to create target directory %s: %w", targetPath, err)
	}

	// Create versions directory
	versionsDir := filepath.Join(b.dstDir, info.Type, "versions")
	if err = file.EnsureDir(versionsDir); err != nil {
		return fmt.Errorf("failed to create versions directory %s: %w", versionsDir, err)
	}

	// Define target paths
	targetZipPath := filepath.Join(b.dstDir, info.TargetZipPath())

	// Process file based on its type
	if info.IsZipFile(filePath) {
		// Copy zip file directly
		if err = file.CopyFile(filePath, targetZipPath); err != nil {
			return fmt.Errorf("failed to copy zip file: %w", err)
		}
	} else {
		// Create zip from binary
		if err = file.CreateZipFromBinary(filePath, targetZipPath); err != nil {
			return fmt.Errorf("failed to create zip from binary: %w", err)
		}
	}

	// Now add the version/platform to the index and write it
	versionsIndex.AddVersion(info.Version, info.OS, info.Arch)
	if err = file.WriteVersionsIndex(versionsIndexPath, versionsIndex); err != nil {
		return fmt.Errorf("failed to write versions index file: %w", err)
	}

	// Create SHA256SUMS file
	shaSumsPath := filepath.Join(b.dstDir, info.TargetSHASumsPath())
	_, err = file.WriteSHA256SumsFile(targetZipPath, shaSumsPath)
	if err != nil {
		return fmt.Errorf("failed to create SHA sums file: %w", err)
	}

	// Sign SHA256SUMS file
	sigPath := filepath.Join(b.dstDir, info.TargetSigPath())
	_, err = file.SignFile(shaSumsPath, sigPath)
	if err != nil {
		return fmt.Errorf("failed to create signature file: %w", err)
	}

	// Create index.json (download)
	downloadIndexPath := filepath.Join(b.dstDir, info.TargetDownloadIndexPath())
	if err = file.WriteDownloadIndex(targetZipPath, shaSumsPath, sigPath, downloadIndexPath); err != nil {
		return fmt.Errorf("failed to create download index file: %w", err)
	}

	return nil
}
