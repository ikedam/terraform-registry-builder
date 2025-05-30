package provider

import (
	"path/filepath"
	"testing"
)

func TestParseProviderFileName(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		wantType    string
		wantVersion string
		wantOS      string
		wantArch    string
		wantExt     string
		wantErr     bool
	}{
		{
			name:        "valid binary filename",
			filename:    "terraform-provider-aws_v1.2.3_linux_amd64",
			wantType:    "aws",
			wantVersion: "1.2.3",
			wantOS:      "linux",
			wantArch:    "amd64",
			wantExt:     "",
			wantErr:     false,
		},
		{
			name:        "valid zip filename",
			filename:    "terraform-provider-google_v2.0.0_darwin_arm64.zip",
			wantType:    "google",
			wantVersion: "2.0.0",
			wantOS:      "darwin",
			wantArch:    "arm64",
			wantExt:     "",
			wantErr:     false,
		},
		{
			name:        "windows binary filename with exe",
			filename:    "terraform-provider-azure_v1.0.0_windows_amd64.exe",
			wantType:    "azure",
			wantVersion: "1.0.0",
			wantOS:      "windows",
			wantArch:    "amd64",
			wantExt:     ".exe",
			wantErr:     false,
		},
		{
			name:     "windows binary filename with exe in zip - should error",
			filename: "terraform-provider-azure_v1.0.0_windows_amd64.exe.zip",
			wantErr:  true,
		},
		{
			name:        "filename with path",
			filename:    "/some/path/to/terraform-provider-azure_v0.12.0_windows_386.zip",
			wantType:    "azure",
			wantVersion: "0.12.0",
			wantOS:      "windows",
			wantArch:    "386",
			wantErr:     false,
		},
		{
			name:     "invalid filename format",
			filename: "not-a-provider-file",
			wantErr:  true,
		},
		{
			name:     "invalid provider name format",
			filename: "terraform-provider.zip",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseProviderFileName(tt.filename)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseProviderFileName() error = nil, wantErr = true")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseProviderFileName() error = %v, wantErr = false", err)
				return
			}

			if got.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", got.Type, tt.wantType)
			}

			if got.Version != tt.wantVersion {
				t.Errorf("Version = %v, want %v", got.Version, tt.wantVersion)
			}

			if got.OS != tt.wantOS {
				t.Errorf("OS = %v, want %v", got.OS, tt.wantOS)
			}

			if got.Arch != tt.wantArch {
				t.Errorf("Arch = %v, want %v", got.Arch, tt.wantArch)
			}

			if got.Ext != tt.wantExt {
				t.Errorf("Ext = %v, want %v", got.Ext, tt.wantExt)
			}
		})
	}
}

func TestProviderInfo_Paths(t *testing.T) {
	info := ProviderInfo{
		Type:    "example",
		Version: "1.0.0",
		OS:      "linux",
		Arch:    "amd64",
	}

	// Test each path generation method
	t.Run("TargetBasePath", func(t *testing.T) {
		expected := "example"
		if got := info.TargetBasePath(); got != expected {
			t.Errorf("TargetBasePath() = %v, want %v", got, expected)
		}
	})

	t.Run("TargetVersionPath", func(t *testing.T) {
		expected := filepath.Join("example", "1.0.0")
		if got := info.TargetVersionPath(); got != expected {
			t.Errorf("TargetVersionPath() = %v, want %v", got, expected)
		}
	})

	t.Run("TargetDownloadPath", func(t *testing.T) {
		expected := filepath.Join("example", "1.0.0", "download", "linux", "amd64")
		if got := info.TargetDownloadPath(); got != expected {
			t.Errorf("TargetDownloadPath() = %v, want %v", got, expected)
		}
	})

	t.Run("TargetVersionsIndexPath", func(t *testing.T) {
		expected := filepath.Join("example", "versions", "index.json")
		if got := info.TargetVersionsIndexPath(); got != expected {
			t.Errorf("TargetVersionsIndexPath() = %v, want %v", got, expected)
		}
	})

	t.Run("TargetDownloadIndexPath", func(t *testing.T) {
		expected := filepath.Join("example", "1.0.0", "download", "linux", "amd64", "index.json")
		if got := info.TargetDownloadIndexPath(); got != expected {
			t.Errorf("TargetDownloadIndexPath() = %v, want %v", got, expected)
		}
	})

	t.Run("TargetZipFileName", func(t *testing.T) {
		expected := "terraform-provider-example_v1.0.0_linux_amd64.zip"
		if got := info.TargetZipFileName(); got != expected {
			t.Errorf("TargetZipFileName() = %v, want %v", got, expected)
		}
	})

	t.Run("TargetZipPath", func(t *testing.T) {
		expected := filepath.Join("example", "1.0.0", "download", "linux", "amd64", "terraform-provider-example_v1.0.0_linux_amd64.zip")
		if got := info.TargetZipPath(); got != expected {
			t.Errorf("TargetZipPath() = %v, want %v", got, expected)
		}
	})

	t.Run("TargetSHASumsFileName", func(t *testing.T) {
		expected := "terraform-provider-example_v1.0.0_linux_amd64_SHA256SUMS"
		if got := info.TargetSHASumsFileName(); got != expected {
			t.Errorf("TargetSHASumsFileName() = %v, want %v", got, expected)
		}
	})

	t.Run("TargetSHASumsPath", func(t *testing.T) {
		expected := filepath.Join("example", "1.0.0", "download", "linux", "amd64", "terraform-provider-example_v1.0.0_linux_amd64_SHA256SUMS")
		if got := info.TargetSHASumsPath(); got != expected {
			t.Errorf("TargetSHASumsPath() = %v, want %v", got, expected)
		}
	})

	t.Run("TargetSigFileName", func(t *testing.T) {
		expected := "terraform-provider-example_v1.0.0_linux_amd64_SHA256SUMS.sig"
		if got := info.TargetSigFileName(); got != expected {
			t.Errorf("TargetSigFileName() = %v, want %v", got, expected)
		}
	})

	t.Run("TargetSigPath", func(t *testing.T) {
		expected := filepath.Join("example", "1.0.0", "download", "linux", "amd64", "terraform-provider-example_v1.0.0_linux_amd64_SHA256SUMS.sig")
		if got := info.TargetSigPath(); got != expected {
			t.Errorf("TargetSigPath() = %v, want %v", got, expected)
		}
	})

	t.Run("IsZipFile", func(t *testing.T) {
		tests := []struct {
			filename string
			want     bool
		}{
			{"terraform-provider-example_v1.0.0_linux_amd64.zip", true},
			{"terraform-provider-example_v1.0.0_linux_amd64", false},
		}

		for _, tt := range tests {
			if got := info.IsZipFile(tt.filename); got != tt.want {
				t.Errorf("IsZipFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		}
	})

	t.Run("InnerZipFileName", func(t *testing.T) {
		// Test without .exe extension
		expected := "terraform-provider-example_v1.0.0"
		if got := info.InnerZipFileName(); got != expected {
			t.Errorf("InnerZipFileName() = %v, want %v", got, expected)
		}

		// Test with .exe extension
		infoWithExt := ProviderInfo{
			Type:    "example",
			Version: "1.0.0",
			OS:      "windows",
			Arch:    "amd64",
			Ext:     ".exe",
		}
		expectedWithExt := "terraform-provider-example_v1.0.0.exe"
		if got := infoWithExt.InnerZipFileName(); got != expectedWithExt {
			t.Errorf("InnerZipFileName() with .exe = %v, want %v", got, expectedWithExt)
		}
	})
}
