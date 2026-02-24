// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUseCase_HasCryptoHashSecretKeyField verifies that the worker UseCase struct
// has a CryptoHashSecretKeyPluginCRM field for centralized configuration
// instead of using os.Getenv("CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM").
func TestUseCase_HasCryptoHashSecretKeyField(t *testing.T) {
	t.Parallel()

	uc := &UseCase{
		CryptoHashSecretKeyPluginCRM: "test-hash-secret-key",
	}

	assert.Equal(t, "test-hash-secret-key", uc.CryptoHashSecretKeyPluginCRM)
}

// TestUseCase_HasCryptoEncryptSecretKeyField verifies that the worker UseCase struct
// has a CryptoEncryptSecretKeyPluginCRM field for centralized configuration
// instead of using os.Getenv("CRYPTO_ENCRYPT_SECRET_KEY_PLUGIN_CRM").
func TestUseCase_HasCryptoEncryptSecretKeyField(t *testing.T) {
	t.Parallel()

	uc := &UseCase{
		CryptoEncryptSecretKeyPluginCRM: "test-encrypt-secret-key",
	}

	assert.Equal(t, "test-encrypt-secret-key", uc.CryptoEncryptSecretKeyPluginCRM)
}
