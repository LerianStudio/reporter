// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package http

import (
	"bytes"
	"io"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/constant"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

// QueryHeader entity from query parameter from get apis
type QueryHeader struct {
	Metadata     *bson.M
	OutputFormat string
	Description  string
	Status       string
	TemplateID   uuid.UUID
	Limit        int
	Page         int
	Cursor       string
	SortOrder    string
	CreatedAt    time.Time
	Alias        string
	UseMetadata  bool
	ToAssetCodes []string
}

// Pagination entity from query parameter from get apis
type Pagination struct {
	Limit     int
	Page      int
	Cursor    string
	SortOrder string
	Alias     string
}

func (qh *QueryHeader) ToOffsetPagination() Pagination {
	return Pagination{
		Limit:     qh.Limit,
		Page:      qh.Page,
		SortOrder: qh.SortOrder,
		Alias:     qh.Alias,
	}
}

// normalizeParams rewrites legacy camelCase query parameter keys to their
// snake_case equivalents so the parsing loop only needs to match one format.
// When both formats are present for the same parameter, snake_case takes precedence.
func normalizeParams(params map[string]string) map[string]string {
	aliases := map[string]string{
		"outputFormat": "output_format",
		"sortOrder":    "sort_order",
		"templateId":   "template_id",
		"createdAt":    "created_at",
	}

	normalized := make(map[string]string, len(params))

	for k, v := range params {
		normalized[k] = v
	}

	for camel, snake := range aliases {
		if _, hasSnake := normalized[snake]; hasSnake {
			// snake_case already present; remove legacy camelCase if it exists
			delete(normalized, camel)
			continue
		}

		if val, hasCamel := normalized[camel]; hasCamel {
			normalized[snake] = val
			delete(normalized, camel)
		}
	}

	return normalized
}

// parsePositiveInt parses a string as an integer and validates that the result
// is at least 1. It returns a validation error referencing paramName on failure.
func parsePositiveInt(value, paramName string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, pkg.ValidateBusinessError(constant.ErrInvalidQueryParameter, "", paramName)
	}

	if parsed < 1 {
		return 0, pkg.ValidateBusinessError(constant.ErrInvalidQueryParameter, "", paramName)
	}

	return parsed, nil
}

// ValidateParameters validate and return struct of default parameters.
// It accepts both snake_case (preferred) and camelCase (deprecated) query parameter names.
func ValidateParameters(params map[string]string) (*QueryHeader, error) {
	params = normalizeParams(params)

	var (
		metadata     *bson.M
		createdAt    time.Time
		cursor       string
		outputFormat string
		description  string
		status       string
		templateID   uuid.UUID
		limit        = constant.DefaultPaginationLimit
		page         = constant.DefaultPaginationPage
		sortOrder    = "desc"
		useMetadata  = false
	)

	for key, value := range params {
		switch {
		case strings.Contains(key, "metadata."):
			metadata = &bson.M{key: value}
			useMetadata = true
		case key == "output_format":
			if !pkg.IsOutputFormatValuesValid(&value) {
				return nil, pkg.ValidateBusinessError(constant.ErrInvalidOutputFormat, "")
			}

			outputFormat = value
		case key == "description":
			description = value
		case key == "status":
			status = value
		case key == "template_id":
			if parsedID, err := uuid.Parse(value); err == nil {
				templateID = parsedID
			}
		case key == "limit":
			parsed, err := parsePositiveInt(value, "limit")
			if err != nil {
				return nil, err
			}

			limit = parsed
		case key == "page":
			parsed, err := parsePositiveInt(value, "page")
			if err != nil {
				return nil, err
			}

			page = parsed
		case key == "cursor":
			cursor = value
		case key == "sort_order":
			sortOrder = strings.ToLower(value)
		case key == "created_at":
			createdAt, _ = time.Parse("2006-01-02", value)
		}
	}

	err := validatePagination(cursor, sortOrder, limit)
	if err != nil {
		return nil, err
	}

	query := &QueryHeader{
		Metadata:     metadata,
		OutputFormat: outputFormat,
		Description:  description,
		Status:       status,
		TemplateID:   templateID,
		Limit:        limit,
		Page:         page,
		Cursor:       cursor,
		SortOrder:    sortOrder,
		CreatedAt:    createdAt,
		UseMetadata:  useMetadata,
	}

	return query, nil
}

// GetFileFromHeader method that get file from header and give a string
func GetFileFromHeader(fileHeader *multipart.FileHeader) (string, error) {
	if !strings.Contains(fileHeader.Filename, fileExtension) {
		return "", pkg.ValidateBusinessError(constant.ErrInvalidFileFormat, "")
	}

	if fileHeader.Size == 0 {
		return "", pkg.ValidateBusinessError(constant.ErrEmptyFile, "")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}

	defer func(file multipart.File) {
		_ = file.Close()
	}(file)

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, file); err != nil {
		return "", pkg.ValidateBusinessError(constant.ErrInvalidFileUploaded, "", err)
	}

	fileString := buf.String()

	return fileString, nil
}

func ReadMultipartFile(fileHeader *multipart.FileHeader) ([]byte, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

func validatePagination(cursor, sortOrder string, limit int) error {
	maxPaginationLimit := pkg.SafeInt64ToInt(pkg.GetenvIntOrDefault("MAX_PAGINATION_LIMIT", constant.DefaultMaxPaginationLimit))

	if limit > maxPaginationLimit {
		return pkg.ValidateBusinessError(constant.ErrPaginationLimitExceeded, "", maxPaginationLimit)
	}

	if (sortOrder != string(constant.Asc)) && (sortOrder != string(constant.Desc)) {
		return pkg.ValidateBusinessError(constant.ErrInvalidSortOrder, "")
	}

	if !pkg.IsNilOrEmpty(&cursor) {
		_, err := DecodeCursor(cursor)
		if err != nil {
			return pkg.ValidateBusinessError(constant.ErrInvalidQueryParameter, "", "cursor")
		}
	}

	return nil
}
