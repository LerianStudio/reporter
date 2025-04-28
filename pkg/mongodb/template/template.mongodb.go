package template

import (
	"context"
	"errors"
	"github.com/LerianStudio/lib-commons/commons"
	libMongo "github.com/LerianStudio/lib-commons/commons/mongo"
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"plugin-template-engine/pkg"
	"plugin-template-engine/pkg/constant"
	"strings"
)

// Repository provides an interface for operations related on mongo a metadata entities.
//
//go:generate mockgen --destination=template.mongodb.mock.go --package=template . Repository
type Repository interface {
	FindByID(ctx context.Context, collection string, id, organizationID uuid.UUID) (*Template, error)
	Create(ctx context.Context, collection string, record *TemplateMongoDBModel) (*Template, error)
	Update(ctx context.Context, collection string, id, organizationID uuid.UUID, updateFields *bson.M) error
	FindOutputFormatByID(ctx context.Context, collection string, id, organizationID uuid.UUID) (*string, error)
}

// TemplateMongoDBRepository is a MongoDD-specific implementation of the PackageRepository.
type TemplateMongoDBRepository struct {
	connection *libMongo.MongoConnection
	Database   string
}

// NewTemplateMongoDBRepository returns a new instance of TemplateMongoDBRepository using the given MongoDB connection.
func NewTemplateMongoDBRepository(mc *libMongo.MongoConnection) *TemplateMongoDBRepository {
	r := &TemplateMongoDBRepository{
		connection: mc,
		Database:   mc.Database,
	}
	if _, err := r.connection.GetDB(context.Background()); err != nil {
		panic("Failed to connect mongo")
	}

	return r
}

// FindByID retrieves a template from the mongodb using the provided entity_id.
func (tm *TemplateMongoDBRepository) FindByID(ctx context.Context, collection string, id, organizationID uuid.UUID) (*Template, error) {
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongodb.find_by_entity")
	defer span.End()

	db, err := tm.connection.GetDB(ctx)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get database", err)

		return nil, err
	}

	coll := db.Database(strings.ToLower(tm.Database)).Collection(strings.ToLower(collection))

	var record *TemplateMongoDBModel

	ctx, spanFindOne := tracer.Start(ctx, "mongodb.find_by_entity.find_one")

	if err = coll.
		FindOne(ctx, bson.M{"_id": id, "organization_id": organizationID, "deleted_at": bson.D{{Key: "$eq", Value: nil}}}).
		Decode(&record); err != nil {
		opentelemetry.HandleSpanError(&spanFindOne, "Failed to find template by entity", err)
		return nil, err
	}

	if nil == record {
		opentelemetry.HandleSpanError(&span, "Failed to get database", err)
		return nil, mongo.ErrNoDocuments
	}

	spanFindOne.End()

	return record.ToEntity(), nil
}

// FindOutputFormatByID retrieves outputFormat of a template provided entity_id.
func (tm *TemplateMongoDBRepository) FindOutputFormatByID(ctx context.Context, collection string, id, organizationID uuid.UUID) (*string, error) {
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongodb.find_by_entity")
	defer span.End()

	db, err := tm.connection.GetDB(ctx)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get database", err)

		return nil, err
	}

	coll := db.Database(strings.ToLower(tm.Database)).Collection(strings.ToLower(collection))

	var record struct {
		OutputFormat string `bson:"output_format"`
	}

	opts := options.FindOne().SetProjection(bson.M{
		"output_format": 1,
		"_id":           0,
	})

	if err = coll.
		FindOne(ctx, bson.M{
			"_id":             id,
			"organization_id": organizationID,
			"deleted_at":      bson.D{{Key: "$eq", Value: nil}},
		}, opts).
		Decode(&record); err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to find template output_format by entity", err)
		return nil, err
	}

	span.End()

	return &record.OutputFormat, nil
}

// Create inserts a new package entity into mongo.
func (tm *TemplateMongoDBRepository) Create(ctx context.Context, collection string, record *TemplateMongoDBModel) (*Template, error) {
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongo.create_template")
	defer span.End()

	db, err := tm.connection.GetDB(ctx)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get database", err)

		return nil, err
	}

	coll := db.Database(strings.ToLower(tm.Database)).Collection(strings.ToLower(collection))

	ctx, spanInsert := tracer.Start(ctx, "mongo.create_template.insert")

	_, err = coll.InsertOne(ctx, record)
	if err != nil {
		opentelemetry.HandleSpanError(&spanInsert, "Failed to insert template", err)

		return nil, err
	}

	spanInsert.End()

	return record.ToEntity(), nil
}

// Update a template entity into mongodb.
func (tm *TemplateMongoDBRepository) Update(ctx context.Context, collection string, id, organizationID uuid.UUID, updateFields *bson.M) error {
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongodb.update_template")
	defer span.End()

	db, err := tm.connection.GetDB(ctx)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get database", err)
		return err
	}

	coll := db.Database(strings.ToLower(tm.Database)).Collection(strings.ToLower(collection))
	opts := options.Update().SetUpsert(false)

	ctx, spanUpdate := tracer.Start(ctx, "mongodb.update_template.update_one")

	err = opentelemetry.SetSpanAttributesFromStruct(&spanUpdate, "update_template_input", updateFields)
	if err != nil {
		opentelemetry.HandleSpanError(&spanUpdate, "Failed to convert template record from entity to JSON string", err)

		return err
	}

	_, err = coll.UpdateOne(ctx, bson.M{"_id": id, "organization_id": organizationID}, updateFields, opts)
	if err != nil {
		opentelemetry.HandleSpanError(&spanUpdate, "Failed to update template", err)

		if errors.Is(err, mongo.ErrNoDocuments) {
			return pkg.ValidateBusinessError(constant.ErrEntityNotFound, "", collection)
		}

		return err
	}

	spanUpdate.End()

	return nil
}
