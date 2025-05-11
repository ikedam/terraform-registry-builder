// Package file provides utilities for file operations needed by the Terraform registry builder.
package file

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
)

// DownloadIndex represents the structure of the download index.json file.
type DownloadIndex struct {
	Protocols           []string          `json:"protocols"`
	OS                  string            `json:"os"`
	Arch                string            `json:"arch"`
	Filename            string            `json:"filename"`
	DownloadURL         string            `json:"download_url"`
	ShasumsURL          string            `json:"shasums_url"`
	ShasumsSignatureURL string            `json:"shasums_signature_url"`
	Shasum              string            `json:"shasum"`
	SigningKeys         SigningKeysObject `json:"signing_keys"`
}

// SigningKeysObject represents the signing keys object in the download index.json file.
type SigningKeysObject struct {
	GPGPublicKeys []GPGPublicKey `json:"gpg_public_keys"`
}

// GPGPublicKey represents a GPG public key in the signing keys object.
type GPGPublicKey struct {
	KeyID      string `json:"key_id"`
	ASCIIArmor string `json:"ascii_armor"`
}

// CalculateSHA256 calculates the SHA256 hash of a file.
func CalculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for hashing: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	hashBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	return hashString, nil
}

// WriteSHA256SumsFile writes the SHA256SUMS file with the hash of the zip file.
func WriteSHA256SumsFile(zipFilePath, shaSumsPath string) (string, error) {
	// Calculate SHA256 hash
	hash, err := CalculateSHA256(zipFilePath)
	if err != nil {
		return "", err
	}

	// Format content: hash + two spaces + filename
	zipFileName := filepath.Base(zipFilePath)
	content := fmt.Sprintf("%s  %s\n", hash, zipFileName)

	// Write to file
	if err := os.WriteFile(shaSumsPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write SHA256SUMS file: %w", err)
	}

	return hash, nil
}

// GetGPGPrivateKey gets the GPG private key from environment variables.
func GetGPGPrivateKey() (string, string, string, error) {
	// Get key ID and private key
	var keyID, privateKey string

	// Get passphrase
	passphrase := os.Getenv("TFREGBUILDER_GPG_PASSPHRASE")

	// Get private key from file or direct content
	keyFile := os.Getenv("TFREGBUILDER_GPG_KEY_FILE")
	if keyFile != "" {
		data, err := os.ReadFile(keyFile)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to read GPG key file: %w", err)
		}
		privateKey = string(data)

		// Extract key ID from the private key if not specified in environment
		keyID = os.Getenv("TFREGBUILDER_GPG_ID")
		if keyID == "" {
			// Parse the key to extract the key ID
			key, err := crypto.NewKeyFromArmored(privateKey)
			if err != nil {
				return "", "", "", fmt.Errorf("failed to parse private key from file: %w", err)
			}
			// Extract fingerprint
			fingerprint := key.GetFingerprint()
			if fingerprint == "" {
				return "", "", "", fmt.Errorf("could not extract key ID from key file")
			}
			// Use last 16 characters of fingerprint as key ID (standard format)
			if len(fingerprint) >= 16 {
				keyID = fingerprint[len(fingerprint)-16:]
			} else {
				keyID = fingerprint // Fallback to using the whole fingerprint
			}
		}
	} else {
		privateKey = os.Getenv("TFREGBUILDER_GPG_KEY")
		if privateKey == "" {
			return "", "", "", fmt.Errorf("either TFREGBUILDER_GPG_KEY_FILE or TFREGBUILDER_GPG_KEY must be set")
		}

		// Get key ID from environment when using direct key content
		keyID = os.Getenv("TFREGBUILDER_GPG_ID")
		if keyID == "" {
			// Also try to extract from direct key content
			key, err := crypto.NewKeyFromArmored(privateKey)
			if err != nil {
				return "", "", "", fmt.Errorf("TFREGBUILDER_GPG_ID environment variable not set and could not extract key ID from key content")
			}
			fingerprint := key.GetFingerprint()
			if fingerprint == "" {
				return "", "", "", fmt.Errorf("TFREGBUILDER_GPG_ID environment variable not set and could not extract key ID")
			}
			if len(fingerprint) >= 16 {
				keyID = fingerprint[len(fingerprint)-16:]
			} else {
				keyID = fingerprint
			}
		}
	}

	return privateKey, passphrase, keyID, nil
}

// SignFile signs a file using GPG.
func SignFile(filePath, signaturePath string) (string, error) {
	// Get GPG key information
	privateKeyArmored, passphrase, keyID, err := GetGPGPrivateKey()
	if err != nil {
		return "", err
	}

	// Read the file to sign
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file to sign: %w", err)
	}

	// Parse the private key
	key, err := crypto.NewKeyFromArmored(privateKeyArmored)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// Unlock the key with passphrase if provided
	if passphrase != "" {
		isLocked, err := key.IsLocked()
		if err != nil {
			return "", fmt.Errorf("failed to check if key is locked: %w", err)
		}

		if isLocked {
			_, err := key.Unlock([]byte(passphrase))
			if err != nil {
				return "", fmt.Errorf("failed to unlock private key: %w", err)
			}
		}
	}

	// Initialize PGP
	pgp := crypto.PGP()

	// Create a signer
	signer, err := pgp.Sign().SigningKey(key).Detached().New()
	if err != nil {
		return "", fmt.Errorf("failed to create signer: %w", err)
	}
	defer signer.ClearPrivateParams()

	// Sign the data (armor=false for binary output)
	signature, err := signer.Sign(fileData, crypto.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to sign file: %w", err)
	}

	// Write signature to file
	if err := os.WriteFile(signaturePath, signature, 0644); err != nil {
		return "", fmt.Errorf("failed to write signature file: %w", err)
	}

	return keyID, nil
}

// GetPublicKey extracts the public key from a private key.
func GetPublicKey(privateKeyArmored string) (string, error) {
	privateKeyObj, err := crypto.NewKeyFromArmored(privateKeyArmored)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// Convert to public key
	publicKey, err := privateKeyObj.ToPublic()
	if err != nil {
		return "", fmt.Errorf("failed to extract public key: %w", err)
	}

	// Armor the public key
	armoredPublicKey, err := publicKey.Armor()
	if err != nil {
		return "", fmt.Errorf("failed to armor public key: %w", err)
	}

	return armoredPublicKey, nil
}

// WriteDownloadIndex creates the download index.json file.
func WriteDownloadIndex(zipPath, shasumsPath, sigPath, downloadIndexPath string) error {
	// Extract relevant information from paths
	zipFileName := filepath.Base(zipPath)
	shasumsFileName := filepath.Base(shasumsPath)
	sigFileName := filepath.Base(sigPath)

	// Get provider info from zip file name
	parts := strings.Split(zipFileName, "_")
	if len(parts) < 2 {
		return fmt.Errorf("invalid zip file name format: %s", zipFileName)
	}
	osPart := parts[len(parts)-2]
	archPart := strings.TrimSuffix(parts[len(parts)-1], ".zip")

	// Calculate SHA256 hash of zip file
	shasum, err := CalculateSHA256(zipPath)
	if err != nil {
		return fmt.Errorf("failed to calculate SHA256 hash: %w", err)
	}

	// Get GPG key information
	privateKey, _, keyID, err := GetGPGPrivateKey()
	if err != nil {
		return err
	}

	// Get public key
	publicKey, err := GetPublicKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	// Create download index
	index := DownloadIndex{
		Protocols:           []string{"6.0"},
		OS:                  osPart,
		Arch:                archPart,
		Filename:            zipFileName,
		DownloadURL:         zipFileName,
		ShasumsURL:          shasumsFileName,
		ShasumsSignatureURL: sigFileName,
		Shasum:              shasum,
		SigningKeys: SigningKeysObject{
			GPGPublicKeys: []GPGPublicKey{
				{
					KeyID:      keyID,
					ASCIIArmor: publicKey,
				},
			},
		},
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal download index: %w", err)
	}

	// Write to file
	if err := os.WriteFile(downloadIndexPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write download index: %w", err)
	}

	return nil
}
