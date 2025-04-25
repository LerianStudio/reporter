package constant

import (
	"errors"
)

// List of errors that can be returned.
// You can standardize errors
var (
	ErrParentExampleIDNotFound = errors.New("0014") // Standardized error

	ErrUnexpectedFieldsInTheRequest = errors.New("0001")
	ErrMissingFieldsInRequest       = errors.New("0002")
	ErrBadRequest                   = errors.New("0003")
	ErrInternalServer               = errors.New("0004")
	ErrCalculationFieldType         = errors.New("0005")
	ErrInvalidQueryParameter        = errors.New("0006")
	ErrInvalidDateFormat            = errors.New("0007")
	ErrInvalidFinalDate             = errors.New("0008")
	ErrDateRangeExceedsLimit        = errors.New("0009")
	ErrInvalidDateRange             = errors.New("0010")
	ErrPaginationLimitExceeded      = errors.New("0011")
	ErrInvalidSortOrder             = errors.New("0011")
	ErrActionNotPermitted           = errors.New("0013")
	ErrMetadataKeyLengthExceeded    = errors.New("0014")
	ErrMetadataValueLengthExceeded  = errors.New("0015")
	ErrInvalidMetadataNesting       = errors.New("0016")

	ErrMissingRequiredFields           = errors.New("TPL-0001")
	ErrInvalidFileFormat               = errors.New("TPL-0002")
	ErrInvalidOutputFormat             = errors.New("TPL-0003")
	ErrInvalidHeaderParameter          = errors.New("TPL-0004")
	ErrInvalidFileUploaded             = errors.New("TPL-0005")
	ErrEmptyFile                       = errors.New("TPL-0006")
	ErrFileContentInvalid              = errors.New("TPL-0007")
	ErrKeyNotAllowed                   = errors.New("TPL-0008")
	ErrInvalidMapFields                = errors.New("TPL-0009")
	ErrMapFieldKeyUnexpected           = errors.New("TPL-0010")
	ErrInvalidPathParameter            = errors.New("TPL-0011")
	ErrOutputFormatWithoutTemplateFile = errors.New("TPL-0012")
	ErrEntityNotFound                  = errors.New("TPL-0013")
)
