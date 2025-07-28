package template

import (
	"context"
	"errors"
	"plugin-smart-templates/pkg"
	"plugin-smart-templates/pkg/constant"
	"plugin-smart-templates/pkg/net/http"
	"strings"
	"time"

	"github.com/LerianStudio/lib-commons/commons"
	libMongo "github.com/LerianStudio/lib-commons/commons/mongo"
	libOpentelemetry "github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/attribute"
)

// Repository provides an interface for operations related on mongo a metadata entities.
//
//go:generate mockgen --destination=template.mongodb.mock.go --package=template . Repository
type Repository interface {
	FindByID(ctx context.Context, id, organizationID uuid.UUID) (*Template, error)
	FindList(ctx context.Context, filters http.QueryHeader) ([]*Template, error)
	Create(ctx context.Context, record *TemplateMongoDBModel) (*Template, error)
	Update(ctx context.Context, id, organizationID uuid.UUID, updateFields *bson.M) error
	SoftDelete(ctx context.Context, id, organizationID uuid.UUID) error
	FindOutputFormatByID(ctx context.Context, id, organizationID uuid.UUID) (*string, error)
	FindMappedFieldsAndOutputFormatByID(ctx context.Context, id, organizationID uuid.UUID) (*string, map[string]map[string][]string, error)
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
func (tm *TemplateMongoDBRepository) FindByID(ctx context.Context, id, organizationID uuid.UUID) (*Template, error) {
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongodb.find_by_entity")
	defer span.End()

	db, err := tm.connection.GetDB(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get database", err)

		return nil, err
	}

	coll := db.Database(strings.ToLower(tm.Database)).Collection(strings.ToLower(constant.MongoCollectionTemplate))

	var record *TemplateMongoDBModel

	ctx, spanFindOne := tracer.Start(ctx, "mongodb.find_by_entity.find_one")

	spanFindOne.SetAttributes(
		attribute.String("template_id", id.String()),
		attribute.String("organization_id", organizationID.String()),
	)

	if err = coll.
		FindOne(ctx, bson.M{"_id": id, "organization_id": organizationID, "deleted_at": bson.D{{Key: "$eq", Value: nil}}}).
		Decode(&record); err != nil {
		libOpentelemetry.HandleSpanError(&spanFindOne, "Failed to find template by entity", err)
		return nil, err
	}

	if nil == record {
		libOpentelemetry.HandleSpanError(&span, "Failed to get database", err)
		return nil, mongo.ErrNoDocuments
	}

	spanFindOne.End()

	return record.ToEntity(), nil
}

// FindList retrieves all templates from the mongodb using the provided organization_id.
func (tm *TemplateMongoDBRepository) FindList(ctx context.Context, filters http.QueryHeader) ([]*Template, error) {
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongodb.find_all_templates")
	defer span.End()

	db, err := tm.connection.GetDB(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get database", err)
		return nil, err
	}

	coll := db.Database(strings.ToLower(tm.Database)).Collection(strings.ToLower(constant.MongoCollectionTemplate))

	queryFilter := bson.M{}

	if !commons.IsNilOrEmpty(&filters.OutputFormat) {
		queryFilter["output_format"] = filters.OutputFormat
	}

	if !filters.CreatedAt.IsZero() {
		end := filters.CreatedAt.Add(24 * time.Hour)
		queryFilter["created_at"] = bson.M{
			"$gte": filters.CreatedAt,
			"$lt":  end,
		}
	}

	if !commons.IsNilOrEmpty(&filters.Description) {
		queryFilter["description"] = bson.M{
			"$regex":   filters.Description,
			"$options": "i", // "i" = case-insensitive
		}
	}

	queryFilter["organization_id"] = filters.OrganizationID
	queryFilter["deleted_at"] = bson.D{{Key: "$eq", Value: nil}}

	limit := int64(filters.Limit)
	skip := int64(filters.Page*filters.Limit - filters.Limit)
	opts := options.FindOptions{Limit: &limit, Skip: &skip}

	ctx, spanFind := tracer.Start(ctx, "mongodb.find_templates.find")

	err = libOpentelemetry.SetSpanAttributesFromStruct(&spanFind, "filters", filters)
	if err != nil {
		libOpentelemetry.HandleSpanError(&spanFind, "Failed to convert filters to JSON string", err)
	}

	cur, err := coll.Find(ctx, queryFilter, &opts)
	if err != nil {
		libOpentelemetry.HandleSpanError(&spanFind, "Failed to find templates", err)
		return nil, err
	}

	spanFind.End()

	var results []*TemplateMongoDBModel

	for cur.Next(ctx) {
		var record TemplateMongoDBModel
		if err := cur.Decode(&record); err != nil {
			libOpentelemetry.HandleSpanError(&span, "Failed to decode template", err)
			return nil, err
		}

		results = append(results, &record)
	}

	if err := cur.Err(); err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to iterate templates", err)
		return nil, err
	}

	if err := cur.Close(ctx); err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to close cursor", err)
		return nil, err
	}

	templates := make([]*Template, 0, len(results))
	for i := range results {
		templates = append(templates, results[i].ToEntity())
	}

	return templates, nil
}

// FindOutputFormatByID retrieves outputFormat of a template provided entity_id.
func (tm *TemplateMongoDBRepository) FindOutputFormatByID(ctx context.Context, id, organizationID uuid.UUID) (*string, error) {
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongodb.find_by_entity")
	defer span.End()

	span.SetAttributes(
		attribute.String("template_id", id.String()),
		attribute.String("organization_id", organizationID.String()),
	)

	db, err := tm.connection.GetDB(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get database", err)

		return nil, err
	}

	coll := db.Database(strings.ToLower(tm.Database)).Collection(strings.ToLower(constant.MongoCollectionTemplate))

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
		libOpentelemetry.HandleSpanError(&span, "Failed to find template output_format by entity", err)
		return nil, err
	}

	span.End()

	return &record.OutputFormat, nil
}

// Create inserts a new package entity into mongo.
func (tm *TemplateMongoDBRepository) Create(ctx context.Context, record *TemplateMongoDBModel) (*Template, error) {
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongo.create_template")
	defer span.End()

	db, err := tm.connection.GetDB(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get database", err)

		return nil, err
	}

	coll := db.Database(strings.ToLower(tm.Database)).Collection(strings.ToLower(constant.MongoCollectionTemplate))

	ctx, spanInsert := tracer.Start(ctx, "mongo.create_template.insert")

	err = libOpentelemetry.SetSpanAttributesFromStruct(&spanInsert, "template_record", record)
	if err != nil {
		libOpentelemetry.HandleSpanError(&spanInsert, "Failed to convert template record to JSON string", err)
	}

	_, err = coll.InsertOne(ctx, record)
	if err != nil {
		libOpentelemetry.HandleSpanError(&spanInsert, "Failed to insert template", err)

		return nil, err
	}

	spanInsert.End()

	return record.ToEntity(), nil
}

// Update a template entity into mongodb.
func (tm *TemplateMongoDBRepository) Update(ctx context.Context, id, organizationID uuid.UUID, updateFields *bson.M) error {
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongodb.update_template")
	defer span.End()

	db, err := tm.connection.GetDB(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get database", err)
		return err
	}

	coll := db.Database(strings.ToLower(tm.Database)).Collection(strings.ToLower(constant.MongoCollectionTemplate))
	opts := options.Update().SetUpsert(false)

	ctx, spanUpdate := tracer.Start(ctx, "mongodb.update_template.update_one")

	spanUpdate.SetAttributes(
		attribute.String("template_id", id.String()),
		attribute.String("organization_id", organizationID.String()),
	)

	err = libOpentelemetry.SetSpanAttributesFromStruct(&spanUpdate, "update_template_input", updateFields)
	if err != nil {
		libOpentelemetry.HandleSpanError(&spanUpdate, "Failed to convert template record from entity to JSON string", err)

		return err
	}

	_, err = coll.UpdateOne(ctx, bson.M{"_id": id, "organization_id": organizationID}, updateFields, opts)
	if err != nil {
		libOpentelemetry.HandleSpanError(&spanUpdate, "Failed to update template", err)

		if errors.Is(err, mongo.ErrNoDocuments) {
			return pkg.ValidateBusinessError(constant.ErrEntityNotFound, "", constant.MongoCollectionTemplate)
		}

		return err
	}

	spanUpdate.End()

	return nil
}

// SoftDelete a template entity into mongodb.
func (tm *TemplateMongoDBRepository) SoftDelete(ctx context.Context, id, organizationID uuid.UUID) error {
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongodb.delete_template")
	defer span.End()

	db, err := tm.connection.GetDB(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get database", err)

		return err
	}

	logger := commons.NewLoggerFromContext(ctx)

	coll := db.Database(strings.ToLower(tm.Database)).Collection(strings.ToLower(constant.MongoCollectionTemplate))

	ctx, spanDelete := tracer.Start(ctx, "mongodb.delete_template.delete_one")

	spanDelete.SetAttributes(
		attribute.String("template_id", id.String()),
		attribute.String("organization_id", organizationID.String()),
	)

	filter := bson.D{{Key: "_id", Value: id}, {Key: "organization_id", Value: organizationID}}
	deletedAt := bson.D{{Key: "$set", Value: bson.D{{Key: "deleted_at", Value: time.Now()}}}}

	deleted, err := coll.UpdateOne(ctx, filter, deletedAt)
	if err != nil {
		libOpentelemetry.HandleSpanError(&spanDelete, "Failed to delete template", err)

		return err
	}

	logger.Infof("Return from delete one: %v", deleted)
	spanDelete.End()

	return nil
}

// FindMappedFieldsAndOutputFormatByID find mapped fields of template and output format.
func (tm *TemplateMongoDBRepository) FindMappedFieldsAndOutputFormatByID(ctx context.Context, id, organizationID uuid.UUID) (*string, map[string]map[string][]string, error) {
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongodb.find_mapped_fields_and_output_format_by_id")
	defer span.End()

	span.SetAttributes(
		attribute.String("template_id", id.String()),
		attribute.String("organization_id", organizationID.String()),
	)

	db, err := tm.connection.GetDB(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get database", err)

		return nil, nil, err
	}

	coll := db.Database(strings.ToLower(tm.Database)).Collection(strings.ToLower(constant.MongoCollectionTemplate))

	var record struct {
		OutputFormat string                         `bson:"output_format"`
		MappedFields map[string]map[string][]string `bson:"mapped_fields"`
	}

	opts := options.FindOne().SetProjection(bson.M{
		"output_format": 1,
		"mapped_fields": 1,
		"_id":           0,
	})

	if err = coll.
		FindOne(ctx, bson.M{
			"_id":             id,
			"organization_id": organizationID,
			"deleted_at":      bson.D{{Key: "$eq", Value: nil}},
		}, opts).
		Decode(&record); err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to find template output_format and mapped_fields by entity ID", err)
		return nil, nil, err
	}

	span.End()

	return &record.OutputFormat, record.MappedFields, nil
}
