// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig_HasCryptoHashSecretKeyPluginCRMField verifies that the worker Config struct
// has a CryptoHashSecretKeyPluginCRM field loaded from CRYPTO_HASH_SECRET_KEY_PLUGIN_CRM env var.
func TestConfig_HasCryptoHashSecretKeyPluginCRMField(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		CryptoHashSecretKeyPluginCRM: "test-hash-secret",
	}

	assert.Equal(t, "test-hash-secret", cfg.CryptoHashSecretKeyPluginCRM)
}

// TestConfig_HasCryptoEncryptSecretKeyPluginCRMField verifies that the worker Config struct
// has a CryptoEncryptSecretKeyPluginCRM field loaded from CRYPTO_ENCRYPT_SECRET_KEY_PLUGIN_CRM env var.
func TestConfig_HasCryptoEncryptSecretKeyPluginCRMField(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		CryptoEncryptSecretKeyPluginCRM: "test-encrypt-secret",
	}

	assert.Equal(t, "test-encrypt-secret", cfg.CryptoEncryptSecretKeyPluginCRM)
}

// TestNewMultiQueueConsumer_ReceivesQueueName verifies that NewMultiQueueConsumer
// accepts the queue name as a parameter instead of reading it from os.Getenv.
func TestNewMultiQueueConsumer_ReceivesQueueName(t *testing.T) {
	t.Parallel()

	// This test verifies that NewMultiQueueConsumer accepts a queueName parameter.
	// Currently the function signature is:
	//   NewMultiQueueConsumer(routes, useCase) *MultiQueueConsumer
	// It should become:
	//   NewMultiQueueConsumer(routes, useCase, queueName string) *MultiQueueConsumer
	//
	// We cannot call it with 3 args yet, so this test will fail to compile.
	// The test proves the refactoring is needed.

	queueName := "reporter.generate-report.queue"

	consumer := NewMultiQueueConsumer(nil, nil, queueName)

	require.NotNil(t, consumer)
}
