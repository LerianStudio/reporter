// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/mongodb/template"
	"github.com/LerianStudio/reporter/pkg/net/http"
	templateUtils "github.com/LerianStudio/reporter/pkg/template_utils"

	"github.com/LerianStudio/lib-commons/v2/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"go.opentelemetry.io/otel/attribute"
)

// CreateTemplate creates a new template with specified parameters, stores it in the repository,
// uploads the file to object storage, and performs a compensating transaction on storage failure.
func (uc *UseCase) CreateTemplate(ctx context.Context, templateFile, outFormat, description string, fileHeader *multipart.FileHeader) (*template.Template, error) {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.template.create")
	defer span.End()

	span.SetAttributes(
		attribute.String("app.request.request_id", reqId),
		attribute.String("app.request.template_file", templateFile),
		attribute.String("app.request.output_format", outFormat),
		attribute.String("app.request.description", description),
	)

	logger.Infof("Creating template")

	// Idempotency check: acquire lock via Redis SetNX before proceeding
	if uc.RedisRepo != nil {
		idempotencyKey, keyErr := uc.buildTemplateIdempotencyKey(ctx, templateFile, outFormat, description)
		if keyErr != nil {
			libOpentelemetry.HandleSpanError(&span, "Failed to compute template idempotency key", keyErr)

			return nil, keyErr
		}

		span.SetAttributes(attribute.String("app.idempotency.key", idempotencyKey))

		logger.Infof("Checking idempotency for key: %s", idempotencyKey)

		acquired, setNXErr := uc.RedisRepo.SetNX(ctx, idempotencyKey, "processing", constant.IdempotencyTTL)
		if setNXErr != nil {
			libOpentelemetry.HandleSpanError(&span, "Failed to acquire template idempotency lock", setNXErr)

			return nil, setNXErr
		}

		if !acquired {
			return uc.handleDuplicateTemplateRequest(ctx, idempotencyKey)
		}
	}

	// Block <script> tags
	if err := templateUtils.ValidateNoScriptTag(templateFile); err != nil {
		return nil, pkg.ValidateBusinessError(constant.ErrScriptTagDetected, "")
	}

	mappedFields := templateUtils.MappedFieldsOfTemplate(templateFile)
	logger.Infof("Mapped Fields is valid to continue %v", mappedFields)

	if errValidateFields := uc.ValidateIfFieldsExistOnTables(ctx, mappedFields); errValidateFields != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to validate fields existence on tables", errValidateFields)

		logger.Errorf("Error to validate fields existence on tables, Error: %v", errValidateFields)

		return nil, errValidateFields
	}

	// Transform mapped fields for storage
	// Get MidazOrganizationID from plugin_crm datasource if template uses it
	var midazOrgID string

	if _, hasPluginCRM := mappedFields["plugin_crm"]; hasPluginCRM {
		if ds, exists := uc.ExternalDataSources.Get("plugin_crm"); exists {
			midazOrgID = ds.MidazOrganizationID
		}
	}

	transformedMappedFields := TransformMappedFieldsForStorage(mappedFields, midazOrgID)
	logger.Infof("Transformed Mapped Fields for storage %v", transformedMappedFields)

	templateId := commons.GenerateUUIDv7()
	fileName := fmt.Sprintf("%s.tpl", templateId.String())

	templateModel := &template.TemplateMongoDBModel{
		ID:           templateId,
		OutputFormat: strings.ToLower(outFormat),
		FileName:     fileName,
		Description:  description,
		MappedFields: transformedMappedFields,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		DeletedAt:    nil,
	}

	resultTemplateModel, err := uc.TemplateRepo.Create(ctx, templateModel)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to create template in repository", err)

		logger.Errorf("Error into creating a template, Error: %v", err)

		return nil, err
	}

	// Read file bytes and upload to object storage
	fileBytes, err := http.ReadMultipartFile(fileHeader)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to read multipart file", err)

		logger.Errorf("Error to get the file content: %v", err)

		return nil, err
	}

	errPutStorage := uc.TemplateSeaweedFS.Put(ctx, resultTemplateModel.FileName, outFormat, fileBytes)
	if errPutStorage != nil {
		libOpentelemetry.HandleSpanError(&span, "Error putting template file on storage", errPutStorage)

		// Compensating transaction: Attempt to roll back the database change to prevent an orphaned record.
		if errDelete := uc.DeleteTemplateByID(ctx, resultTemplateModel.ID, true); errDelete != nil {
			logger.Errorf("Failed to roll back template creation for ID %s after storage failure. Error: %s", resultTemplateModel.ID.String(), errDelete.Error())
		}

		logger.Errorf("Error putting template file on storage: %s", errPutStorage.Error())

		return nil, errPutStorage
	}

	// Cache the successful result for idempotency deduplication of future identical requests
	if uc.RedisRepo != nil {
		idempotencyKey, keyErr := uc.buildTemplateIdempotencyKey(ctx, templateFile, outFormat, description)
		if keyErr == nil {
			uc.cacheTemplateIdempotencyResult(ctx, idempotencyKey, resultTemplateModel)
		}
	}

	return resultTemplateModel, nil
}

// templateIdempotencyInput is the internal struct used to compute idempotency hashes
// for template creation requests. It captures the unique combination of template content,
// output format, and description that defines a distinct template.
type templateIdempotencyInput struct {
	TemplateFile string `json:"templateFile"`
	OutputFormat string `json:"outputFormat"`
	Description  string `json:"description"`
}

// buildTemplateIdempotencyKey resolves the idempotency key for the template creation request.
// If a client-provided Idempotency-Key header value exists in context, it is used as-is.
// Otherwise, a SHA256 hash of the JSON-serialized request fields is computed.
func (uc *UseCase) buildTemplateIdempotencyKey(ctx context.Context, templateFile, outFormat, description string) (string, error) {
	logger, tracer, _, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.template.build_idempotency_key")
	defer span.End()

	// Check for client-provided idempotency key from context
	if clientKey, ok := ctx.Value(constant.IdempotencyKeyCtx).(string); ok && clientKey != "" {
		key := constant.IdempotencyKeyPrefix + ":template:" + clientKey

		logger.Infof("Using client-provided template idempotency key: %s", key)

		return key, nil
	}

	// Compute SHA256 hash of the serialized request fields
	input := templateIdempotencyInput{
		TemplateFile: templateFile,
		OutputFormat: outFormat,
		Description:  description,
	}

	data, err := json.Marshal(input)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to marshal template input for idempotency hash", err)

		return "", fmt.Errorf("failed to marshal template input for idempotency hash: %w", err)
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	key := constant.IdempotencyKeyPrefix + ":template:" + hash

	logger.Infof("Computed template idempotency key from request body hash: %s", key)

	return key, nil
}

// handleDuplicateTemplateRequest handles the case where SetNX returned false (key already exists).
// It attempts to retrieve the cached response from Redis. If a cached response exists,
// it is unmarshaled and returned. If no cached response exists yet (in-flight request),
// an error is returned indicating a duplicate in-flight request.
func (uc *UseCase) handleDuplicateTemplateRequest(ctx context.Context, idempotencyKey string) (*template.Template, error) {
	logger, tracer, _, _ := commons.NewTrackingFromContext(ctx)

	ctx, childSpan := tracer.Start(ctx, "service.template.handle_duplicate_request")
	defer childSpan.End()

	logger.Infof("Duplicate template request detected for idempotency key: %s", idempotencyKey)

	cachedResponse, getErr := uc.RedisRepo.Get(ctx, idempotencyKey)
	if getErr != nil {
		libOpentelemetry.HandleSpanError(&childSpan, "Failed to retrieve cached template idempotency response", getErr)

		return nil, getErr
	}

	// If the cached value is empty or still "processing", the first request is still in-flight
	if cachedResponse == "" || cachedResponse == "processing" {
		libOpentelemetry.HandleSpanBusinessErrorEvent(&childSpan, "Duplicate in-flight template request detected", constant.ErrDuplicateRequestInFlight)

		return nil, pkg.ValidateBusinessError(constant.ErrDuplicateRequestInFlight, "template")
	}

	// Unmarshal the cached response
	var cachedTemplate template.Template
	if unmarshalErr := json.Unmarshal([]byte(cachedResponse), &cachedTemplate); unmarshalErr != nil {
		libOpentelemetry.HandleSpanError(&childSpan, "Failed to unmarshal cached template idempotency response", unmarshalErr)

		return nil, fmt.Errorf("failed to unmarshal cached template idempotency response: %w", unmarshalErr)
	}

	logger.Infof("Returning cached idempotent template response for key: %s", idempotencyKey)

	return &cachedTemplate, nil
}

// cacheTemplateIdempotencyResult caches the successful template creation result in Redis
// so that future duplicate requests can return the cached response.
func (uc *UseCase) cacheTemplateIdempotencyResult(ctx context.Context, idempotencyKey string, result *template.Template) {
	logger, tracer, _, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "service.template.cache_idempotency_result")
	defer span.End()

	data, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to marshal template result for idempotency cache", marshalErr)
		logger.Errorf("Failed to marshal template result for idempotency cache: %v", marshalErr)

		return
	}

	if setErr := uc.RedisRepo.Set(ctx, idempotencyKey, string(data), constant.IdempotencyTTL); setErr != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to cache template idempotency result", setErr)
		logger.Errorf("Failed to cache template idempotency result: %v", setErr)
	}
}
