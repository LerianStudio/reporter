package mongodb

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/LerianStudio/reporter/v3/pkg/constant"
	"github.com/LerianStudio/reporter/v3/pkg/model"

	"github.com/LerianStudio/lib-commons/v2/commons/log"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	libMongo "github.com/LerianStudio/lib-commons/v2/commons/mongo"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
// Returns nil and error if connection fails.
func NewDataSourceRepository(mongoURI string, dbName string, logger log.Logger) (*ExternalDataSource, error) {
	mongoConnection := &libMongo.MongoConnection{
		ConnectionStringSource: mongoURI,
		Database:               dbName,
		MaxPoolSize:            100,
		Logger:                 logger,
	}

	if _, err := mongoConnection.GetDB(context.Background()); err != nil {
		logger.Errorf("Failed to establish MongoDB connection: %v", err)
		return nil, fmt.Errorf("failed to establish MongoDB connection: %w", err)
	}

	return &ExternalDataSource{
		connection: mongoConnection,
		Database:   dbName,
	}, nil
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
	)

	err := libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.repository_filter", map[string]any{
		"collection": collection,
		"fields":     fields,
		"filter":     filter,
	})
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert repository filter to JSON string", err)
	}

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

	// Create timeout context for query execution
	queryCtx, cancel := context.WithTimeout(ctx, constant.QueryTimeoutMedium)
	defer cancel()

	database := client.Database(ds.Database)

	cursor, err := database.Collection(collection).Find(queryCtx, mongoFilter, findOptions)
	if err != nil {
		if queryCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("mongodb query timeout after %v for collection %s: %w", constant.QueryTimeoutMedium, collection, err)
		}

		return nil, err
	}

	defer cursor.Close(queryCtx)

	var results []map[string]any

	for cursor.Next(queryCtx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			logger.Warnf("Error decoding document: %v", err)
			continue
		}

		resultMap := convertBsonToMap(result)
		results = append(results, resultMap)
	}

	if err := cursor.Err(); err != nil {
		if queryCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("mongodb query result iteration timeout after %v for collection %s: %w", constant.QueryTimeoutMedium, collection, err)
		}

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

	logger.Info("Retrieving MongoDB schema information using hybrid approach")

	// Create timeout context for schema discovery (longer timeout for this operation)
	schemaCtx, cancel := context.WithTimeout(ctx, constant.SchemaDiscoveryTimeout)
	defer cancel()

	client, err := ds.connection.GetDB(ctx)
	if err != nil {
		return nil, err
	}

	database := client.Database(ds.Database)

	collections, err := database.ListCollectionNames(schemaCtx, bson.M{})
	if err != nil {
		if schemaCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("mongodb schema discovery timeout after %v while listing collections: %w", constant.SchemaDiscoveryTimeout, err)
		}

		return nil, err
	}

	schema := make([]CollectionSchema, 0, len(collections))

	for _, collName := range collections {
		coll := database.Collection(collName)

		logger.Infof("Analyzing collection: %s", collName)

		allFields, err := ds.discoverAllFieldsWithAggregation(schemaCtx, coll)
		if err != nil {
			logger.Warnf("Aggregation failed for collection %s, falling back to sampling: %v", collName, err)

			allFields = make(map[string]bool)
		}

		fieldTypes, additionalFields, err := ds.sampleMultipleDocuments(schemaCtx, coll)
		if err != nil {
			logger.Warnf("Document sampling failed for collection %s: %v", collName, err)

			fieldTypes = make(map[string]string)
			additionalFields = make(map[string]bool)
		}

		for field := range additionalFields {
			allFields[field] = true
		}

		collSchema := CollectionSchema{
			CollectionName: collName,
			Fields:         []FieldInformation{},
		}

		for fieldName := range allFields {
			dataType := fieldTypes[fieldName]
			if dataType == "" {
				dataType = "unknown"
			}

			collSchema.Fields = append(collSchema.Fields, FieldInformation{
				Name:     fieldName,
				DataType: dataType,
			})
		}

		logger.Infof("Discovered %d fields in collection %s", len(collSchema.Fields), collName)
		schema = append(schema, collSchema)
	}

	logger.Infof("Retrieved schema for %d collections", len(schema))

	return schema, nil
}

// discoverAllFieldsWithAggregation uses MongoDB aggregation with sampling for large collections
func (ds *ExternalDataSource) discoverAllFieldsWithAggregation(ctx context.Context, coll *mongo.Collection) (map[string]bool, error) {
	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	// For large collections (>10k docs), use sampling instead of full aggregation
	if count > 10000 {
		return ds.discoverFieldsWithSampling(ctx, coll, count)
	}

	// For small collections, use optimized aggregation with $limit
	pipeline := []bson.M{
		// Limit processing to a reasonable sample size even for small collections
		{
			"$limit": func() int64 {
				if count > 1000 {
					return 1000
				}
				return count
			}(),
		},
		{
			"$project": bson.M{
				"arrayofkeyvalue": bson.M{"$objectToArray": "$$ROOT"},
			},
		},
		{
			"$unwind": "$arrayofkeyvalue",
		},
		{
			"$group": bson.M{
				"_id":     nil,
				"allkeys": bson.M{"$addToSet": "$arrayofkeyvalue.k"},
			},
		},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	allFields := make(map[string]bool)

	if cursor.Next(ctx) {
		var result struct {
			AllKeys []string `bson:"allkeys"`
		}

		if err := cursor.Decode(&result); err == nil {
			for _, key := range result.AllKeys {
				allFields[key] = true
			}
		}
	}

	return allFields, nil
}

// discoverFieldsWithSampling uses intelligent sampling for large collections
func (ds *ExternalDataSource) discoverFieldsWithSampling(ctx context.Context, coll *mongo.Collection, totalDocs int64) (map[string]bool, error) {
	sampleSize := ds.calculateOptimalSampleSize(totalDocs)
	pipeline := []bson.M{
		{
			"$sample": bson.M{"size": sampleSize},
		},
		{
			"$project": bson.M{
				"arrayofkeyvalue": bson.M{"$objectToArray": "$$ROOT"},
			},
		},
		{
			"$unwind": "$arrayofkeyvalue",
		},
		{
			"$group": bson.M{
				"_id":     nil,
				"allkeys": bson.M{"$addToSet": "$arrayofkeyvalue.k"},
			},
		},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	allFields := make(map[string]bool)

	if cursor.Next(ctx) {
		var result struct {
			AllKeys []string `bson:"allkeys"`
		}

		if err := cursor.Decode(&result); err == nil {
			for _, key := range result.AllKeys {
				allFields[key] = true
			}
		}
	}

	return allFields, nil
}

// calculateOptimalSampleSize calculates the optimal sample size based on collection size
func (ds *ExternalDataSource) calculateOptimalSampleSize(totalDocs int64) int {
	// Statistical sampling: 95% confidence, 5% margin of error
	// For schema discovery, we don't need perfect accuracy, just good coverage
	switch {
	case totalDocs <= 1000:
		return int(totalDocs) // Use all documents for small collections
	case totalDocs <= 10000:
		return 1000 // 10% sample
	case totalDocs <= 100000:
		return 2000 // 2% sample
	case totalDocs <= 1000000:
		return 5000 // 0.5% sample
	default:
		return 10000 // 0.1% sample for very large collections
	}
}

// sampleMultipleDocuments samples multiple documents to discover field types and additional fields
func (ds *ExternalDataSource) sampleMultipleDocuments(ctx context.Context, coll *mongo.Collection) (map[string]string, map[string]bool, error) {
	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, nil, err
	}

	sampleSize := 50
	if count < 50 {
		sampleSize = int(count)
	}

	var cursor *mongo.Cursor

	if count > 1000 {
		pipeline := []bson.M{
			{"$sample": bson.M{"size": sampleSize}},
		}
		cursor, err = coll.Aggregate(ctx, pipeline)
	} else {
		cursor, err = coll.Find(ctx, bson.M{}, options.Find().SetLimit(int64(sampleSize)))
	}

	if err != nil {
		return nil, nil, err
	}

	defer cursor.Close(ctx)

	fieldTypes := make(map[string]string)
	allFields := make(map[string]bool)

	docCount := 0

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		docCount++

		for fieldName, value := range doc {
			allFields[fieldName] = true
			dataType := ds.inferDataType(value)

			if currentType, exists := fieldTypes[fieldName]; !exists || ds.isMoreSpecificType(dataType, currentType) {
				fieldTypes[fieldName] = dataType
			}
		}
	}

	return fieldTypes, allFields, nil
}

// inferDataType determines the MongoDB data type from a Go value
func (ds *ExternalDataSource) inferDataType(value any) string {
	switch value.(type) {
	case string:
		return "string"
	case int, int32, int64, float32, float64:
		return "number"
	case bool:
		return "boolean"
	case bson.A:
		return "array"
	case bson.M, bson.D:
		return "object"
	case primitive.DateTime:
		return "date"
	case primitive.ObjectID:
		return "objectId"
	case primitive.Binary:
		return "binData"
	case primitive.Regex:
		return "regex"
	case primitive.Timestamp:
		return "timestamp"
	case primitive.Decimal128:
		return "decimal"
	case primitive.MinKey, primitive.MaxKey:
		return "minKey/maxKey"
	default:
		return "unknown"
	}
}

// isMoreSpecificType determines if one type is more specific than another
func (ds *ExternalDataSource) isMoreSpecificType(newType, currentType string) bool {
	typeHierarchy := map[string]int{
		"objectId":      10,
		"date":          9,
		"timestamp":     8,
		"decimal":       7,
		"binData":       6,
		"regex":         5,
		"minKey/maxKey": 4,
		"number":        3,
		"string":        2,
		"boolean":       2,
		"array":         2,
		"object":        2,
		"unknown":       1,
	}

	newLevel := typeHierarchy[newType]
	currentLevel := typeHierarchy[currentType]

	return newLevel > currentLevel
}

// QueryWithAdvancedFilters executes a query with advanced FilterCondition support
func (ds *ExternalDataSource) QueryWithAdvancedFilters(ctx context.Context, collection string, fields []string, filter map[string]model.FilterCondition) ([]map[string]any, error) {
	logger, tracer, reqId, _ := libCommons.NewTrackingFromContext(ctx)

	logger.Infof("Querying %s collection with advanced filters on fields %v", collection, fields)

	_, span := tracer.Start(ctx, "mongodb.data_source.query_with_advanced_filters")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
	)

	err := libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.repository_filter", map[string]any{
		"collection": collection,
		"fields":     fields,
		"filter":     filter,
	})
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert repository filter to JSON string", err)
	}

	client, err := ds.connection.GetDB(ctx)
	if err != nil {
		return nil, err
	}

	mongoFilter, err := ds.buildMongoFilter(filter)
	if err != nil {
		return nil, err
	}

	findOptions := ds.buildFindOptions(fields)

	cursor, queryCtx, cancel, err := ds.executeFindQuery(ctx, client, collection, mongoFilter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cancel()
	defer cursor.Close(queryCtx)

	return ds.processQueryResults(queryCtx, cursor, collection, logger)
}

// buildMongoFilter converts FilterCondition map to MongoDB filter format
func (ds *ExternalDataSource) buildMongoFilter(filter map[string]model.FilterCondition) (bson.M, error) {
	mongoFilter := bson.M{}

	for field, condition := range filter {
		if isFilterConditionEmpty(condition) {
			continue
		}

		fieldFilter, err := ds.convertFilterConditionToMongoFilter(field, condition)
		if err != nil {
			return nil, fmt.Errorf("error converting filter for field '%s': %w", field, err)
		}

		for k, v := range fieldFilter {
			mongoFilter[k] = v
		}
	}

	return mongoFilter, nil
}

// buildFindOptions creates MongoDB find options with field projection
func (ds *ExternalDataSource) buildFindOptions(fields []string) *options.FindOptions {
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

	return findOptions
}

// executeFindQuery executes the MongoDB find query with timeout
func (ds *ExternalDataSource) executeFindQuery(
	ctx context.Context,
	client *mongo.Client,
	collection string,
	mongoFilter bson.M,
	findOptions *options.FindOptions,
) (*mongo.Cursor, context.Context, context.CancelFunc, error) {
	queryCtx, cancel := context.WithTimeout(ctx, constant.QueryTimeoutSlow)

	database := client.Database(ds.Database)

	cursor, err := database.Collection(collection).Find(queryCtx, mongoFilter, findOptions)
	if err != nil {
		cancel()

		if queryCtx.Err() == context.DeadlineExceeded {
			return nil, nil, nil, fmt.Errorf("mongodb advanced filter query timeout after %v for collection %s: %w", constant.QueryTimeoutSlow, collection, err)
		}

		return nil, nil, nil, err
	}

	return cursor, queryCtx, cancel, nil
}

// processQueryResults iterates through cursor and converts results
func (ds *ExternalDataSource) processQueryResults(
	queryCtx context.Context,
	cursor *mongo.Cursor,
	collection string,
	logger log.Logger,
) ([]map[string]any, error) {
	var results []map[string]any

	for cursor.Next(queryCtx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			logger.Warnf("Error decoding document: %v", err)
			continue
		}

		resultMap := convertBsonToMap(result)
		results = append(results, resultMap)
	}

	if err := cursor.Err(); err != nil {
		if queryCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("mongodb advanced filter result iteration timeout after %v for collection %s: %w", constant.QueryTimeoutSlow, collection, err)
		}

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
