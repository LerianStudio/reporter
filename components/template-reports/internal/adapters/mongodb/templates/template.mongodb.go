package templates

import (
	"context"
	pkgMongo "plugin-template-engine/pkg/mongo"
)

// Repository provides an interface for operations related on mongo a metadata entities.
//
//go:generate mockgen --destination=../../../mocks/mongodb/templates/template_mondogdb_mock.go --package=templates . Repository
type Repository interface {
	Create(ctx context.Context, collection string, ex *Example) (*Example, error)
}

// TemplateMongoDBRepository is a MongoDD-specific implementation of the PackageRepository.
type TemplateMongoDBRepository struct {
	connection *pkgMongo.MongoConnection
	Database   string
}

// NewTemplateMongoDBRepository returns a new instance of TemplateMongoDBRepository using the given MongoDB connection.
func NewTemplateMongoDBRepository(mc *pkgMongo.MongoConnection) *TemplateMongoDBRepository {
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
func (pm *TemplateMongoDBRepository) Create(ctx context.Context, collection string, ex *Example) (*Example, error) {
	return nil, nil
}
