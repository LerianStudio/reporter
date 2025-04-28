package mongodb

import (
	"context"
	"encoding/hex"
	libCommons "github.com/LerianStudio/lib-commons/commons"
	libMongo "github.com/LerianStudio/lib-commons/commons/mongo"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository defines an interface for querying data from MongoDB collections.
//
//go:generate mockgen --destination=datasource.mongodb.mock.go --package=mongodb . Repository
type Repository interface {
	Query(ctx context.Context, collection string, fields []string, filter map[string][]any) ([]map[string]any, error)
	GetDatabaseSchema(ctx context.Context) ([]CollectionSchema, error)
}

// CollectionSchema represents the structure of a MongoDB collection
type CollectionSchema struct {
	CollectionName string             `json:"collection_name"`
	Fields         []FieldInformation `json:"fields"`
}

// FieldInformation contains the details of a MongoDB field
type FieldInformation struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
}

// ExternalDataSource provides an interface for interacting with a MongoDB database connection.
type ExternalDataSource struct {
	connection *libMongo.MongoConnection
	Database   string
}

// NewDataSourceRepository creates a new ExternalDataSource instance using the provided MongoDB connection string and database name.
func NewDataSourceRepository(mongoURI string, dbName string) *ExternalDataSource {
	logger := libCommons.NewLoggerFromContext(context.Background()) // TODO

	mongoConnection := &libMongo.MongoConnection{
		ConnectionStringSource: mongoURI,
		Database:               dbName,
		MaxPoolSize:            100,
		Logger:                 logger,
	}

	if _, err := mongoConnection.GetDB(context.Background()); err != nil {
		panic(err)
	}

	return &ExternalDataSource{
		connection: mongoConnection,
		Database:   dbName,
	}
}

// Query executes a query on the specified collection with the given fields and filter criteria.
func (ds *ExternalDataSource) Query(ctx context.Context, collection string, fields []string, filter map[string][]any) ([]map[string]any, error) {
	logger := libCommons.NewLoggerFromContext(ctx)
	logger.Infof("Querying %s collection with fields %v", collection, fields)

	client, err := ds.connection.GetDB(ctx)
	if err != nil {
		return nil, err
	}

	// Convert filter to MongoDB format
	mongoFilter := bson.M{}

	for key, values := range filter {
		if len(values) == 1 {
			mongoFilter[key] = values[0]
		} else if len(values) > 1 {
			mongoFilter[key] = bson.M{"$in": values}
		}
	}

	// Create projection for specified fields
	projection := bson.M{}

	if len(fields) > 0 && fields[0] != "*" {
		for _, field := range fields {
			projection[field] = 1
		}
	}

	findOptions := options.Find()
	if len(projection) > 0 {
		findOptions.SetProjection(projection)
	}

	database := client.Database(ds.Database)

	cursor, err := database.Collection(collection).Find(ctx, mongoFilter, findOptions)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var results []map[string]any

	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			logger.Warnf("Error decoding document: %v", err)
			continue
		}

		resultMap := convertBsonToMap(result)
		results = append(results, resultMap)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// convertBsonToMap converts a bson.M to map[string]any recursively
// ensuring that all nested objects are also properly converted
func convertBsonToMap(bsonDoc bson.M) map[string]any {
	result := make(map[string]any)

	for k, v := range bsonDoc {
		result[k] = convertBsonValue(v)
	}

	return result
}

// convertBsonValue converts a BSON value to its Go equivalent recursively
func convertBsonValue(value any) any {
	switch v := value.(type) {
	case bson.M:
		// Nested object - convert recursively
		return convertBsonToMap(v)

	case bson.A:
		// Array - convert each element
		result := make([]any, len(v))
		for i, elem := range v {
			result[i] = convertBsonValue(elem)
		}

		return result

	case bson.D:
		// Ordered document - convert to map
		doc := make(map[string]any)
		for _, elem := range v {
			doc[elem.Key] = convertBsonValue(elem.Value)
		}

		return doc

	case primitive.DateTime:
		// Convert to time.Time for easier template usage
		return v.Time()

	case primitive.ObjectID:
		// Convert ObjectID to string
		return v.Hex()

	case primitive.Binary:
		// Check if Binary is a UUID
		if len(v.Data) == 16 {
			u, err := uuid.FromBytes(v.Data)
			if err == nil {
				return u.String() // Retorna UUID formatado
			}
		}

		// For non-UUID binary data or if UUID parsing fails, fall back to hex
		return hex.EncodeToString(v.Data)

	case nil:
		return nil

	default:
		// Other primitive types (string, int, float, bool, etc.) don't need conversion
		return v
	}
}

// GetDatabaseSchema retrieves all collections and infers their schema from sample documents
func (ds *ExternalDataSource) GetDatabaseSchema(ctx context.Context) ([]CollectionSchema, error) {
	logger := libCommons.NewLoggerFromContext(ctx)
	logger.Info("Retrieving MongoDB schema information")

	client, err := ds.connection.GetDB(ctx)
	if err != nil {
		return nil, err
	}

	database := client.Database(ds.Database)

	collections, err := database.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	schema := make([]CollectionSchema, 0, len(collections))

	for _, collName := range collections {
		coll := database.Collection(collName)

		// Sample a document to infer schema
		sampleDoc := bson.M{}

		cursor, err := coll.Find(ctx, bson.M{}, options.Find().SetLimit(1))
		if err != nil {
			logger.Warnf("Could not query collection %s: %v", collName, err)
			continue
		}

		hasDoc := false

		if cursor.Next(ctx) {
			if err := cursor.Decode(&sampleDoc); err == nil {
				hasDoc = true
			}
		}

		err = cursor.Close(ctx)
		if err != nil {
			return nil, err
		}

		collSchema := CollectionSchema{
			CollectionName: collName,
			Fields:         []FieldInformation{},
		}

		if hasDoc {
			for fieldName, value := range sampleDoc {
				dataType := "unknown"
				switch value.(type) {
				case string:
					dataType = "string"
				case int, int32, int64, float32, float64:
					dataType = "number"
				case bool:
					dataType = "boolean"
				case bson.A:
					dataType = "array"
				case bson.M, bson.D:
					dataType = "object"
				}

				collSchema.Fields = append(collSchema.Fields, FieldInformation{
					Name:     fieldName,
					DataType: dataType,
				})
			}
		}

		schema = append(schema, collSchema)
	}

	logger.Infof("Retrieved schema for %d collections", len(schema))

	return schema, nil
}
