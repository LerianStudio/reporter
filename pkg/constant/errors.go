package constant

import (
	"errors"
)

// List of errors that can be returned.
// You can standardize errors
// Standardized error
var (
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
	ErrInvalidTemplateID               = errors.New("TPL-0014")
	ErrInvalidLedgerIDList             = errors.New("TPL-0015")
	ErrMissingTableFields              = errors.New("TPL-0016")
	ErrUnexpectedFieldsInTheRequest    = errors.New("TPL-0017")
	ErrMissingFieldsInRequest          = errors.New("TPL-0018")
	ErrBadRequest                      = errors.New("TPL-0019")
	ErrInternalServer                  = errors.New("TPL-0020")
	ErrInvalidQueryParameter           = errors.New("TPL-0021")
	ErrInvalidDateFormat               = errors.New("TPL-0022")
	ErrInvalidFinalDate                = errors.New("TPL-0023")
	ErrDateRangeExceedsLimit           = errors.New("TPL-0024")
	ErrInvalidDateRange                = errors.New("TPL-0025")
	ErrPaginationLimitExceeded         = errors.New("TPL-0026")
	ErrInvalidSortOrder                = errors.New("TPL-0027")
	ErrMetadataKeyLengthExceeded       = errors.New("TPL-0028")
	ErrMetadataValueLengthExceeded     = errors.New("TPL-0029")
	ErrInvalidMetadataNesting          = errors.New("TPL-0030")
	ErrReportStatusNotFinished         = errors.New("TPL-0031")
	ErrMissingSchemaTable              = errors.New("TPL-0032")
)
