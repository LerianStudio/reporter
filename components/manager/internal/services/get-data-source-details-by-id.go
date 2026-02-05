// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/constant"
	"github.com/LerianStudio/reporter/v4/pkg/model"
	"github.com/LerianStudio/reporter/v4/pkg/mongodb"

	"github.com/LerianStudio/lib-commons/v2/commons"
	libConstants "github.com/LerianStudio/lib-commons/v2/commons/constants"
	"github.com/LerianStudio/lib-commons/v2/commons/log"
	"go.opentelemetry.io/otel/attribute"
)

var (
	// Define encrypted fields that should be excluded
	encryptedFields = map[string]bool{
		"document": true,
		"name":     true,
	}

	// Define search fields that should be included (these are hashes, not encrypted)
	searchFields = map[string]bool{
		"search.document":                true,
		"search.banking_details_account": true,
		"search.banking_details_iban":    true,
		"search.contact_primary_email":   true,
		"search.contact_secondary_email": true,
		"search.contact_mobile_phone":    true,
		"search.contact_other_phone":     true,
		"search":                         true, // Include the search object itself
	}

	// Define nested encrypted fields that should be excluded
	nestedEncryptedFields = map[string]bool{
		"contact.primary_email":                  true,
		"contact.secondary_email":                true,
		"contact.mobile_phone":                   true,
		"contact.other_phone":                    true,
		"banking_details.account":                true,
		"banking_details.iban":                   true,
		"legal_person.representative.name":       true,
		"legal_person.representative.document":   true,
		"legal_person.representative.email":      true,
		"natural_person.mother_name":             true,
		"natural_person.father_name":             true,
		"regulatory_fields.participant_document": true,
		"related_parties.document":               true,
		"related_parties.name":                   true,
	}
)

// GetDataSourceDetailsByID retrieves the data source information by data source id
func (uc *UseCase) GetDataSourceDetailsByID(ctx context.Context, dataSourceID string) (*model.DataSourceDetails, error) {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "get_data_source_details_by_id")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.data_source_id", dataSourceID),
	)

	logger.Infof("Retrieving data source details for id %v", dataSourceID)

	cacheKey := constant.DataSourceDetailsKeyPrefix + ":" + dataSourceID
	if cached, ok := uc.getDataSourceDetailsFromCache(ctx, cacheKey); ok {
		logger.Infof("Cache hit for data source details id %v", dataSourceID)
		return cached, nil
	}

	dataSource, ok := uc.ExternalDataSources[dataSourceID]
	if !ok {
		return nil, pkg.ValidateBusinessError(constant.ErrMissingDataSource, "", dataSourceID)
	}

	if err := uc.ensureDataSourceConnected(logger, dataSourceID, &dataSource); err != nil {
		logger.Errorf("Error initializing database connection for '%s', Err: %s", dataSourceID, err)
		return nil, err
	}

	var (
		result           *model.DataSourceDetails
		errGetDataSource error
	)

	switch dataSource.DatabaseType {
	case pkg.PostgreSQLType:
		result, errGetDataSource = uc.getDataSourceDetailsOfPostgresDatabase(ctx, logger, dataSourceID, dataSource)

		errClose := dataSource.PostgresRepository.CloseConnection()
		if errClose != nil {
			logger.Errorf("Error to close postgres connection, Err: %s", errClose)
			return nil, errClose
		}
	case pkg.MongoDBType:
		result, errGetDataSource = uc.getDataSourceDetailsOfMongoDBDatabase(ctx, logger, dataSourceID, dataSource)

		errClose := dataSource.MongoDBRepository.CloseConnection(ctx)
		if errClose != nil {
			return nil, errClose
		}
	default:
		return nil, pkg.ValidateBusinessError(constant.ErrMissingDataSource, "", dataSourceID)
	}

	if errGetDataSource != nil {
		logger.Errorf("Error to get data source details, Err: %s", errGetDataSource)
		return nil, pkg.ValidateBusinessError(constant.ErrMissingDataSource, "", dataSourceID)
	}

	logger.Info("Close the connection to the database.")

	errSet := uc.setDataSourceDetailsToCache(ctx, cacheKey, result)
	if errSet != nil {
		logger.Errorf("Error to set data source details to cache, Err: %s", errSet)
		return nil, errSet
	}

	return result, nil
}

// getDataSourceDetailsFromCache tries to get and unmarshal DataSourceDetails from Redis
func (uc *UseCase) getDataSourceDetailsFromCache(ctx context.Context, cacheKey string) (*model.DataSourceDetails, bool) {
	if uc.RedisRepo == nil {
		return nil, false
	}

	cached, err := uc.RedisRepo.Get(ctx, cacheKey)
	if err != nil || cached == "" {
		return nil, false
	}

	var details model.DataSourceDetails
	if err := json.Unmarshal([]byte(cached), &details); err != nil {
		return nil, false
	}

	return &details, true
}

// setDataSourceDetailsToCache marshals and sets DataSourceDetails in Redis
func (uc *UseCase) setDataSourceDetailsToCache(ctx context.Context, cacheKey string, details *model.DataSourceDetails) error {
	if uc.RedisRepo == nil || details == nil {
		return nil
	}

	if marshaled, err := json.Marshal(details); err == nil {
		if errCache := uc.RedisRepo.Set(ctx, cacheKey, string(marshaled), time.Second*constant.RedisTTL); errCache != nil {
			return errCache
		}
	}

	return nil
}

// ensureDataSourceConnected ensures the data source is initialized/connected
func (uc *UseCase) ensureDataSourceConnected(logger log.Logger, dataSourceID string, dataSource *pkg.DataSource) error {
	// Check if datasource is marked as unavailable
	if dataSource.Status == libConstants.DataSourceStatusUnavailable {
		logger.Warnf("Datasource '%s' is marked as unavailable - attempting to connect anyway", dataSourceID)
	}

	switch dataSource.DatabaseType {
	case pkg.PostgreSQLType:
		if !dataSource.Initialized || !dataSource.DatabaseConfig.Connected {
			logger.Infof("Connecting to PostgreSQL datasource '%s' on-demand...", dataSourceID)
			return pkg.ConnectToDataSource(dataSourceID, dataSource, logger, uc.ExternalDataSources)
		}
	case pkg.MongoDBType:
		if !dataSource.Initialized {
			logger.Infof("Connecting to MongoDB datasource '%s' on-demand...", dataSourceID)
			return pkg.ConnectToDataSource(dataSourceID, dataSource, logger, uc.ExternalDataSources)
		}
	}

	return nil
}

// getDataSourceDetailsOfMongoDBDatabase retrieves the data source information of a MongoDB database
func (uc *UseCase) getDataSourceDetailsOfMongoDBDatabase(ctx context.Context, logger log.Logger, dataSourceID string, dataSource pkg.DataSource) (*model.DataSourceDetails, error) {
	var (
		schema []mongodb.CollectionSchema
		err    error
	)

	schema, err = dataSource.MongoDBRepository.GetDatabaseSchema(ctx)
	if err != nil {
		logger.Errorf("Error get schemas of mongo db: %s", err.Error())
		return nil, err
	}

	tableDetails := uc.processCollectionsForDataSource(schema, dataSourceID)

	result := &model.DataSourceDetails{
		Id:           dataSourceID,
		ExternalName: dataSource.MongoDBName,
		Type:         dataSource.DatabaseType,
		Tables:       tableDetails,
	}

	return result, nil
}

// processCollectionsForDataSource processes collections and returns table details
func (uc *UseCase) processCollectionsForDataSource(schema []mongodb.CollectionSchema, dataSourceID string) []model.TableDetails {
	tableDetails := make([]model.TableDetails, 0)

	for _, collection := range schema {
		fields := uc.getFieldsForCollection(collection, dataSourceID)
		displayName := uc.getDisplayNameForCollection(collection.CollectionName, dataSourceID)

		tableSchema := model.TableDetails{
			Name:   displayName,
			Fields: fields,
		}

		tableDetails = append(tableDetails, tableSchema)
	}

	return tableDetails
}

// getFieldsForCollection determines which fields to include for a collection
func (uc *UseCase) getFieldsForCollection(collection mongodb.CollectionSchema, dataSourceID string) []string {
	if dataSourceID == "plugin_crm" {
		return uc.getFieldsForPluginCRM(collection)
	}

	// For other databases, include all fields
	fields := make([]string, 0, len(collection.Fields))
	for _, collectionField := range collection.Fields {
		fields = append(fields, collectionField.Name)
	}

	return fields
}

// getFieldsForPluginCRM gets fields for plugin_crm collections with special handling
func (uc *UseCase) getFieldsForPluginCRM(collection mongodb.CollectionSchema) []string {
	baseCollectionName := uc.getBaseCollectionName(collection.CollectionName)

	expandedFields := uc.getExpandedFieldsForPluginCRM(baseCollectionName)
	if expandedFields != nil {
		return expandedFields
	}

	// Fallback to filtering raw schema fields
	fields := make([]string, 0)

	for _, collectionField := range collection.Fields {
		if uc.shouldIncludeFieldForPluginCRM(collectionField.Name, baseCollectionName) {
			fields = append(fields, collectionField.Name)
		}
	}

	return fields
}

// getBaseCollectionName extracts the base collection name by removing organization suffix
func (uc *UseCase) getBaseCollectionName(collectionName string) string {
	if !strings.Contains(collectionName, "_") {
		return collectionName
	}

	parts := strings.Split(collectionName, "_")
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], "_")
	}

	return collectionName
}

// getDisplayNameForCollection gets the display name for a collection
func (uc *UseCase) getDisplayNameForCollection(collectionName, dataSourceID string) string {
	if dataSourceID == "plugin_crm" {
		return uc.getBaseCollectionName(collectionName)
	}

	return collectionName
}

// shouldIncludeFieldForPluginCRM determines if a field should be included for plugin_crm based on encryption status
func (uc *UseCase) shouldIncludeFieldForPluginCRM(fieldName, collectionName string) bool {
	// Check if it's a search field (include these)
	if searchFields[fieldName] {
		return true
	}

	// Check if it's a top-level encrypted field (exclude these)
	if encryptedFields[fieldName] {
		return false
	}

	// Check if it's a nested encrypted field (exclude these)
	if nestedEncryptedFields[fieldName] {
		return false
	}

	// For holders and aliases collections, be more specific about what to include
	if collectionName == "holders" || collectionName == "aliases" {
		// Include all non-encrypted fields
		return true
	}

	// For other collections, include all fields
	return true
}

// getExpandedFieldsForPluginCRM returns the expanded field list for plugin_crm collections
func (uc *UseCase) getExpandedFieldsForPluginCRM(collectionName string) []string {
	switch collectionName {
	case "holders":
		return []string{
			"_id",
			"external_id",
			"type",
			"addresses",
			"created_at",
			"updated_at",
			"deleted_at",
			"metadata",
			"search.document",
			// Natural person fields (non-encrypted)
			"natural_person.favorite_name",
			"natural_person.social_name",
			"natural_person.gender",
			"natural_person.birth_date",
			"natural_person.civil_status",
			"natural_person.nationality",
			"natural_person.status",
			// Legal person fields (non-encrypted)
			"legal_person.trade_name",
			"legal_person.activity",
			"legal_person.type",
			"legal_person.founding_date",
			"legal_person.size",
			"legal_person.status",
			"legal_person.representative.role",
		}
	case "aliases":
		return []string{
			"_id",
			"account_id",
			"holder_id",
			"ledger_id",
			"type",
			"created_at",
			"updated_at",
			"deleted_at",
			"metadata",
			// Search fields (hashes, not encrypted)
			"search.document",
			"search.banking_details_account",
			"search.banking_details_iban",
			"search.regulatory_fields_participant_document",
			"search.related_party_documents",
			// Banking details fields (non-encrypted)
			"banking_details.branch",
			"banking_details.type",
			"banking_details.opening_date",
			"banking_details.closing_date",
			"banking_details.country_code",
			"banking_details.bank_id",
			// Regulatory fields (non-encrypted)
			"regulatory_fields",
			// Related parties fields (non-encrypted)
			"related_parties",
			"related_parties._id",
			"related_parties.role",
			"related_parties.start_date",
			"related_parties.end_date",
		}
	default:
		return nil
	}
}

// getDataSourceDetailsOfPostgresDatabase retrieves the data source information of a PostgresSQL database
func (uc *UseCase) getDataSourceDetailsOfPostgresDatabase(ctx context.Context, logger log.Logger, dataSourceID string, dataSource pkg.DataSource) (*model.DataSourceDetails, error) {
	// Use configured schemas or default to public
	configuredSchemas := dataSource.Schemas
	if len(configuredSchemas) == 0 {
		configuredSchemas = []string{"public"}
	}

	schemas, err := dataSource.PostgresRepository.GetDatabaseSchema(ctx, configuredSchemas)
	if err != nil {
		logger.Errorf("Error get schemas of postgres: %s", err.Error())

		return nil, err
	}

	tableDetails := make([]model.TableDetails, 0)

	for _, tableSchema := range schemas {
		fields := make([]string, 0)
		for _, field := range tableSchema.Columns {
			fields = append(fields, field.Name)
		}

		tableDetail := model.TableDetails{
			Name:   tableSchema.QualifiedName(), // Returns "schema.table" format
			Fields: fields,
		}

		tableDetails = append(tableDetails, tableDetail)
	}

	result := &model.DataSourceDetails{
		Id:           dataSourceID,
		ExternalName: dataSource.MongoDBName,
		Type:         dataSource.DatabaseType,
		Tables:       tableDetails,
	}

	return result, nil
}
