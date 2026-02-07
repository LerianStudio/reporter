// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/model"
	"github.com/LerianStudio/reporter/pkg/mongodb/report"

	"github.com/LerianStudio/lib-commons/v2/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/attribute"
)

// CreateReport create a new report
func (uc *UseCase) CreateReport(ctx context.Context, reportInput *model.CreateReportInput) (*report.Report, error) {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.report.create")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_id", reportInput.TemplateID),
	)

	err := libOpentelemetry.SetSpanAttributesFromStruct(&span, "app.request.payload", reportInput)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to convert report input to JSON string", err)
	}

	logger.Infof("Creating report")

	// Idempotency check: acquire lock via Redis SetNX before proceeding
	if uc.RedisRepo != nil {
		idempotencyKey, keyErr := uc.buildIdempotencyKey(ctx, reportInput)
		if keyErr != nil {
			libOpentelemetry.HandleSpanError(&span, "Failed to compute idempotency key", keyErr)

			return nil, keyErr
		}

		span.SetAttributes(attribute.String("app.idempotency.key", idempotencyKey))

		logger.Infof("Checking idempotency for key: %s", idempotencyKey)

		acquired, setNXErr := uc.RedisRepo.SetNX(ctx, idempotencyKey, "processing", constant.IdempotencyTTL)
		if setNXErr != nil {
			libOpentelemetry.HandleSpanError(&span, "Failed to acquire idempotency lock", setNXErr)

			return nil, setNXErr
		}

		if !acquired {
			return uc.handleDuplicateRequest(ctx, idempotencyKey)
		}
	}

	// Validate templateID is UUID
	templateId, errParseUUID := uuid.Parse(reportInput.TemplateID)
	if errParseUUID != nil {
		errInvalidID := pkg.ValidateBusinessError(constant.ErrInvalidTemplateID, "")

		return nil, errInvalidID
	}

	// Find a template to generate a report
	tOutputFormat, tMappedFields, err := uc.TemplateRepo.FindMappedFieldsAndOutputFormatByID(ctx, templateId)
	if err != nil {
		logger.Errorf("Error to find template by id, Error: %v", err)

		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, pkg.ValidateBusinessError(constant.ErrEntityNotFound, "", constant.MongoCollectionTemplate)
		}

		libOpentelemetry.HandleSpanError(&span, "Failed to find template by ID", err)

		return nil, err
	}

	if reportInput.Filters != nil {
		filtersMapped := uc.convertFiltersToMappedFieldsType(reportInput.Filters)
		if errValidateFields := uc.ValidateIfFieldsExistOnTables(ctx, filtersMapped); errValidateFields != nil {
			libOpentelemetry.HandleSpanError(&span, "Failed to validate filter fields existence on tables", errValidateFields)

			return nil, errValidateFields
		}
	}

	// Build the report model using constructor with invariant validation
	reportModel, err := report.NewReport(
		commons.GenerateUUIDv7(),
		templateId,
		constant.ProcessingStatus,
		reportInput.Filters,
	)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to create report entity", err)

		return nil, err
	}

	result, err := uc.ReportRepo.Create(ctx, reportModel)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to create report in repository", err)

		logger.Errorf("Error creating report in database: %v", err)

		return nil, err
	}

	// Build report message model
	reportMessage := model.ReportMessage{
		TemplateID:   templateId,
		ReportID:     result.ID,
		Filters:      reportInput.Filters,
		OutputFormat: *tOutputFormat,
		MappedFields: tMappedFields,
	}

	logger.Infof("Sending report to reports queue...")

	if err := uc.SendReportQueueReports(ctx, reportMessage); err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to send report to queue", err)

		logger.Errorf("Error sending report to queue: %v", err)

		// Update report status to error since queue send failed
		metadata := map[string]any{
			"error": "Failed to send report to queue",
		}
		if updateErr := uc.ReportRepo.UpdateReportStatusById(ctx, constant.ErrorStatus, result.ID, time.Now(), metadata); updateErr != nil {
			libOpentelemetry.HandleSpanError(&span, "Failed to update report status to error", updateErr)
			logger.Errorf("Error updating report status to error: %v", updateErr)
		}

		return nil, err
	}

	// Cache the successful result for idempotency deduplication of future identical requests
	if uc.RedisRepo != nil {
		idempotencyKey, keyErr := uc.buildIdempotencyKey(ctx, reportInput)
		if keyErr == nil {
			uc.cacheIdempotencyResult(ctx, idempotencyKey, result)
		}
	}

	return result, nil
}

// buildIdempotencyKey resolves the idempotency key for the request.
// If a client-provided Idempotency-Key header value exists in context, it is used as-is.
// Otherwise, a SHA256 hash of the JSON-serialized request body is computed.
func (uc *UseCase) buildIdempotencyKey(ctx context.Context, reportInput *model.CreateReportInput) (string, error) {
	logger, tracer, _, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.report.build_idempotency_key")
	defer span.End()

	// Check for client-provided idempotency key from context
	if clientKey, ok := ctx.Value(constant.IdempotencyKeyCtx).(string); ok && clientKey != "" {
		key := constant.IdempotencyKeyPrefix + ":" + clientKey

		logger.Infof("Using client-provided idempotency key: %s", key)

		return key, nil
	}

	// Compute SHA256 hash of the serialized request body
	data, err := json.Marshal(reportInput)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to marshal report input for idempotency hash", err)

		return "", fmt.Errorf("failed to marshal report input for idempotency hash: %w", err)
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	key := constant.IdempotencyKeyPrefix + ":" + hash

	logger.Infof("Computed idempotency key from request body hash: %s", key)

	return key, nil
}

// handleDuplicateRequest handles the case where SetNX returned false (key already exists).
// It attempts to retrieve the cached response from Redis. If a cached response exists,
// it is unmarshaled and returned. If no cached response exists yet (in-flight request),
// an error is returned indicating a duplicate in-flight request.
func (uc *UseCase) handleDuplicateRequest(ctx context.Context, idempotencyKey string) (*report.Report, error) {
	logger, tracer, _, _ := commons.NewTrackingFromContext(ctx)

	ctx, childSpan := tracer.Start(ctx, "service.report.handle_duplicate_request")
	defer childSpan.End()

	logger.Infof("Duplicate request detected for idempotency key: %s", idempotencyKey)

	cachedResponse, getErr := uc.RedisRepo.Get(ctx, idempotencyKey)
	if getErr != nil {
		libOpentelemetry.HandleSpanError(&childSpan, "Failed to retrieve cached idempotency response", getErr)

		return nil, getErr
	}

	// If the cached value is empty or still "processing", the first request is still in-flight
	if cachedResponse == "" || cachedResponse == "processing" {
		libOpentelemetry.HandleSpanBusinessErrorEvent(&childSpan, "Duplicate in-flight request detected", constant.ErrDuplicateRequestInFlight)

		return nil, pkg.ValidateBusinessError(constant.ErrDuplicateRequestInFlight, "report")
	}

	// Unmarshal the cached response
	var cachedReport report.Report
	if unmarshalErr := json.Unmarshal([]byte(cachedResponse), &cachedReport); unmarshalErr != nil {
		libOpentelemetry.HandleSpanError(&childSpan, "Failed to unmarshal cached idempotency response", unmarshalErr)

		return nil, fmt.Errorf("failed to unmarshal cached idempotency response: %w", unmarshalErr)
	}

	logger.Infof("Returning cached idempotent response for key: %s", idempotencyKey)

	return &cachedReport, nil
}

// cacheIdempotencyResult caches the successful report creation result in Redis
// so that future duplicate requests can return the cached response.
func (uc *UseCase) cacheIdempotencyResult(ctx context.Context, idempotencyKey string, result *report.Report) {
	logger, tracer, _, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.report.cache_idempotency_result")
	defer span.End()

	data, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to marshal report result for idempotency cache", marshalErr)
		logger.Errorf("Failed to marshal report result for idempotency cache: %v", marshalErr)

		return
	}

	if setErr := uc.RedisRepo.Set(ctx, idempotencyKey, string(data), constant.IdempotencyTTL); setErr != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to cache idempotency result", setErr)
		logger.Errorf("Failed to cache idempotency result: %v", setErr)
	}
}

// convertFiltersToMappedFieldsType transforms a deeply nested filter map into a mapped fields structure with limited keys per level.
func (uc *UseCase) convertFiltersToMappedFieldsType(filters map[string]map[string]map[string]model.FilterCondition) map[string]map[string][]string {
	output := make(map[string]map[string][]string)

	for topKey, nested := range filters {
		output[topKey] = make(map[string][]string)

		for midKey, inner := range nested {
			var keys []string

			count := 0

			for innerKey := range inner {
				keys = append(keys, innerKey)

				count++
				if count == constant.MaxSchemaPreviewKeys {
					break
				}
			}

			output[topKey][midKey] = keys
		}
	}

	return output
}
