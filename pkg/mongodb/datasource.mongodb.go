package mongodb

import (
	"context"
	"encoding/hex"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	"fmt"
	"plugin-smart-templates/v2/pkg/model"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	libMongo "github.com/LerianStudio/lib-commons/v2/commons/mongo"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/attribute"
)

// Repository defines an interface for querying data from MongoDB collections.
//
//go:generate mockgen --destination=datasource.mongodb.mock.go --package=mongodb . Repository
type Repository interface {
	Query(ctx context.Context, collection string, fields []string, filter map[string][]any) ([]map[string]any, error)
	QueryWithAdvancedFilters(ctx context.Context, collection string, fields []string, filter map[string]model.FilterCondition) ([]map[string]any, error)
	GetDatabaseSchema(ctx context.Context) ([]CollectionSchema, error)
	CloseConnection(ctx context.Context) error
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
func NewDataSourceRepository(mongoURI string, dbName string, logger log.Logger) *ExternalDataSource {
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

// CloseConnection close the connection with MongoDB.
func (ds *ExternalDataSource) CloseConnection(ctx context.Context) error {
	if ds.connection.DB != nil {
		ds.connection.Logger.Info("Closing MongoDB connection...")

		err := ds.connection.DB.Disconnect(ctx)
		if err != nil {
			ds.connection.Logger.Errorf("Error closing MongoDB connection: %v", err)
			return err
		}

		ds.connection.DB = nil
		ds.connection.Connected = false

		ds.connection.Logger.Info("MongoDB connection closed successfully.")
	}

	return nil
}

// Query executes a query on the specified collection with the given fields and filter criteria.
func (ds *ExternalDataSource) Query(ctx context.Context, collection string, fields []string, filter map[string][]any) ([]map[string]any, error) {
	logger, tracer, reqId, _ := libCommons.NewTrackingFromContext(ctx)

	logger.Infof("Querying %s collection with fields %v", collection, fields)

	_, span := tracer.Start(ctx, "mongodb.data_source.query")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.collection", collection),
		attribute.StringSlice("app.request.fields", fields),
	)

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
	logger, tracer, reqId, _ := libCommons.NewTrackingFromContext(ctx)

	_, span := tracer.Start(ctx, "mongodb.data_source.get_database_schema")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

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

// QueryWithAdvancedFilters executes a query with advanced FilterCondition support
func (ds *ExternalDataSource) QueryWithAdvancedFilters(ctx context.Context, collection string, fields []string, filter map[string]model.FilterCondition) ([]map[string]any, error) {
	logger := libCommons.NewLoggerFromContext(ctx)
	tracer := libCommons.NewTracerFromContext(ctx)
	reqId := libCommons.NewHeaderIDFromContext(ctx)

	logger.Infof("Querying %s collection with advanced filters on fields %v", collection, fields)

	_, span := tracer.Start(ctx, "mongodb.data_source.query_with_advanced_filters")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.collection", collection),
		attribute.StringSlice("app.request.fields", fields),
	)

	client, err := ds.connection.GetDB(ctx)
	if err != nil {
		return nil, err
	}

	// Convert FilterCondition to MongoDB format
	mongoFilter := bson.M{}

	for field, condition := range filter {
		if isFilterConditionEmpty(condition) {
			continue
		}

		fieldFilter, err := ds.convertFilterConditionToMongoFilter(field, condition)
		if err != nil {
			return nil, fmt.Errorf("error converting filter for field '%s': %w", field, err)
		}

		if fieldFilter != nil {
			for k, v := range fieldFilter {
				mongoFilter[k] = v
			}
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

// convertFilterConditionToMongoFilter converts a FilterCondition to MongoDB filter
func (ds *ExternalDataSource) convertFilterConditionToMongoFilter(field string, condition model.FilterCondition) (map[string]any, error) {
	if isFilterConditionEmpty(condition) {
		return nil, nil
	}

	if err := ds.validateFilterCondition(field, condition); err != nil {
		return nil, err
	}

	filter := make(map[string]any)
	fieldFilter := make(map[string]any)

	// Handle equals
	if len(condition.Equals) > 0 {
		if len(condition.Equals) == 1 {
			filter[field] = condition.Equals[0]
		} else {
			fieldFilter["$in"] = condition.Equals
		}
	}

	// Handle greater than
	if len(condition.GreaterThan) > 0 {
		fieldFilter["$gt"] = condition.GreaterThan[0]
	}

	// Handle greater than or equal
	if len(condition.GreaterOrEqual) > 0 {
		fieldFilter["$gte"] = condition.GreaterOrEqual[0]
	}

	// Handle less than
	if len(condition.LessThan) > 0 {
		fieldFilter["$lt"] = condition.LessThan[0]
	}

	// Handle less than or equal
	if len(condition.LessOrEqual) > 0 {
		fieldFilter["$lte"] = condition.LessOrEqual[0]
	}

	// Handle between (using $gte and $lte)
	if len(condition.Between) > 0 {
		fieldFilter["$gte"] = condition.Between[0]
		fieldFilter["$lte"] = condition.Between[1]
	}

	// Handle in
	if len(condition.In) > 0 {
		fieldFilter["$in"] = condition.In
	}

	// Handle not in
	if len(condition.NotIn) > 0 {
		fieldFilter["$nin"] = condition.NotIn
	}

	// If we have complex field filters, use them, otherwise use the simple filter
	if len(fieldFilter) > 0 {
		filter[field] = fieldFilter
	}

	return filter, nil
}

// isFilterConditionEmpty checks if a FilterCondition has no active filters
func isFilterConditionEmpty(condition model.FilterCondition) bool {
	return len(condition.Equals) == 0 &&
		len(condition.GreaterThan) == 0 &&
		len(condition.GreaterOrEqual) == 0 &&
		len(condition.LessThan) == 0 &&
		len(condition.LessOrEqual) == 0 &&
		len(condition.Between) == 0 &&
		len(condition.In) == 0 &&
		len(condition.NotIn) == 0
}

// validateFilterCondition validates that a FilterCondition has proper values for each operator
func (ds *ExternalDataSource) validateFilterCondition(fieldName string, condition model.FilterCondition) error {
	// Validate between operator has exactly 2 values
	if len(condition.Between) > 0 && len(condition.Between) != 2 {
		return fmt.Errorf("between operator for field '%s' must have exactly 2 values, got %d", fieldName, len(condition.Between))
	}

	// Validate single-value operators have exactly 1 value
	singleValueOps := map[string][]any{
		"gt":  condition.GreaterThan,
		"gte": condition.GreaterOrEqual,
		"lt":  condition.LessThan,
		"lte": condition.LessOrEqual,
	}

	for opName, values := range singleValueOps {
		if len(values) > 0 && len(values) != 1 {
			return fmt.Errorf("%s operator for field '%s' must have exactly 1 value, got %d", opName, fieldName, len(values))
		}
	}

	return nil
}
