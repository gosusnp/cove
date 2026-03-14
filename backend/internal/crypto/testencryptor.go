// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package crypto

import "encoding/base64"

// NewTestEncryptor returns an AES-256-GCM encryptor with a fixed key for use
// in tests. Never use this key outside of tests.
func NewTestEncryptor() Encryptor {
	// 32 zero bytes, base64-encoded — deterministic for testing only.
	key := base64.StdEncoding.EncodeToString(make([]byte, 32))
	enc, err := NewAESEncryptor(0, map[byte]string{0: key})
	if err != nil {
		panic("crypto.NewTestEncryptor: " + err.Error())
	}
	return enc
}
