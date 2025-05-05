package report

import (
	"context"
	"github.com/LerianStudio/lib-commons/commons"
	libMongo "github.com/LerianStudio/lib-commons/commons/mongo"
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
	"time"
)

// Repository provides an interface for operations related to reports collection in MongoDB.
//
//go:generate mockgen --destination=report.mongodb.mock.go --package=report . Repository
type Repository interface {
	UpdateReportStatusById(ctx context.Context, collection string, id uuid.UUID, status string, completedAt time.Time, metadata map[string]any) error
	Create(ctx context.Context, collection string, record *Report, organizationID uuid.UUID) (*Report, error)
	FindByID(ctx context.Context, collection string, id, organizationID uuid.UUID) (*Report, error)
}

// ReportMongoDBRepository is a MongoDB-specific implementation of the ReportRepository.
type ReportMongoDBRepository struct {
	connection *libMongo.MongoConnection
	Database   string
}

// NewReportMongoDBRepository returns a new instance of ReportMongoDBRepository using the given MongoDB connection.
func NewReportMongoDBRepository(mc *libMongo.MongoConnection) *ReportMongoDBRepository {
	r := &ReportMongoDBRepository{
		connection: mc,
		Database:   mc.Database,
	}
	if _, err := r.connection.GetDB(context.Background()); err != nil {
		panic("Failed to connect mongo")
	}

	return r
}

// UpdateReportStatusById updates only the status, completedAt and metadata fields of a report document by UUID.
func (rm *ReportMongoDBRepository) UpdateReportStatusById(
	ctx context.Context,
	collection string,
	id uuid.UUID,
	status string,
	completedAt time.Time,
	metadata map[string]any,
) error {
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongo.update_report_status")
	defer span.End()

	db, err := rm.connection.GetDB(ctx)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get database", err)
		return err
	}

	coll := db.Database(strings.ToLower(rm.Database)).Collection(strings.ToLower(collection))

	// Create a filter using the UUID directly for matching the _id field stored as BinData
	filter := bson.M{"_id": id}

	ctx, spanUpdate := tracer.Start(ctx, "mongo.update_report_status.update")
	defer spanUpdate.End()

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

	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		opentelemetry.HandleSpanError(&spanUpdate, "Failed to update report status", err)
		return err
	}

	if result.MatchedCount == 0 {
		opentelemetry.HandleSpanError(&spanUpdate, "No report found with the provided UUID", nil)
	}

	return nil
}

// Create inserts a new report entity into mongo.
func (rm *ReportMongoDBRepository) Create(ctx context.Context, collection string, report *Report, organizationID uuid.UUID) (*Report, error) {
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongo.create_report")
	defer span.End()

	db, err := rm.connection.GetDB(ctx)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get database", err)

		return nil, err
	}

	coll := db.Database(strings.ToLower(rm.Database)).Collection(strings.ToLower(collection))
	record := &ReportMongoDBModel{}

	if err := record.FromEntity(report, organizationID); err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to convert report to model", err)

		return nil, err
	}

	ctx, spanInsert := tracer.Start(ctx, "mongo.create_report.insert")

	_, err = coll.InsertOne(ctx, record)
	if err != nil {
		opentelemetry.HandleSpanError(&spanInsert, "Failed to insert report", err)

		return nil, err
	}

	spanInsert.End()

	return record.ToEntity(report.Filters), nil
}

// FindByID retrieves a report from the mongodb using the provided entity_id.
func (rm *ReportMongoDBRepository) FindByID(ctx context.Context, collection string, id, organizationID uuid.UUID) (*Report, error) {
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongodb.find_by_entity")
	defer span.End()

	db, err := rm.connection.GetDB(ctx)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get database", err)

		return nil, err
	}

	coll := db.Database(strings.ToLower(rm.Database)).Collection(strings.ToLower(collection))

	var record *ReportMongoDBModel

	ctx, spanFindOne := tracer.Start(ctx, "mongodb.find_by_entity.find_one")

	if err = coll.
		FindOne(ctx, bson.M{"_id": id, "organization_id": organizationID, "deleted_at": bson.D{{Key: "$eq", Value: nil}}}).
		Decode(&record); err != nil {
		opentelemetry.HandleSpanError(&spanFindOne, "Failed to find report by entity", err)
		return nil, err
	}

	if nil == record {
		opentelemetry.HandleSpanError(&span, "Failed to get database", err)
		return nil, mongo.ErrNoDocuments
	}

	spanFindOne.End()

	return record.ToEntityFindByID(), nil
}
