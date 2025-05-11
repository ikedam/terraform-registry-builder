// Package file provides utilities for file operations needed by the Terraform registry builder.
package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// VersionsIndex represents the structure of the versions index.json file.
type VersionsIndex struct {
	ID       string        `json:"id"`
	Versions []VersionInfo `json:"versions"`
	Warnings []string      `json:"warnings,omitempty"`
}

// VersionInfo represents a single version entry in the versions index.
type VersionInfo struct {
	Version   string     `json:"version"`
	Protocols []string   `json:"protocols"`
	Platforms []Platform `json:"platforms"`
}

// Platform represents a platform entry in version info.
type Platform struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

// ReadVersionsIndex reads a versions index.json file if it exists, otherwise returns a new empty one.
func ReadVersionsIndex(path string, id string) (*VersionsIndex, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Return a new empty index
		return &VersionsIndex{
			ID:       id,
			Versions: []VersionInfo{},
		}, nil
	}

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read versions index file: %w", err)
	}

	// Parse the JSON
	var index VersionsIndex
	if len(data) > 0 {
		err = json.Unmarshal(data, &index)
		if err != nil {
			return nil, fmt.Errorf("failed to parse versions index file: %w", err)
		}
	} else {
		// File exists but is empty
		index = VersionsIndex{
			ID:       id,
			Versions: []VersionInfo{},
		}
	}

	return &index, nil
}

// AddVersion adds or updates a version in the index.
func (vi *VersionsIndex) AddVersion(version, os, arch string) {
	// Check if this version already exists
	var existingVersion *VersionInfo
	for i := range vi.Versions {
		if vi.Versions[i].Version == version {
			existingVersion = &vi.Versions[i]
			break
		}
	}

	// If version doesn't exist, create it
	if existingVersion == nil {
		vi.Versions = append(vi.Versions, VersionInfo{
			Version:   version,
			Protocols: []string{"6.0"},
			Platforms: []Platform{
				{
					OS:   os,
					Arch: arch,
				},
			},
		})
	} else {
		// Check if this platform already exists for the version
		platformExists := false
		for _, platform := range existingVersion.Platforms {
			if platform.OS == os && platform.Arch == arch {
				platformExists = true
				break
			}
		}

		// Add platform if it doesn't exist
		if !platformExists {
			existingVersion.Platforms = append(existingVersion.Platforms, Platform{
				OS:   os,
				Arch: arch,
			})
		}
	}

	// Sort versions in descending order (newest first)
	sort.Slice(vi.Versions, func(i, j int) bool {
		return vi.Versions[i].Version > vi.Versions[j].Version
	})
}

// WriteVersionsIndex writes the versions index to a file.
func WriteVersionsIndex(path string, index *VersionsIndex) error {
	// Ensure directory exists
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return fmt.Errorf("failed to create directory for versions index: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal versions index: %w", err)
	}

	// Write to file
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write versions index: %w", err)
	}

	return nil
}
