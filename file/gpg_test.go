package file

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
)

// SetupTestGPG creates a temporary GPG key and sets environment variables for tests
func SetupTestGPG(t *testing.T) func() {
	// Save old env vars to restore later
	oldKeyFile := os.Getenv("TFREGBUILDER_GPG_KEY_FILE")
	oldKey := os.Getenv("TFREGBUILDER_GPG_KEY")
	oldPassphrase := os.Getenv("TFREGBUILDER_GPG_PASSPHRASE")
	oldKeyID := os.Getenv("TFREGBUILDER_GPG_ID")

	// Generate a test GPG key
	name := "terraform-registry-builder-test"
	email := "test@example.com"
	passphrase := "testpassphrase"

	// Create a key with RSA 4096 bits using PGP API
	pgp := crypto.PGP()
	keyGen := pgp.KeyGeneration().
		AddUserId(name, email).
		OverrideProfileAlgorithm(crypto.KeyGenerationRSA4096) // Use RSA 4096 instead of 2048 since it's more secure

	keyGeneration := keyGen.New()
	key, err := keyGeneration.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate GPG key: %v", err)
	}

	// Armor the key
	armored, err := key.Armor()
	if err != nil {
		t.Fatalf("Failed to armor GPG key: %v", err)
	}

	// Get key ID (extract from key)
	keyObj, err := crypto.NewKeyFromArmored(armored)
	if err != nil {
		t.Fatalf("Failed to parse key: %v", err)
	}
	fingerprint := keyObj.GetFingerprint()
	keyID := fingerprint[len(fingerprint)-16:] // last 16 chars of fingerprint

	// Set environment variables
	os.Setenv("TFREGBUILDER_GPG_KEY", armored)
	os.Setenv("TFREGBUILDER_GPG_PASSPHRASE", passphrase)
	os.Setenv("TFREGBUILDER_GPG_ID", keyID)

	// Return cleanup function to restore previous environment
	return func() {
		os.Setenv("TFREGBUILDER_GPG_KEY_FILE", oldKeyFile)
		os.Setenv("TFREGBUILDER_GPG_KEY", oldKey)
		os.Setenv("TFREGBUILDER_GPG_PASSPHRASE", oldPassphrase)
		os.Setenv("TFREGBUILDER_GPG_ID", oldKeyID)
	}
}

func TestGPGFunctions(t *testing.T) {
	// Setup test environment
	cleanup := SetupTestGPG(t)
	defer cleanup()

	// Test calculating SHA256 hash
	t.Run("CalculateSHA256", func(t *testing.T) {
		// Create a temporary file
		tmpfile, err := os.CreateTemp("", "hash-test")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpfile.Name())

		content := []byte("hello world")
		if _, err := tmpfile.Write(content); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		if err := tmpfile.Close(); err != nil {
			t.Fatalf("Failed to close temp file: %v", err)
		}

		// Calculate hash
		hash, err := CalculateSHA256(tmpfile.Name())
		if err != nil {
			t.Fatalf("CalculateSHA256 error: %v", err)
		}

		// Expected hash for "hello world"
		expectedHash := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
		if hash != expectedHash {
			t.Errorf("Hash = %s, want %s", hash, expectedHash)
		}
	})

	// Test SHA256SUMS file generation
	t.Run("WriteSHA256SumsFile", func(t *testing.T) {
		// Create a temporary directory
		tmpDir, err := os.MkdirTemp("", "sha-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create a test file
		zipPath := tmpDir + "/test.zip"
		content := []byte("test zip content")
		if err := os.WriteFile(zipPath, content, 0644); err != nil {
			t.Fatalf("Failed to create zip file: %v", err)
		}

		// Generate SHA256SUMS file
		shaPath := tmpDir + "/SHA256SUMS"
		hash, err := WriteSHA256SumsFile(zipPath, shaPath)
		if err != nil {
			t.Fatalf("WriteSHA256SumsFile error: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(shaPath); os.IsNotExist(err) {
			t.Fatalf("SHA256SUMS file not created")
		}

		// Verify file content
		data, err := os.ReadFile(shaPath)
		if err != nil {
			t.Fatalf("Failed to read SHA256SUMS file: %v", err)
		}

		expectedContent := hash + "  test.zip\n"
		if string(data) != expectedContent {
			t.Errorf("SHA256SUMS content = %q, want %q", string(data), expectedContent)
		}
	})

	// Test signing and verifying
	t.Run("SignFile", func(t *testing.T) {
		// Create a temporary directory
		tmpDir, err := os.MkdirTemp("", "sign-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create a test file to sign
		filePath := tmpDir + "/file.txt"
		content := []byte("file to sign")
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		// Sign the file
		sigPath := tmpDir + "/file.txt.sig"
		keyID, err := SignFile(filePath, sigPath)
		if err != nil {
			t.Fatalf("SignFile error: %v", err)
		}

		// Verify key ID matches
		if keyID != os.Getenv("TFREGBUILDER_GPG_ID") {
			t.Errorf("Key ID = %s, want %s", keyID, os.Getenv("TFREGBUILDER_GPG_ID"))
		}

		// Verify signature file exists
		if _, err := os.Stat(sigPath); os.IsNotExist(err) {
			t.Fatalf("Signature file not created")
		}
	})

	// Test public key extraction
	t.Run("GetPublicKey", func(t *testing.T) {
		privateKey := os.Getenv("TFREGBUILDER_GPG_KEY")
		publicKey, err := GetPublicKey(privateKey)
		if err != nil {
			t.Fatalf("GetPublicKey error: %v", err)
		}

		// Verify it's not empty
		if publicKey == "" {
			t.Error("Got empty public key")
		}

		// Verify it's different from private key
		if publicKey == privateKey {
			t.Error("Public key is identical to private key")
		}
	})

	// Test download index generation
	t.Run("WriteDownloadIndex", func(t *testing.T) {
		// Create a temporary directory
		tmpDir, err := os.MkdirTemp("", "index-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create mock paths
		zipPath := tmpDir + "/terraform-provider-test-v1.0.0_linux_amd64.zip"
		shaPath := tmpDir + "/terraform-provider-test-v1.0.0_linux_amd64.zip_SHA256SUMS"
		sigPath := tmpDir + "/terraform-provider-test-v1.0.0_linux_amd64.zip_SHA256SUMS.sig"
		indexPath := tmpDir + "/index.json"

		// Create mock content
		zipContent := []byte("mock zip content")
		shaContent := "abcdef1234567890  terraform-provider-test-v1.0.0_linux_amd64.zip\n"
		sigContent := []byte("mock signature")

		// Write mock files
		if err := os.WriteFile(zipPath, zipContent, 0644); err != nil {
			t.Fatalf("Failed to create zip file: %v", err)
		}
		if err := os.WriteFile(shaPath, []byte(shaContent), 0644); err != nil {
			t.Fatalf("Failed to create SHA file: %v", err)
		}
		if err := os.WriteFile(sigPath, sigContent, 0644); err != nil {
			t.Fatalf("Failed to create sig file: %v", err)
		}

		// Generate index.json
		err = WriteDownloadIndex(zipPath, shaPath, sigPath, indexPath)
		if err != nil {
			t.Fatalf("WriteDownloadIndex error: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			t.Fatalf("index.json file not created")
		}

		// Verify JSON content
		data, err := os.ReadFile(indexPath)
		if err != nil {
			t.Fatalf("Failed to read index.json file: %v", err)
		}

		var index DownloadIndex
		if err := json.Unmarshal(data, &index); err != nil {
			t.Fatalf("Failed to parse index.json: %v", err)
		}

		// Verify basic structure
		if len(index.Protocols) == 0 {
			t.Error("No protocols in index")
		}
		if index.OS != "linux" {
			t.Errorf("OS = %s, want 'linux'", index.OS)
		}
		if index.Arch != "amd64" {
			t.Errorf("Arch = %s, want 'amd64'", index.Arch)
		}
		if len(index.SigningKeys.GPGPublicKeys) == 0 {
			t.Error("No signing keys in index")
		}
	})
}

func TestGetGPGPrivateKeyNoID(t *testing.T) {
	// Setup test environment
	cleanup := SetupTestGPG(t)
	defer cleanup()

	// Get the original key and ID for verification later
	originalKeyID := os.Getenv("TFREGBUILDER_GPG_ID")
	originalKey := os.Getenv("TFREGBUILDER_GPG_KEY")

	// Create a temporary key file
	tmpKeyFile, err := os.CreateTemp("", "gpg-key-*.asc")
	if err != nil {
		t.Fatalf("Failed to create temp key file: %v", err)
	}
	defer os.Remove(tmpKeyFile.Name())

	// Write the key to the file
	if _, err := tmpKeyFile.WriteString(originalKey); err != nil {
		t.Fatalf("Failed to write key to temp file: %v", err)
	}
	if err := tmpKeyFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Test 1: Key file with no ID set
	t.Run("KeyFileWithNoID", func(t *testing.T) {
		// Temporarily unset TFREGBUILDER_GPG_ID and set TFREGBUILDER_GPG_KEY_FILE
		os.Unsetenv("TFREGBUILDER_GPG_ID")
		os.Unsetenv("TFREGBUILDER_GPG_KEY")
		os.Setenv("TFREGBUILDER_GPG_KEY_FILE", tmpKeyFile.Name())

		// Call GetGPGPrivateKey
		_, _, extractedKeyID, err := GetGPGPrivateKey()
		if err != nil {
			t.Fatalf("GetGPGPrivateKey error: %v", err)
		}

		// Verify extracted key ID matches the original
		if extractedKeyID == "" {
			t.Error("Failed to extract key ID from key file")
		}
		if extractedKeyID != originalKeyID {
			t.Errorf("Extracted key ID = %s, want %s", extractedKeyID, originalKeyID)
		}
	})

	// Test 2: Direct key content with no ID set
	t.Run("DirectKeyWithNoID", func(t *testing.T) {
		// Temporarily unset TFREGBUILDER_GPG_ID and use direct key content
		os.Unsetenv("TFREGBUILDER_GPG_ID")
		os.Unsetenv("TFREGBUILDER_GPG_KEY_FILE")
		os.Setenv("TFREGBUILDER_GPG_KEY", originalKey)

		// Call GetGPGPrivateKey
		_, _, extractedKeyID, err := GetGPGPrivateKey()
		if err != nil {
			t.Fatalf("GetGPGPrivateKey error: %v", err)
		}

		// Verify extracted key ID matches the original
		if extractedKeyID == "" {
			t.Error("Failed to extract key ID from key content")
		}
		if extractedKeyID != originalKeyID {
			t.Errorf("Extracted key ID = %s, want %s", extractedKeyID, originalKeyID)
		}
	})
}
