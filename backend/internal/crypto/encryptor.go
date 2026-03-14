// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

// Package crypto provides encryption primitives for sensitive data at rest.
package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// Encryptor encrypts and decrypts opaque byte slices.
//
// additionalData is authenticated but not encrypted (GCM additional data /
// AAD). Pass the owning record's user_id bytes so that the ciphertext is
// cryptographically bound to that user: decrypting a row with a different
// user_id will fail even if the key is correct. Pass nil when no binding is
// required.
//
// The context is reserved for KMS/Vault implementations that may need it for
// remote calls.
type Encryptor interface {
	Encrypt(ctx context.Context, plaintext, additionalData []byte) ([]byte, error)
	Decrypt(ctx context.Context, ciphertext, additionalData []byte) ([]byte, error)
}

// NewAESEncryptor returns an AES-256-GCM encryptor with key versioning.
// currentVersion is the version byte written on every new Encrypt call.
// keys maps each version byte to its base64-encoded 32-byte key — must include
// currentVersion. Old versions are kept for Decrypt during a rotation window.
//
// Ciphertext format: [1-byte version][12-byte nonce][ciphertext+GCM tag].
//
// Use this for local development and CI via the SESSION_ENCRYPTION_KEY env var.
func NewAESEncryptor(currentVersion byte, keys map[byte]string) (Encryptor, error) {
	if len(keys) == 0 {
		return nil, errors.New("encryption: at least one key required")
	}
	if _, ok := keys[currentVersion]; !ok {
		return nil, fmt.Errorf("encryption: current version %d not present in keys map", currentVersion)
	}

	parsed := make(map[byte][32]byte, len(keys))
	for v, b64 := range keys {
		raw, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("encryption: decode key for version %d: %w", v, err)
		}
		if len(raw) != 32 {
			return nil, fmt.Errorf("encryption: key for version %d must be 32 bytes (AES-256)", v)
		}
		var arr [32]byte
		copy(arr[:], raw)
		parsed[v] = arr
	}
	return &aesEncryptor{current: currentVersion, keys: parsed}, nil
}

type aesEncryptor struct {
	current byte
	keys    map[byte][32]byte
}

// Encrypt encrypts plaintext using AES-256-GCM with the current key version.
// additionalData is bound into the GCM tag — pass the owning user_id bytes.
// Output: [1-byte version][12-byte nonce][ciphertext+GCM tag].
func (e *aesEncryptor) Encrypt(_ context.Context, plaintext, additionalData []byte) ([]byte, error) {
	key := e.keys[e.current]
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("encryption: generate nonce: %w", err)
	}
	// Prepend the version byte, then nonce, then the sealed ciphertext.
	out := make([]byte, 0, 1+len(nonce)+len(plaintext)+gcm.Overhead())
	out = append(out, e.current)
	out = append(out, nonce...)
	out = gcm.Seal(out, nonce, plaintext, additionalData)
	return out, nil
}

// Decrypt decrypts ciphertext produced by Encrypt.
// additionalData must match the value used during Encrypt.
func (e *aesEncryptor) Decrypt(_ context.Context, ciphertext, additionalData []byte) ([]byte, error) {
	if len(ciphertext) < 1 {
		return nil, errors.New("encryption: ciphertext too short")
	}
	version := ciphertext[0]
	key, ok := e.keys[version]
	if !ok {
		return nil, fmt.Errorf("encryption: unknown key version %d", version)
	}
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	rest := ciphertext[1:]
	nonceSize := gcm.NonceSize()
	if len(rest) < nonceSize {
		return nil, errors.New("encryption: ciphertext too short")
	}
	nonce, data := rest[:nonceSize], rest[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, data, additionalData)
	if err != nil {
		return nil, fmt.Errorf("encryption: decrypt: %w", err)
	}
	return plaintext, nil
}

func newGCM(key [32]byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("encryption: create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("encryption: create gcm: %w", err)
	}
	return gcm, nil
}
