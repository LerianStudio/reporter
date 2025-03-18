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
	ErrEntityNotFound               = errors.New("0012")
	ErrActionNotPermitted           = errors.New("0013")
	ErrMetadataKeyLengthExceeded    = errors.New("0014")
	ErrMetadataValueLengthExceeded  = errors.New("0015")
	ErrInvalidMetadataNesting       = errors.New("0016")
	ErrInvalidFileFormat            = errors.New("TPE-0001")
	ErrEmptyFile                    = errors.New("TPE-0002")
)
