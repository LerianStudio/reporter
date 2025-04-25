package pkg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"math"
	"os/exec"
	"plugin-template-engine/pkg/constant"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// GenerateUUIDv7 generate a new uuid v7 using google/uuid package and return it. If an error occurs, it will return the error.
func GenerateUUIDv7() uuid.UUID {
	u := uuid.Must(uuid.NewV7())
	return u
}

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

// ReplaceUUIDWithPlaceholder replaces UUIDs with a placeholder in a given path string.
func ReplaceUUIDWithPlaceholder(path string) string {
	re := regexp.MustCompile(`[0-9a-fA-F-]{36}`)

	return re.ReplaceAllString(path, ":id")
}

// StructToJSONString convert a struct to json string
func StructToJSONString(s any) (string, error) {
	jsonByte, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	return string(jsonByte), nil
}

type SyscmdI interface {
	ExecCmd(name string, arg ...string) ([]byte, error)
}

type Syscmd struct{}

func (r *Syscmd) ExecCmd(name string, arg ...string) ([]byte, error) {
	return exec.Command(name, arg...).Output()
}

// GetCPUUsage get the current CPU usage
func GetCPUUsage(ctx context.Context, exc SyscmdI) int64 {
	logger := NewLoggerFromContext(ctx)

	out, err := exc.ExecCmd("sh", "-c", "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1}'")
	if err != nil {
		fmt.Println("Error executing command:", err)
		return 0
	}

	usageStr := strings.Split(strings.TrimSpace(string(out)), "\n")[0]

	usage, err := strconv.ParseFloat(usageStr, 64)
	if err != nil {
		logger.Errorf("Error parsing CPU usage: %v", err)

		return 0
	}

	return int64(usage)
}

// GetMemUsage get the current memory usage
func GetMemUsage(ctx context.Context, exc SyscmdI) int64 {
	logger := NewLoggerFromContext(ctx)

	out, err := exc.ExecCmd("sh", "-c", "free | grep Mem | awk '{print $3/$2 * 100.0}'")
	if err != nil {
		return 0
	}

	usageStr := strings.Split(strings.TrimSpace(string(out)), "\n")[0]

	usage, err := strconv.ParseFloat(usageStr, 64)
	if err != nil {
		logger.Errorf("Error parsing memory usage: %v", err)

		return 0
	}

	return int64(usage)
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
	return outFormatUpper == "HTML" || outFormatUpper == "JSON" || outFormatUpper == "XML"
}

// ValidateFileFormat returns error if the templateFile content is not the same of outputFormat
func ValidateFileFormat(outFormat, templateFile string) error {
	format := strings.ToUpper(outFormat)

	switch format {
	case "HTML":
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

// Contains checks if an item is in a slice. This function uses type parameters to work with any slice type.
func Contains[T comparable](slice []T, item T) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
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

// SafeIntToUint64 safe mode to converter int to uint64
func SafeIntToUint64(val int) uint64 {
	if val < 0 {
		return uint64(1)
	}

	return uint64(val)
}

// SafeIntToInt32 Function to safely convert int to int32 with overflow check
func SafeIntToInt32(val int) (int32, error) {
	if val > math.MaxInt32 || val < math.MinInt32 {
		return 0, errors.New("integer overflow: value out of range for int32")
	}

	return int32(val), nil
}
