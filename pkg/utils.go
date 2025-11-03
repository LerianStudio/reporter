package pkg

import (
	"math"
	"os/exec"
	"reflect"
	"regexp"
	"strings"

	"github.com/LerianStudio/reporter/v4/pkg/constant"
)

// GetMapNumKinds get the map of numeric kinds to use in validations and conversions.
//
// The numeric kinds are:
// - int
// - int8
// - int16
// - int32
// - int64
// - float32
// - float64
func GetMapNumKinds() map[reflect.Kind]bool {
	numKinds := make(map[reflect.Kind]bool)

	numKinds[reflect.Int] = true
	numKinds[reflect.Int8] = true
	numKinds[reflect.Int16] = true
	numKinds[reflect.Int32] = true
	numKinds[reflect.Int64] = true
	numKinds[reflect.Float32] = true
	numKinds[reflect.Float64] = true

	return numKinds
}

type SyscmdI interface {
	ExecCmd(name string, arg ...string) ([]byte, error)
}

type Syscmd struct{}

func (r *Syscmd) ExecCmd(name string, arg ...string) ([]byte, error) {
	return exec.Command(name, arg...).Output()
}

// IsNilOrEmpty returns a boolean indicating if a *string is nil or empty.
// It's use TrimSpace so, a string "  " and "" and "null" and "nil" will be considered empty
func IsNilOrEmpty(s *string) bool {
	return s == nil || strings.TrimSpace(*s) == "" || strings.TrimSpace(*s) == "null" || strings.TrimSpace(*s) == "nil"
}

// ValidateFormDataFields returns error if data from form data is invalid
func ValidateFormDataFields(outFormat, description *string) error {
	if IsNilOrEmpty(outFormat) {
		return ValidateBusinessError(constant.ErrMissingRequiredFields, "")
	}

	if IsNilOrEmpty(description) {
		return ValidateBusinessError(constant.ErrMissingRequiredFields, "")
	}

	if !IsOutputFormatValuesValid(outFormat) {
		return ValidateBusinessError(constant.ErrInvalidOutputFormat, "")
	}

	return nil
}

// IsOutputFormatValuesValid returns a boolean indicating if the output format value is valid
func IsOutputFormatValuesValid(outFormat *string) bool {
	outFormatUpper := strings.ToUpper(*outFormat)
	return outFormatUpper == "HTML" || outFormatUpper == "PDF" || outFormatUpper == "CSV" || outFormatUpper == "XML" || outFormatUpper == "TXT"
}

// ValidateFileFormat returns error if the templateFile content is not the same of outputFormat
func ValidateFileFormat(outFormat, templateFile string) error {
	format := strings.ToUpper(outFormat)

	switch format {
	case "HTML":
		if !strings.Contains(templateFile, "<html") && !strings.Contains(templateFile, "<!DOCTYPE html") {
			return ValidateBusinessError(constant.ErrFileContentInvalid, "", outFormat)
		}
	case "PDF":
		if !strings.Contains(templateFile, "<html") && !strings.Contains(templateFile, "<!DOCTYPE html") {
			return ValidateBusinessError(constant.ErrFileContentInvalid, "", outFormat)
		}
	case "XML":
		if !strings.Contains(templateFile, "<?xml") && !strings.Contains(templateFile, "<") {
			return ValidateBusinessError(constant.ErrFileContentInvalid, "", outFormat)
		}
	case "CSV":
		lines := strings.Split(templateFile, "\n")
		if len(lines) < 2 || !strings.Contains(lines[0], ",") && !strings.Contains(lines[0], ";") {
			return ValidateBusinessError(constant.ErrFileContentInvalid, "", outFormat)
		}
	case "TXT":
		trimmed := strings.TrimSpace(templateFile)
		if len(trimmed) == 0 {
			return ValidateBusinessError(constant.ErrFileContentInvalid, "", outFormat)
		}
	}

	return nil
}

// ValidateServerAddress checks if the value matches the pattern <some-address>:<some-port> and returns the value if it does.
func ValidateServerAddress(value string) string {
	matched, _ := regexp.MatchString(`^[^:]+:\d+$`, value)
	if !matched {
		return ""
	}

	return value
}

// SafeInt64ToInt safely converts int64 to int
func SafeInt64ToInt(val int64) int {
	if val > math.MaxInt {
		return math.MaxInt
	} else if val < math.MinInt {
		return math.MinInt
	}

	return int(val)
}

// Supported database types
const (
	// PostgreSQLType represents PostgreSQL database type
	PostgreSQLType = "postgresql"

	// MongoDBType represents the MongoDB database type constant.
	MongoDBType = "mongodb"
)
