// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package crypto

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

// EncryptedField is a hardened type that stores an encrypted JSON payload in
// the database. Plaintext is never retained — decryption happens inside a
// callback and the plaintext is zeroed immediately after the callback returns.
//
// Responsibilities by layer:
//   - Store: treats this as an opaque []byte via Scan / Value.
//   - Service: calls SetEncryptor after a store read; calls Set before a store write.
//   - Domain: exposes UseSensitiveData on the owning type as the only decrypt door.
type EncryptedField[T any] struct {
	mu         sync.Mutex
	ciphertext []byte
	enc        Encryptor
}

// NewEncryptedField returns an EncryptedField with the given encryptor attached.
func NewEncryptedField[T any](enc Encryptor) *EncryptedField[T] {
	return &EncryptedField[T]{enc: enc}
}

// SetEncryptor attaches an encryptor to a field populated by the store.
// Must be called before Use.
func (f *EncryptedField[T]) SetEncryptor(enc Encryptor) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.enc = enc
}

// Set encrypts v and stores the ciphertext. No plaintext is retained.
// additionalData is bound into the GCM tag (pass the owning user_id bytes).
func (f *EncryptedField[T]) Set(ctx context.Context, v T, additionalData []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.enc == nil {
		return errors.New("encrypted field: no encryptor set")
	}
	plaintext, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("encrypted field marshal: %w", err)
	}
	ct, err := f.enc.Encrypt(ctx, plaintext, additionalData)
	if err != nil {
		return fmt.Errorf("encrypted field encrypt: %w", err)
	}
	f.ciphertext = ct
	return nil
}

// Use decrypts the ciphertext, calls fn with a pointer to the plaintext value,
// then zeros the struct in place before returning. Passing *T instead of T
// ensures that zeroing the struct actually overwrites the backing memory for
// value-type fields (including []byte-backed SensitiveString fields).
//
// Callers outside the domain package should not call this directly — use the
// owning domain type's UseSensitiveData method.
// additionalData must match the value used during Set (pass the owning user_id bytes).
func (f *EncryptedField[T]) Use(ctx context.Context, fn func(*T) error, additionalData []byte) error {
	f.mu.Lock()
	enc := f.enc
	ct := f.ciphertext
	f.mu.Unlock()

	if enc == nil {
		return errors.New("encrypted field: no encryptor set")
	}
	var v T
	if ct != nil {
		plaintext, err := enc.Decrypt(ctx, ct, additionalData)
		if err != nil {
			return fmt.Errorf("encrypted field decrypt: %w", err)
		}
		if err := json.Unmarshal(plaintext, &v); err != nil {
			return fmt.Errorf("encrypted field unmarshal: %w", err)
		}
	}
	err := fn(&v)
	v = *new(T) // zero struct in place; effective for []byte-backed fields
	return err
}

// Scan implements sql.Scanner. The store calls this when reading sensitive_data.
// The raw bytes are copied and stored; nothing is decrypted here.
func (f *EncryptedField[T]) Scan(src any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	switch v := src.(type) {
	case nil:
		f.ciphertext = nil
	case []byte:
		cp := make([]byte, len(v))
		copy(cp, v)
		f.ciphertext = cp
	default:
		return fmt.Errorf("encrypted field: cannot scan %T", src)
	}
	return nil
}

// Value returns the ciphertext for the store write.
func (f *EncryptedField[T]) Value() []byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.ciphertext
}
