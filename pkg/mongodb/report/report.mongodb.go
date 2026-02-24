// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package report

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/net/http"

	"github.com/LerianStudio/lib-commons/v2/commons"
	libMongo "github.com/LerianStudio/lib-commons/v2/commons/mongo"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/attribute"
)

// Repository provides an interface for operations related to reports collection in MongoDB.
//
//go:generate mockgen --destination=report.mongodb.mock.go --package=report --copyright_file=../../../COPYRIGHT . Repository
type Repository interface {
	UpdateReportStatusById(ctx context.Context, status string, id uuid.UUID, completedAt time.Time, metadata map[string]any) error
	Create(ctx context.Context, record *Report) (*Report, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Report, error)
	FindList(ctx context.Context, filters http.QueryHeader) ([]*Report, error)
}

// ReportMongoDBRepository is a MongoDB-specific implementation of the ReportRepository.
type ReportMongoDBRepository struct {
	connection *libMongo.MongoConnection
	Database   string
}

// Compile-time interface satisfaction check.
var _ Repository = (*ReportMongoDBRepository)(nil)

// NewReportMongoDBRepository returns a new instance of ReportMongoDBRepository using the given MongoDB connection.
func NewReportMongoDBRepository(mc *libMongo.MongoConnection) (*ReportMongoDBRepository, error) {
	r := &ReportMongoDBRepository{
		connection: mc,
		Database:   mc.Database,
	}
	if _, err := r.connection.GetDB(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb for reports: %w", err)
	}

	return r, nil
}

// UpdateReportStatusById updates only the status, completedAt and metadata fields of a report document by UUID.
func (rm *ReportMongoDBRepository) UpdateReportStatusById(
	ctx context.Context,
	status string,
	id uuid.UUID,
	completedAt time.Time,
	metadata map[string]any,
) error {
	_, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "repository.report.update_status")
	defer span.End()

	attributes := []attribute.KeyValue{
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.report_id", id.String()),
		attribute.String("app.request.status", status),
		attribute.String("app.request.completed_at", completedAt.String()),
	}

	span.SetAttributes(attributes...)

	db, err := rm.connection.GetDB(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get database", err)
		return err
	}

	coll := db.Database(strings.ToLower(rm.Database)).Collection(strings.ToLower(constant.MongoCollectionReport))

	// Create a filter using the UUID directly for matching the _id field stored as BinData
	filter := bson.M{"_id": id}

	ctx, spanUpdate := tracer.Start(ctx, "repository.report.update_status_exec")
	defer spanUpdate.End()

	spanUpdate.SetAttributes(attributes...)

	// Create an update document with only the fields we want to update
	updateFields := bson.M{}

	if status != "" {
		updateFields["status"] = status
	}

	// Only set completedAt if it's not a zero time
	if !completedAt.IsZero() {
		updateFields["completed_at"] = completedAt
	}

	// Only set metadata if it's not nil
	if metadata != nil {
		updateFields["metadata"] = metadata
	}

	// Use $set to update only the specified fields
	update := bson.M{
		"$set": updateFields,
	}

	err = libOpentelemetry.SetSpanAttributesFromStruct(&spanUpdate, "app.request.repository_input", update)
	if err != nil {
		libOpentelemetry.HandleSpanError(&spanUpdate, "Failed to convert update to JSON string", err)
	}

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		libOpentelemetry.HandleSpanError(&spanUpdate, "Failed to update report status", err)
		return err
	}

	if result.MatchedCount == 0 {
		libOpentelemetry.HandleSpanBusinessErrorEvent(&spanUpdate, "No report found with the provided UUID", constant.ErrEntityNotFound)
	}

	return nil
}

// Create inserts a new report entity into mongo.
func (rm *ReportMongoDBRepository) Create(ctx context.Context, report *Report) (*Report, error) {
	_, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "repository.report.create")
	defer span.End()

	attributes := []attribute.KeyValue{
		attribute.String("app.request.request_id", reqId),
	}

	span.SetAttributes(attributes...)

	err := libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.payload", report)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert report record to JSON string", err)
	}

	db, err := rm.connection.GetDB(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get database", err)

		return nil, err
	}

	coll := db.Database(strings.ToLower(rm.Database)).Collection(strings.ToLower(constant.MongoCollectionReport))
	record := &ReportMongoDBModel{}

	if err := record.FromEntity(report); err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert report to model", err)

		return nil, err
	}

	ctx, spanInsert := tracer.Start(ctx, "repository.report.create_exec")

	spanInsert.SetAttributes(attributes...)

	err = libOpentelemetry.SetSpanAttributesFromStruct(&spanInsert, "app.request.repository_input", record)
	if err != nil {
		libOpentelemetry.HandleSpanError(&spanInsert, "Failed to convert report record to JSON string", err)
	}

	_, err = coll.InsertOne(ctx, record)
	if err != nil {
		libOpentelemetry.HandleSpanError(&spanInsert, "Failed to insert report", err)

		return nil, err
	}

	spanInsert.End()

	return record.ToEntity(report.Filters), nil
}

// FindByID retrieves a report from the mongodb using the provided entity_id.
func (rm *ReportMongoDBRepository) FindByID(ctx context.Context, id uuid.UUID) (*Report, error) {
	_, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "repository.report.find_by_id")
	defer span.End()

	attributes := []attribute.KeyValue{
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.report_id", id.String()),
	}

	span.SetAttributes(attributes...)

	db, err := rm.connection.GetDB(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get database", err)

		return nil, err
	}

	coll := db.Database(strings.ToLower(rm.Database)).Collection(strings.ToLower(constant.MongoCollectionReport))

	var record *ReportMongoDBModel

	ctx, spanFindOne := tracer.Start(ctx, "repository.report.find_by_id_exec")

	spanFindOne.SetAttributes(attributes...)

	filter := bson.M{"_id": id, "deleted_at": bson.D{{Key: "$eq", Value: nil}}}

	if err = coll.
		FindOne(ctx, filter).
		Decode(&record); err != nil {
		libOpentelemetry.HandleSpanError(&spanFindOne, "Failed to find report by entity", err)
		return nil, err
	}

	if nil == record {
		libOpentelemetry.HandleSpanError(&span, "Failed to get database", err)
		return nil, mongo.ErrNoDocuments
	}

	spanFindOne.End()

	return record.ToEntityFindByID(), nil
}

// FindList retrieves all reports from the mongodb with filtering and pagination support.
func (rm *ReportMongoDBRepository) FindList(ctx context.Context, filters http.QueryHeader) ([]*Report, error) {
	_, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "repository.report.find_list")
	defer span.End()

	attributes := []attribute.KeyValue{
		attribute.String("app.request.request_id", reqId),
	}

	span.SetAttributes(attributes...)

	err := libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.payload", filters)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert filters to JSON string", err)
	}

	db, err := rm.connection.GetDB(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get database", err)
		return nil, err
	}

	coll := db.Database(strings.ToLower(rm.Database)).Collection(strings.ToLower(constant.MongoCollectionReport))

	queryFilter := bson.M{}

	// Filter by status
	if !commons.IsNilOrEmpty(&filters.Status) {
		queryFilter["status"] = filters.Status
	}

	// Filter by template_id
	if filters.TemplateID != uuid.Nil {
		queryFilter["template_id"] = filters.TemplateID
	}

	// Filter by created_at date range
	if !filters.CreatedAt.IsZero() {
		end := filters.CreatedAt.Add(constant.HoursPerDay * time.Hour)
		queryFilter["created_at"] = bson.M{
			"$gte": filters.CreatedAt,
			"$lt":  end,
		}
	}

	// Filter non-deleted records
	queryFilter["deleted_at"] = bson.D{{Key: "$eq", Value: nil}}

	// Pagination
	limit := int64(filters.Limit)
	skip := int64(filters.Page*filters.Limit - filters.Limit)
	opts := options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
		Sort:  bson.D{{Key: "created_at", Value: -1}}, // Sort by created_at desc
	}

	ctx, spanFind := tracer.Start(ctx, "repository.report.find_list_exec")

	spanFind.SetAttributes(attributes...)

	err = libOpentelemetry.SetSpanAttributesFromStruct(&spanFind, "app.request.repository_filter", filters)
	if err != nil {
		libOpentelemetry.HandleSpanError(&spanFind, "Failed to convert filters to JSON string", err)
	}

	cur, err := coll.Find(ctx, queryFilter, &opts)
	if err != nil {
		libOpentelemetry.HandleSpanError(&spanFind, "Failed to find reports", err)
		return nil, err
	}

	spanFind.End()

	var results []*ReportMongoDBModel

	for cur.Next(ctx) {
		var record ReportMongoDBModel
		if err := cur.Decode(&record); err != nil {
			libOpentelemetry.HandleSpanError(&span, "Failed to decode report", err)
			return nil, err
		}

		results = append(results, &record)
	}

	if err := cur.Err(); err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to iterate reports", err)
		return nil, err
	}

	if err := cur.Close(ctx); err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to close cursor", err)
		return nil, err
	}

	reports := make([]*Report, 0, len(results))
	for i := range results {
		reports = append(reports, results[i].ToEntityFindByID())
	}

	return reports, nil
}
