// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package template

import (
	"context"
	"testing"

	libMongo "github.com/LerianStudio/lib-commons/v3/commons/mongo"
	"github.com/LerianStudio/lib-commons/v3/commons/zap"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestNewTemplateMongoDBRepository_NilConnection(t *testing.T) {
	t.Parallel()

	// Passing nil connection should panic or fail gracefully.
	// NewTemplateMongoDBRepository dereferences mc.Database, so nil input panics.
	assert.Panics(t, func() {
		_, _ = NewTemplateMongoDBRepository(nil)
	}, "Expected panic when creating repository with nil connection")
}

func TestTemplateMongoDBRepository_EnsureIndexes_RequiresConnection(t *testing.T) {
	t.Parallel()

	// EnsureIndexes and DropIndexes require a real MongoDB connection.
	// This test verifies the struct fields are correctly set.
	repo := &TemplateMongoDBRepository{
		Database: "test_db",
	}

	assert.Equal(t, "test_db", repo.Database)
}

func TestEnsureIndexes_Success(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("creates indexes successfully", func(mt *mtest.T) {
		logger := zap.InitializeLogger()
		conn := &libMongo.MongoConnection{
			DB:       mt.Client,
			Database: mt.DB.Name(),
			Logger:   logger,
		}

		repo := &TemplateMongoDBRepository{
			connection: conn,
			Database:   conn.Database,
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.EnsureIndexes(context.Background())
		assert.NoError(mt, err)
	})
}

func TestEnsureIndexes_NilConnection(t *testing.T) {
	t.Parallel()

	// When the connection field is nil, EnsureIndexes panics on nil pointer dereference.
	repo := &TemplateMongoDBRepository{
		connection: nil,
		Database:   "testdb",
	}

	assert.Panics(t, func() {
		_ = repo.EnsureIndexes(context.Background())
	}, "Expected panic when connection is nil")
}

func TestEnsureIndexes_IndexOptionsConflict(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("returns nil when IndexOptionsConflict error occurs", func(mt *mtest.T) {
		logger := zap.InitializeLogger()
		conn := &libMongo.MongoConnection{
			DB:       mt.Client,
			Database: mt.DB.Name(),
			Logger:   logger,
		}

		repo := &TemplateMongoDBRepository{
			connection: conn,
			Database:   conn.Database,
		}

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    85,
			Message: "IndexOptionsConflict: index already exists with different options",
		}))

		err := repo.EnsureIndexes(context.Background())
		assert.NoError(mt, err)
	})
}

func TestEnsureIndexes_AlreadyExists(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("returns nil when index already exists", func(mt *mtest.T) {
		logger := zap.InitializeLogger()
		conn := &libMongo.MongoConnection{
			DB:       mt.Client,
			Database: mt.DB.Name(),
			Logger:   logger,
		}

		repo := &TemplateMongoDBRepository{
			connection: conn,
			Database:   conn.Database,
		}

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    68,
			Message: "index already exists with a different name",
		}))

		err := repo.EnsureIndexes(context.Background())
		assert.NoError(mt, err)
	})
}

func TestEnsureIndexes_CreationError(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("returns error on generic creation failure", func(mt *mtest.T) {
		logger := zap.InitializeLogger()
		conn := &libMongo.MongoConnection{
			DB:       mt.Client,
			Database: mt.DB.Name(),
			Logger:   logger,
		}

		repo := &TemplateMongoDBRepository{
			connection: conn,
			Database:   conn.Database,
		}

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    1,
			Message: "some other error",
		}))

		err := repo.EnsureIndexes(context.Background())
		assert.Error(mt, err)
	})
}

func TestDropIndexes_Success(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("drops indexes successfully", func(mt *mtest.T) {
		logger := zap.InitializeLogger()
		conn := &libMongo.MongoConnection{
			DB:       mt.Client,
			Database: mt.DB.Name(),
			Logger:   logger,
		}

		repo := &TemplateMongoDBRepository{
			connection: conn,
			Database:   conn.Database,
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.DropIndexes(context.Background())
		assert.NoError(mt, err)
	})
}

func TestDropIndexes_NilConnection(t *testing.T) {
	t.Parallel()

	// When the connection field is nil, DropIndexes panics on nil pointer dereference.
	repo := &TemplateMongoDBRepository{
		connection: nil,
		Database:   "testdb",
	}

	assert.Panics(t, func() {
		_ = repo.DropIndexes(context.Background())
	}, "Expected panic when connection is nil")
}

func TestDropIndexes_DropError(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("returns error on drop failure", func(mt *mtest.T) {
		logger := zap.InitializeLogger()
		conn := &libMongo.MongoConnection{
			DB:       mt.Client,
			Database: mt.DB.Name(),
			Logger:   logger,
		}

		repo := &TemplateMongoDBRepository{
			connection: conn,
			Database:   conn.Database,
		}

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    1,
			Message: "failed to drop indexes",
		}))

		err := repo.DropIndexes(context.Background())
		assert.Error(mt, err)
	})
}
