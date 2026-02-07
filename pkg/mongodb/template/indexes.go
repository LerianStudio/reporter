// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package template

import (
	"context"
	"strings"

	"github.com/LerianStudio/reporter/pkg/constant"

	"github.com/LerianStudio/lib-commons/v2/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/attribute"
)

// EnsureIndexes creates all indexes for the templates collection.
func (tm *TemplateMongoDBRepository) EnsureIndexes(ctx context.Context) error {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "repository.template.ensure_indexes")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.collection", constant.MongoCollectionTemplate),
	)

	logger.Infof("Creating indexes for %s collection", constant.MongoCollectionTemplate)

	db, err := tm.connection.GetDB(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get database", err)
		return err
	}

	coll := db.Database(strings.ToLower(tm.Database)).Collection(strings.ToLower(constant.MongoCollectionTemplate))

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "_id", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().
				SetName("idx_template_id_deleted"),
		},

		{
			Keys: bson.D{
				{Key: "deleted_at", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().
				SetName("idx_template_list_main").
				SetPartialFilterExpression(bson.D{
					{Key: "deleted_at", Value: nil},
				}),
		},

		{
			Keys: bson.D{
				{Key: "deleted_at", Value: 1},
				{Key: "output_format", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().
				SetName("idx_template_format").
				SetPartialFilterExpression(bson.D{
					{Key: "deleted_at", Value: nil},
				}),
		},

		{
			Keys: bson.D{
				{Key: "description", Value: "text"},
			},
			Options: options.Index().
				SetName("idx_template_description_text").
				SetWeights(bson.D{
					{Key: "description", Value: constant.MongoTextSearchWeight},
				}),
		},
	}

	ctx, cancel := context.WithTimeout(ctx, constant.MongoIndexCreateTimeout)
	defer cancel()

	logger.Infof("Attempting to create %d indexes for %s collection (removed SetBackground - deprecated since MongoDB 4.2)", len(indexes), constant.MongoCollectionTemplate)

	indexNames, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		// Check if error is due to indexes already existing
		if strings.Contains(err.Error(), "IndexOptionsConflict") ||
			strings.Contains(err.Error(), "already exists") {
			logger.Infof("Indexes for %s already exist (detected during creation)", constant.MongoCollectionTemplate)
			return nil
		}

		libOpentelemetry.HandleSpanError(&span, "Failed to create indexes", err)
		logger.Errorf("Failed to create indexes for %s: %v", constant.MongoCollectionTemplate, err)

		return err
	}

	logger.Infof("Successfully created %d indexes for %s collection: %v",
		len(indexNames), constant.MongoCollectionTemplate, indexNames)

	return nil
}

// DropIndexes removes all custom indexes for the templates collection.
func (tm *TemplateMongoDBRepository) DropIndexes(ctx context.Context) error {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "repository.template.drop_indexes")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.collection", constant.MongoCollectionTemplate),
	)

	logger.Warnf("Dropping all custom indexes for %s collection", constant.MongoCollectionTemplate)

	db, err := tm.connection.GetDB(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get database", err)
		return err
	}

	coll := db.Database(strings.ToLower(tm.Database)).Collection(strings.ToLower(constant.MongoCollectionTemplate))

	ctx, cancel := context.WithTimeout(ctx, constant.MongoIndexDropTimeout)
	defer cancel()

	if _, err := coll.Indexes().DropAll(ctx); err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to drop indexes", err)
		logger.Errorf("Failed to drop indexes for %s: %v", constant.MongoCollectionTemplate, err)

		return err
	}

	logger.Infof("Successfully dropped all custom indexes for %s collection", constant.MongoCollectionTemplate)

	return nil
}
