package template

import (
	"context"
	"plugin-smart-templates/v3/pkg"
	"plugin-smart-templates/v3/pkg/constant"
	"plugin-smart-templates/v3/pkg/net/http"
	"strings"
	"time"

	"github.com/LerianStudio/lib-commons/v2/commons"
	libMongo "github.com/LerianStudio/lib-commons/v2/commons/mongo"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
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
	Delete(ctx context.Context, id, organizationID uuid.UUID, hardDelete bool) error
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
	_, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongodb.find_by_entity")
	defer span.End()

	attributes := []attribute.KeyValue{
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_id", id.String()),
		attribute.String("app.request.organization_id", organizationID.String()),
	}

	span.SetAttributes(attributes...)

	db, err := tm.connection.GetDB(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get database", err)

		return nil, err
	}

	coll := db.Database(strings.ToLower(tm.Database)).Collection(strings.ToLower(constant.MongoCollectionTemplate))

	var record *TemplateMongoDBModel

	ctx, spanFindOne := tracer.Start(ctx, "mongodb.find_by_entity.find_one")

	spanFindOne.SetAttributes(attributes...)

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

// FindList retrieves all templates from the mongodb using the provided filters.
func (tm *TemplateMongoDBRepository) FindList(ctx context.Context, filters http.QueryHeader) ([]*Template, error) {
	_, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongodb.find_all_templates")
	defer span.End()

	attributes := []attribute.KeyValue{
		attribute.String("app.request.request_id", reqId),
	}

	span.SetAttributes(attributes...)

	err := libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.payload", filters)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert filters to JSON string", err)
	}

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

	spanFind.SetAttributes(attributes...)

	err = libOpentelemetry.SetSpanAttributesFromStruct(&spanFind, "app.request.repository_filter", filters)
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
	_, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongodb.find_by_entity")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_id", id.String()),
		attribute.String("app.request.organization_id", organizationID.String()),
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

	return &record.OutputFormat, nil
}

// Create inserts a new package entity into mongo.
func (tm *TemplateMongoDBRepository) Create(ctx context.Context, record *TemplateMongoDBModel) (*Template, error) {
	_, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongo.create_template")
	defer span.End()

	attributes := []attribute.KeyValue{
		attribute.String("app.request.request_id", reqId),
	}

	span.SetAttributes(attributes...)

	err := libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.payload", record)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert template record to JSON string", err)
	}

	db, err := tm.connection.GetDB(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get database", err)

		return nil, err
	}

	coll := db.Database(strings.ToLower(tm.Database)).Collection(strings.ToLower(constant.MongoCollectionTemplate))

	ctx, spanInsert := tracer.Start(ctx, "mongo.create_template.insert")

	spanInsert.SetAttributes(attributes...)

	err = libOpentelemetry.SetSpanAttributesFromStruct(&spanInsert, "app.request.repository_input", record)
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
	_, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongodb.update_template")
	defer span.End()

	attributes := []attribute.KeyValue{
		attribute.String("app.request.request_id", reqId),
	}

	span.SetAttributes(attributes...)

	err := libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.payload", updateFields)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert template record to JSON string", err)
	}

	db, err := tm.connection.GetDB(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get database", err)
		return err
	}

	coll := db.Database(strings.ToLower(tm.Database)).Collection(strings.ToLower(constant.MongoCollectionTemplate))
	opts := options.Update().SetUpsert(false)

	ctx, spanUpdate := tracer.Start(ctx, "mongodb.update_template.update_one")

	spanUpdate.SetAttributes(attributes...)

	err = libOpentelemetry.SetSpanAttributesFromStruct(&spanUpdate, "app.request.repository_input", updateFields)
	if err != nil {
		libOpentelemetry.HandleSpanError(&spanUpdate, "Failed to convert template record from entity to JSON string", err)
	}

	_, err = coll.UpdateOne(ctx, bson.M{"_id": id, "organization_id": organizationID}, updateFields, opts)
	if err != nil {
		libOpentelemetry.HandleSpanError(&spanUpdate, "Failed to update template", err)
		return err
	}

	spanUpdate.End()

	return nil
}

// Delete a template entity into mongodb with soft delete or not.
func (tm *TemplateMongoDBRepository) Delete(ctx context.Context, id, organizationID uuid.UUID, hardDelete bool) error {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongodb.delete_template")
	defer span.End()

	attributes := []attribute.KeyValue{
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_id", id.String()),
		attribute.String("app.request.organization_id", organizationID.String()),
	}

	span.SetAttributes(attributes...)

	db, err := tm.connection.GetDB(ctx)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to get database", err)

		return err
	}

	opts := options.Delete()

	coll := db.Database(strings.ToLower(tm.Database)).Collection(strings.ToLower(constant.MongoCollectionTemplate))

	ctx, spanDelete := tracer.Start(ctx, "mongodb.delete_template.delete_one")

	spanDelete.SetAttributes(attributes...)

	filter := bson.D{
		{Key: "_id", Value: id},
		{Key: "organization_id", Value: organizationID},
		{Key: "deleted_at", Value: nil},
	}

	if hardDelete {
		deleted, err := coll.DeleteOne(ctx, filter, opts)
		if err != nil {
			libOpentelemetry.HandleSpanError(&spanDelete, "Failed to delete template", err)

			return err
		}

		spanDelete.End()

		if deleted.DeletedCount == 0 {
			return pkg.ValidateBusinessError(constant.ErrEntityNotFound, "", constant.MongoCollectionTemplate)
		}
	} else {
		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "deleted_at", Value: time.Now()},
			}},
		}

		updateResult, err := coll.UpdateOne(ctx, filter, update)
		if err != nil {
			libOpentelemetry.HandleSpanError(&spanDelete, "Failed to soft delete template", err)

			return err
		}

		if updateResult.MatchedCount == 0 {
			return pkg.ValidateBusinessError(constant.ErrEntityNotFound, "", constant.MongoCollectionTemplate)
		}
	}

	spanDelete.End()

	logger.Infoln("Deleted a template with id: ", id.String(), " (hard delete: ", hardDelete, ")")

	return nil
}

// FindMappedFieldsAndOutputFormatByID find mapped fields of template and output format.
func (tm *TemplateMongoDBRepository) FindMappedFieldsAndOutputFormatByID(ctx context.Context, id, organizationID uuid.UUID) (*string, map[string]map[string][]string, error) {
	_, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongodb.find_mapped_fields_and_output_format_by_id")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_id", id.String()),
		attribute.String("app.request.organization_id", organizationID.String()),
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

	return &record.OutputFormat, record.MappedFields, nil
}
