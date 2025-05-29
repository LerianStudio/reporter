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
	ErrInvalidMapFields                = errors.New("TPL-0008")
	ErrInvalidPathParameter            = errors.New("TPL-0009")
	ErrOutputFormatWithoutTemplateFile = errors.New("TPL-0010")
	ErrEntityNotFound                  = errors.New("TPL-0011")
	ErrInvalidTemplateID               = errors.New("TPL-0012")
	ErrInvalidLedgerIDList             = errors.New("TPL-0013")
	ErrMissingTableFields              = errors.New("TPL-0014")
	ErrUnexpectedFieldsInTheRequest    = errors.New("TPL-0015")
	ErrMissingFieldsInRequest          = errors.New("TPL-0016")
	ErrBadRequest                      = errors.New("TPL-0017")
	ErrInternalServer                  = errors.New("TPL-0018")
	ErrInvalidQueryParameter           = errors.New("TPL-0019")
	ErrInvalidDateFormat               = errors.New("TPL-0020")
	ErrInvalidFinalDate                = errors.New("TPL-0021")
	ErrDateRangeExceedsLimit           = errors.New("TPL-0022")
	ErrInvalidDateRange                = errors.New("TPL-0023")
	ErrPaginationLimitExceeded         = errors.New("TPL-0024")
	ErrInvalidSortOrder                = errors.New("TPL-0025")
	ErrMetadataKeyLengthExceeded       = errors.New("TPL-0026")
	ErrMetadataValueLengthExceeded     = errors.New("TPL-0027")
	ErrInvalidMetadataNesting          = errors.New("TPL-0028")
	ErrReportStatusNotFinished         = errors.New("TPL-0029")
	ErrMissingSchemaTable              = errors.New("TPL-0030")
	ErrMissingDataSource               = errors.New("TPL-0031")
)
