package pkg

import (
	"errors"
	"fmt"
	"plugin-smart-templates/pkg/constant"
	"strings"
)

// EntityNotFoundError records an error indicating an entity was not found in any case that caused it.
// You can use it to representing a Database not found, cache not found or any other repository.
type EntityNotFoundError struct {
	EntityType string
	Title      string
	Message    string
	Code       string
	Err        error
}

// Error implements the error interface.
func (e EntityNotFoundError) Error() string {
	if strings.TrimSpace(e.Message) == "" {
		if strings.TrimSpace(e.EntityType) != "" {
			return fmt.Sprintf("Entity %s not found", e.EntityType)
		}

		if e.Err != nil && strings.TrimSpace(e.Message) == "" {
			return e.Err.Error()
		}

		return "entity not found"
	}

	return e.Message
}

// Unwrap implements the error interface introduced in Go 1.13 to unwrap the internal error.
func (e EntityNotFoundError) Unwrap() error {
	return e.Err
}

// ValidationError records an error indicating an entity was not found in any case that caused it.
// You can use it to representing a Database not found, cache not found or any other repository.
type ValidationError struct {
	EntityType string `json:"entityType,omitempty"`
	Title      string
	Message    string
	Code       string
	Err        error `json:"err,omitempty"`
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	if strings.TrimSpace(e.Code) != "" {
		return fmt.Sprintf("%s - %s", e.Code, e.Message)
	}

	return e.Message
}

// Unwrap implements the error interface introduced in Go 1.13 to unwrap the internal error.
func (e ValidationError) Unwrap() error {
	return e.Err
}

// EntityConflictError records an error indicating an entity already exists in some repository
// You can use it to representing a Database conflict, cache or any other repository.
type EntityConflictError struct {
	EntityType string
	Title      string
	Message    string
	Code       string
	Err        error
}

// ValidationKnownFieldsError records an error that occurred during a validation of known fields.
type ValidationKnownFieldsError struct {
	EntityType string           `json:"entityType,omitempty"`
	Title      string           `json:"title,omitempty"`
	Code       string           `json:"code,omitempty"`
	Message    string           `json:"message,omitempty"`
	Fields     FieldValidations `json:"fields,omitempty"`
}

// Error implements the error interface.
func (e EntityConflictError) Error() string {
	if e.Err != nil && strings.TrimSpace(e.Message) == "" {
		return e.Err.Error()
	}

	return e.Message
}

// Unwrap implements the error interface introduced in Go 1.13 to unwrap the internal error.
func (e EntityConflictError) Unwrap() error {
	return e.Err
}

// UnauthorizedError indicates an operation that couldn't be performant because there's no user authenticated.
type UnauthorizedError struct {
	EntityType string `json:"entityType,omitempty"`
	Title      string `json:"title,omitempty"`
	Message    string `json:"message,omitempty"`
	Code       string `json:"code,omitempty"`
	Err        error  `json:"err,omitempty"`
}

func (e UnauthorizedError) Error() string {
	return e.Message
}

// ForbiddenError indicates an operation that couldn't be performant because the authenticated user has no sufficient privileges.
type ForbiddenError struct {
	EntityType string `json:"entityType,omitempty"`
	Title      string `json:"title,omitempty"`
	Message    string `json:"message,omitempty"`
	Code       string `json:"code,omitempty"`
	Err        error  `json:"err,omitempty"`
}

func (e ForbiddenError) Error() string {
	return e.Message
}

// UnprocessableOperationError indicates an operation that couldn't be performant because it's invalid.
type UnprocessableOperationError struct {
	EntityType string
	Title      string
	Message    string
	Code       string
	Err        error
}

func (e UnprocessableOperationError) Error() string {
	return e.Message
}

// HTTPError indicates a http error raised in a http client.
type HTTPError struct {
	EntityType string
	Title      string
	Message    string
	Code       string
	Err        error
}

func (e HTTPError) Error() string {
	return e.Message
}

// FailedPreconditionError indicates a precondition failed during an operation.
type FailedPreconditionError struct {
	EntityType string `json:"entityType,omitempty"`
	Title      string `json:"title,omitempty"`
	Message    string `json:"message,omitempty"`
	Code       string `json:"code,omitempty"`
	Err        error  `json:"err,omitempty"`
}

func (e FailedPreconditionError) Error() string {
	return e.Message
}

// InternalServerError indicates a precondition failed during an operation.
type InternalServerError struct {
	EntityType string `json:"entityType,omitempty"`
	Title      string `json:"title,omitempty"`
	Message    string `json:"message,omitempty"`
	Code       string `json:"code,omitempty"`
	Err        error  `json:"err,omitempty"`
}

func (e InternalServerError) Error() string {
	return e.Message
}

// ResponseError is a struct used to return errors to the client.
type ResponseError struct {
	Code    int    `json:"code,omitempty"`
	Title   string `json:"title,omitempty"`
	Message string `json:"message,omitempty"`
}

// Error returns the message of the ResponseError.
//
// No parameters.
// Returns a string.
func (r ResponseError) Error() string {
	return r.Message
}

// Error returns the error message for a ValidationKnownFieldsError.
//
// No parameters.
// Returns a string.
func (r ValidationKnownFieldsError) Error() string {
	return r.Message
}

// FieldValidations is a map of known fields and their validation errors.
type FieldValidations map[string]string

// ValidationUnknownFieldsError records an error that occurred during a validation of known fields.
type ValidationUnknownFieldsError struct {
	EntityType string        `json:"entityType,omitempty"`
	Title      string        `json:"title,omitempty"`
	Code       string        `json:"code,omitempty"`
	Message    string        `json:"message,omitempty"`
	Fields     UnknownFields `json:"fields,omitempty"`
}

// Error returns the error message for a ValidationUnknownFieldsError.
//
// No parameters.
// Returns a string.
func (r ValidationUnknownFieldsError) Error() string {
	return r.Message
}

// UnknownFields is a map of unknown fields and their error messages.
type UnknownFields map[string]any

// Methods to create errors for different scenarios:

// ValidateInternalError validates the error and returns an appropriate InternalServerError.
//
// Parameters:
// - err: The error to be validated.
// - entityType: The type of the entity associated with the error.
//
// Returns:
// - An InternalServerError with the appropriate code, title, message.
func ValidateInternalError(err error, entityType string) error {
	return InternalServerError{
		EntityType: entityType,
		Code:       constant.ErrInternalServer.Error(),
		Title:      "Internal Server Error",
		Message:    "The server encountered an unexpected error. Please try again later or contact support.",
		Err:        err,
	}
}

// ValidateBadRequestFieldsError validates the error and returns the appropriate bad request error code, title, message, and the invalid fields.
//
// Parameters:
// - requiredFields: A map of missing required fields and their error messages.
// - knownInvalidFields: A map of known invalid fields and their validation errors.
// - entityType: The type of the entity associated with the error.
// - unknownFields: A map of unknown fields and their error messages.
//
// Returns:
// - An error indicating the validation result, which could be a ValidationUnknownFieldsError or a ValidationKnownFieldsError.
func ValidateBadRequestFieldsError(requiredFields, knownInvalidFields map[string]string, entityType string, unknownFields map[string]any) error {
	if len(unknownFields) == 0 && len(knownInvalidFields) == 0 && len(requiredFields) == 0 {
		return errors.New("expected knownInvalidFields, unknownFields and requiredFields to be non-empty")
	}

	if len(unknownFields) > 0 {
		return ValidationUnknownFieldsError{
			EntityType: entityType,
			Code:       constant.ErrUnexpectedFieldsInTheRequest.Error(),
			Title:      "Unexpected Fields in the Request",
			Message:    "The request body contains more fields than expected. Please send only the allowed fields as per the documentation. The unexpected fields are listed in the fields object.",
			Fields:     unknownFields,
		}
	}

	if len(requiredFields) > 0 {
		return ValidationKnownFieldsError{
			EntityType: entityType,
			Code:       constant.ErrMissingFieldsInRequest.Error(),
			Title:      "Missing Fields in Request",
			Message:    "Your request is missing one or more required fields. Please refer to the documentation to ensure all necessary fields are included in your request.",
			Fields:     requiredFields,
		}
	}

	return ValidationKnownFieldsError{
		EntityType: entityType,
		Code:       constant.ErrBadRequest.Error(),
		Title:      "Bad Request",
		Message:    "The server could not understand the request due to malformed syntax. Please check the listed fields and try again.",
		Fields:     knownInvalidFields,
	}
}

// ValidateBusinessError validates the error and returns the appropriate business error code, title, and message.
// error: The appropriate business error with code, title, and message.
func ValidateBusinessError(err error, entityType string, args ...any) error {
	errorMap := map[error]error{
		constant.ErrInvalidQueryParameter: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrInvalidQueryParameter.Error(),
			Title:      "Invalid Query Parameter",
			Message:    fmt.Sprintf("One or more query parameters are in an incorrect format. Please check the following parameters '%v' and ensure they meet the required format before trying again.", args),
		},
		constant.ErrInvalidDateFormat: EntityNotFoundError{
			EntityType: entityType,
			Code:       constant.ErrInvalidDateFormat.Error(),
			Title:      "Invalid Date Format Error",
			Message:    "The 'initialDate', 'finalDate', or both are in the incorrect format. Please use the 'yyyy-mm-dd' format and try again.",
		},
		constant.ErrInvalidFinalDate: EntityNotFoundError{
			EntityType: entityType,
			Code:       constant.ErrInvalidFinalDate.Error(),
			Title:      "Invalid Final Date Error",
			Message:    "The 'finalDate' cannot be earlier than the 'initialDate'. Please verify the dates and try again.",
		},
		constant.ErrDateRangeExceedsLimit: EntityNotFoundError{
			EntityType: entityType,
			Code:       constant.ErrDateRangeExceedsLimit.Error(),
			Title:      "Date Range Exceeds Limit Error",
			Message:    fmt.Sprintf("The range between 'initialDate' and 'finalDate' exceeds the permitted limit of %v months. Please adjust the dates and try again.", args...),
		},
		constant.ErrInvalidDateRange: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrInvalidDateRange.Error(),
			Title:      "Invalid Date Range Error",
			Message:    "Both 'initialDate' and 'finalDate' fields are required and must be in the 'yyyy-mm-dd' format. Please provide valid dates and try again.",
		},
		constant.ErrPaginationLimitExceeded: EntityNotFoundError{
			EntityType: entityType,
			Code:       constant.ErrPaginationLimitExceeded.Error(),
			Title:      "Pagination Limit Exceeded",
			Message:    fmt.Sprintf("The pagination limit exceeds the maximum allowed of %v items per page. Please verify the limit and try again.", args...),
		},
		constant.ErrInvalidSortOrder: EntityNotFoundError{
			EntityType: entityType,
			Code:       constant.ErrInvalidSortOrder.Error(),
			Title:      "Invalid Sort Order",
			Message:    "The 'sort_order' field must be 'asc' or 'desc'. Please provide a valid sort order and try again.",
		},
		constant.ErrEntityNotFound: EntityNotFoundError{
			EntityType: entityType,
			Code:       constant.ErrEntityNotFound.Error(),
			Title:      "Entity Not Found",
			Message:    fmt.Sprintf("No %v entity was found for the given ID. Please make sure to use the correct ID for the entity you are trying to manage.", args...),
		},
		constant.ErrMetadataKeyLengthExceeded: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrMetadataKeyLengthExceeded.Error(),
			Title:      "Metadata Key Length Exceeded",
			Message:    fmt.Sprintf("The metadata key %v exceeds the maximum allowed length of %v characters. Please use a shorter key.", args...),
		},
		constant.ErrMetadataValueLengthExceeded: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrMetadataValueLengthExceeded.Error(),
			Title:      "Metadata Value Length Exceeded",
			Message:    fmt.Sprintf("The metadata value %v exceeds the maximum allowed length of %v characters. Please use a shorter value.", args...),
		},
		constant.ErrInvalidMetadataNesting: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrInvalidMetadataNesting.Error(),
			Title:      "Invalid Metadata Nesting",
			Message:    fmt.Sprintf("The metadata object cannot contain nested values. Please ensure that the value %v is not nested and try again.", args...),
		},

		constant.ErrMissingRequiredFields: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrMissingRequiredFields.Error(),
			Title:      "Missing required fields",
			Message:    "One or more required fields are missing. Please ensure all required fields like 'description', 'template', and 'outputFormat' are included.",
		},
		constant.ErrInvalidFileFormat: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrInvalidFileFormat.Error(),
			Title:      "Invalid file format",
			Message:    "The uploaded file must be a .tpl file. Other formats are not supported.",
		},
		constant.ErrInvalidOutputFormat: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrInvalidOutputFormat.Error(),
			Title:      "Invalid output format",
			Message:    "The outputFormat field must be one of: html, csv, or xml. ",
		},
		constant.ErrInvalidHeaderParameter: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrInvalidHeaderParameter.Error(),
			Title:      "Invalid header",
			Message:    fmt.Sprintf("One or more header values are missing or incorrectly formatted. Please verify required headers %v.", args),
		},
		constant.ErrInvalidFileUploaded: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrInvalidFileUploaded.Error(),
			Title:      "Invalid File Uploaded",
			Message:    fmt.Sprintf("The file you submitted is invalid. Please check the uploaded file with error: %v", args),
		},
		constant.ErrEmptyFile: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrEmptyFile.Error(),
			Title:      "Error File Empty",
			Message:    "The file you submitted is empty. Please check the uploaded file.",
		},
		constant.ErrFileContentInvalid: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrFileContentInvalid.Error(),
			Title:      "Error File Content Invalid",
			Message:    fmt.Sprintf("The file content is invalid because is not %s. Please check the uploaded file.", args),
		},
		constant.ErrKeyNotAllowed: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrKeyNotAllowed.Error(),
			Title:      "Error Key Not Allowed",
			Message:    fmt.Sprintf("The key %s on mapped fields is not allowed. Please check the uploaded file.", args),
		},
		constant.ErrInvalidMapFields: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrInvalidMapFields.Error(),
			Title:      "Invalid Map Fields",
			Message:    fmt.Sprintf("The field on template file is invalid. Invalid field %s on %s.", args...),
		},
		constant.ErrMapFieldKeyUnexpected: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrMapFieldKeyUnexpected.Error(),
			Title:      "Error Key Unexpected",
			Message:    fmt.Sprintf("The key %s is unexpected. Please check the uploaded file.", args),
		},
		constant.ErrInvalidPathParameter: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrInvalidPathParameter.Error(),
			Title:      "Invalid Path Parameter",
			Message:    fmt.Sprintf("Path parameters is in an incorrect format. Please check the following parameter %v and ensure they meet the required format before trying again.", args),
		},
		constant.ErrOutputFormatWithoutTemplateFile: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrOutputFormatWithoutTemplateFile.Error(),
			Title:      "Update Output format without template File",
			Message:    "Can not update output format without passing template file. Please check information passed and try again.",
		},
		constant.ErrInvalidTemplateID: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrInvalidTemplateID.Error(),
			Title:      "Invalid templateID",
			Message:    "The specified templateID is not a valid UUID. Please check the value passed.",
		},
		constant.ErrInvalidLedgerIDList: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrInvalidLedgerIDList.Error(),
			Title:      "Invalid ledgerID",
			Message:    fmt.Sprintf("The specified ledgerID inside ledger ID list is not a valid UUID. Please check the value passed %v.", args),
		},
		constant.ErrMissingTableFields: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrMissingTableFields.Error(),
			Title:      "Missing required fields",
			Message:    fmt.Sprintf("The fields mapped on template file is missing on tables schema. Please check the fields passed %v.", args...),
		},
		constant.ErrReportStatusNotFinished: ValidationError{
			EntityType: entityType,
			Code:       constant.ErrReportStatusNotFinished.Error(),
			Title:      "Report status not Finished",
			Message:    "The Report is not ready to download. Report is processing yet.",
		},
	}

	if mappedError, found := errorMap[err]; found {
		return mappedError
	}

	return err
}
