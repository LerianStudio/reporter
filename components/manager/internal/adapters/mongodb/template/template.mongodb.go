package template

import (
	"context"
	"github.com/LerianStudio/lib-commons/commons"
	libMongo "github.com/LerianStudio/lib-commons/commons/mongo"
	"github.com/LerianStudio/lib-commons/commons/opentelemetry"
	"github.com/google/uuid"
	"strings"
)

// Repository provides an interface for operations related on mongo a metadata entities.
//
//go:generate mockgen --destination=../../../mocks/mongodb/templates/template_mondogdb_mock.go --package=templates . Repository
type Repository interface {
	Create(ctx context.Context, collection string, t *Template, organizationID uuid.UUID) (*Template, error)
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

// Create inserts a new package entity into mongo.
func (tm *TemplateMongoDBRepository) Create(ctx context.Context, collection string, t *Template, organizationID uuid.UUID) (*Template, error) {
	tracer := commons.NewTracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "mongo.create_template")
	defer span.End()

	db, err := tm.connection.GetDB(ctx)
	if err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to get database", err)

		return nil, err
	}

	coll := db.Database(strings.ToLower(tm.Database)).Collection(strings.ToLower(collection))
	record := &TemplateMongoDBModel{}

	if err := record.FromEntity(t, organizationID); err != nil {
		opentelemetry.HandleSpanError(&span, "Failed to convert template to model", err)

		return nil, err
	}

	ctx, spanInsert := tracer.Start(ctx, "mongo.create_template.insert")

	_, err = coll.InsertOne(ctx, record)
	if err != nil {
		opentelemetry.HandleSpanError(&spanInsert, "Failed to insert template", err)

		return nil, err
	}

	spanInsert.End()

	return record.ToEntity(), nil
}
