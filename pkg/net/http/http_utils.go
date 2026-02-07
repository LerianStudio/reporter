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

// ValidateParameters validate and return struct of default parameters
func ValidateParameters(params map[string]string) (*QueryHeader, error) {
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
		case strings.Contains(key, "outputFormat"):
			if !pkg.IsOutputFormatValuesValid(&value) {
				return nil, pkg.ValidateBusinessError(constant.ErrInvalidOutputFormat, "")
			}

			outputFormat = value
		case strings.Contains(key, "description"):
			description = value
		case strings.Contains(key, "status"):
			status = value
		case strings.Contains(key, "templateId"):
			if parsedID, err := uuid.Parse(value); err == nil {
				templateID = parsedID
			}
		case strings.Contains(key, "limit"):
			limit, _ = strconv.Atoi(value)
		case strings.Contains(key, "page"):
			page, _ = strconv.Atoi(value)
		case strings.Contains(key, "cursor"):
			cursor = value
		case strings.Contains(key, "sortOrder"):
			sortOrder = strings.ToLower(value)
		case strings.Contains(key, "createdAt"):
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
