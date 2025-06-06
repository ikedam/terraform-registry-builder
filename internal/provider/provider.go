// Package provider provides types and utilities for parsing and managing Terraform provider information.
package provider

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ProviderInfo represents the parsed information from a provider file name.
type ProviderInfo struct {
	Type    string // Provider type, e.g., "aws"
	Version string // Provider version, e.g., "0.1.0"
	OS      string // Operating system, e.g., "linux"
	Arch    string // Architecture, e.g., "amd64"
	Ext     string // Extension for the binary (e.g., ".exe" for Windows)
}

var (
	// Regular expression to match provider file names.
	// Format: terraform-provider-(TYPE)_v(VERSION)_(OS)_(ARCH)[.exe] or
	// terraform-provider-(TYPE)_v(VERSION)_(OS)_(ARCH).zip
	// Note: .exe.zip is not allowed
	providerRegex = regexp.MustCompile(`^terraform-provider-([^-]+)_v([^_]+)_([^_]+)_([^.]+)(?:(\.exe)$|(?:\.zip)$)?$`)
)

// ParseProviderFileName parses a provider file name and returns the provider information.
func ParseProviderFileName(filename string) (*ProviderInfo, error) {
	// Extract just the base name
	baseName := filepath.Base(filename)

	// Match against the regex
	matches := providerRegex.FindStringSubmatch(baseName)
	if matches == nil || len(matches) != 6 {
		return nil, fmt.Errorf("invalid provider file name format: %s", filename)
	}

	// Extract the components
	ext := ""
	if matches[5] != "" {
		ext = matches[5] // This will be ".exe" or empty string
	}

	return &ProviderInfo{
		Type:    matches[1],
		Version: matches[2],
		OS:      matches[3],
		Arch:    matches[4],
		Ext:     ext,
	}, nil
}

// TargetBasePath returns the base path for this provider in the registry structure.
func (p *ProviderInfo) TargetBasePath() string {
	return p.Type
}

// TargetVersionPath returns the version-specific path for this provider in the registry structure.
func (p *ProviderInfo) TargetVersionPath() string {
	return filepath.Join(p.Type, p.Version)
}

// TargetDownloadPath returns the download path for this provider in the registry structure.
func (p *ProviderInfo) TargetDownloadPath() string {
	return filepath.Join(p.Type, p.Version, "download", p.OS, p.Arch)
}

// TargetVersionsIndexPath returns the path to the versions index file.
func (p *ProviderInfo) TargetVersionsIndexPath() string {
	return filepath.Join(p.Type, "versions", "index.json")
}

// TargetDownloadIndexPath returns the path to the download index file.
func (p *ProviderInfo) TargetDownloadIndexPath() string {
	return filepath.Join(p.TargetDownloadPath(), "index.json")
}

// TargetZipFileName returns the name of the target zip file.
func (p *ProviderInfo) TargetZipFileName() string {
	return fmt.Sprintf("terraform-provider-%s_v%s_%s_%s.zip", p.Type, p.Version, p.OS, p.Arch)
}

// TargetZipPath returns the full path to the target zip file.
func (p *ProviderInfo) TargetZipPath() string {
	return filepath.Join(p.TargetDownloadPath(), p.TargetZipFileName())
}

// TargetSHASumsFileName returns the name of the SHA sums file.
func (p *ProviderInfo) TargetSHASumsFileName() string {
	return fmt.Sprintf("terraform-provider-%s_v%s_%s_%s_SHA256SUMS", p.Type, p.Version, p.OS, p.Arch)
}

// TargetSHASumsPath returns the full path to the SHA sums file.
func (p *ProviderInfo) TargetSHASumsPath() string {
	return filepath.Join(p.TargetDownloadPath(), p.TargetSHASumsFileName())
}

// TargetSigFileName returns the name of the signature file.
func (p *ProviderInfo) TargetSigFileName() string {
	return p.TargetSHASumsFileName() + ".sig"
}

// TargetSigPath returns the full path to the signature file.
func (p *ProviderInfo) TargetSigPath() string {
	return filepath.Join(p.TargetDownloadPath(), p.TargetSigFileName())
}

// IsZipFile returns whether the original file is a zip file.
func (p *ProviderInfo) IsZipFile(filename string) bool {
	return strings.HasSuffix(filename, ".zip")
}

// InnerZipFileName returns the file name to be used inside the zip file,
// without OS and ARCH information.
func (p *ProviderInfo) InnerZipFileName() string {
	return fmt.Sprintf("terraform-provider-%s_v%s%s", p.Type, p.Version, p.Ext)
}
