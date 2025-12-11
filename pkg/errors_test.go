package pkg

import (
	"errors"
	"testing"

	"github.com/LerianStudio/reporter/v4/pkg/constant"

	"github.com/stretchr/testify/assert"
)

func TestEntityNotFoundError_Error_WithMessage(t *testing.T) {
	err := EntityNotFoundError{
		Message: "Custom message",
	}

	assert.Equal(t, "Custom message", err.Error())
}

func TestEntityNotFoundError_Error_WithEntityType(t *testing.T) {
	err := EntityNotFoundError{
		EntityType: "User",
	}

	assert.Equal(t, "Entity User not found", err.Error())
}

func TestEntityNotFoundError_Error_WithWrappedError(t *testing.T) {
	innerErr := errors.New("inner error")
	err := EntityNotFoundError{
		Err: innerErr,
	}

	assert.Equal(t, "inner error", err.Error())
}

func TestEntityNotFoundError_Error_Default(t *testing.T) {
	err := EntityNotFoundError{}

	assert.Equal(t, "entity not found", err.Error())
}

func TestEntityNotFoundError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	err := EntityNotFoundError{
		Err: innerErr,
	}

	assert.Equal(t, innerErr, err.Unwrap())
}

func TestValidationError_Error_WithCode(t *testing.T) {
	err := ValidationError{
		Code:    "VALIDATION_001",
		Message: "Invalid field",
	}

	assert.Equal(t, "VALIDATION_001 - Invalid field", err.Error())
}

func TestValidationError_Error_WithoutCode(t *testing.T) {
	err := ValidationError{
		Message: "Invalid field",
	}

	assert.Equal(t, "Invalid field", err.Error())
}

func TestValidationError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	err := ValidationError{
		Err: innerErr,
	}

	assert.Equal(t, innerErr, err.Unwrap())
}

func TestEntityConflictError_Error_WithMessage(t *testing.T) {
	err := EntityConflictError{
		Message: "Entity already exists",
	}

	assert.Equal(t, "Entity already exists", err.Error())
}

func TestEntityConflictError_Error_WithWrappedError(t *testing.T) {
	innerErr := errors.New("inner error")
	err := EntityConflictError{
		Err: innerErr,
	}

	assert.Equal(t, "inner error", err.Error())
}

func TestEntityConflictError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	err := EntityConflictError{
		Err: innerErr,
	}

	assert.Equal(t, innerErr, err.Unwrap())
}

func TestUnauthorizedError_Error(t *testing.T) {
	err := UnauthorizedError{
		Message: "Unauthorized",
	}

	assert.Equal(t, "Unauthorized", err.Error())
}

func TestForbiddenError_Error(t *testing.T) {
	err := ForbiddenError{
		Message: "Forbidden",
	}

	assert.Equal(t, "Forbidden", err.Error())
}

func TestUnprocessableOperationError_Error(t *testing.T) {
	err := UnprocessableOperationError{
		Message: "Unprocessable",
	}

	assert.Equal(t, "Unprocessable", err.Error())
}

func TestHTTPError_Error(t *testing.T) {
	err := HTTPError{
		Message: "HTTP Error",
	}

	assert.Equal(t, "HTTP Error", err.Error())
}

func TestFailedPreconditionError_Error(t *testing.T) {
	err := FailedPreconditionError{
		Message: "Precondition failed",
	}

	assert.Equal(t, "Precondition failed", err.Error())
}

func TestInternalServerError_Error(t *testing.T) {
	err := InternalServerError{
		Message: "Internal error",
	}

	assert.Equal(t, "Internal error", err.Error())
}

func TestResponseError_Error(t *testing.T) {
	err := ResponseError{
		Message: "Response error",
	}

	assert.Equal(t, "Response error", err.Error())
}

func TestValidationKnownFieldsError_Error(t *testing.T) {
	err := ValidationKnownFieldsError{
		Message: "Validation error",
	}

	assert.Equal(t, "Validation error", err.Error())
}

func TestValidationUnknownFieldsError_Error(t *testing.T) {
	err := ValidationUnknownFieldsError{
		Message: "Unknown fields error",
	}

	assert.Equal(t, "Unknown fields error", err.Error())
}

func TestValidateInternalError(t *testing.T) {
	innerErr := errors.New("db connection failed")

	err := ValidateInternalError(innerErr, "User")

	assert.IsType(t, InternalServerError{}, err)

	internalErr := err.(InternalServerError)
	assert.Equal(t, "User", internalErr.EntityType)
	assert.Equal(t, constant.ErrInternalServer.Error(), internalErr.Code)
	assert.Equal(t, innerErr, internalErr.Err)
}

func TestValidateBadRequestFieldsError_EmptyFields(t *testing.T) {
	err := ValidateBadRequestFieldsError(nil, nil, "User", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected knownInvalidFields")
}

func TestValidateBadRequestFieldsError_UnknownFields(t *testing.T) {
	unknownFields := map[string]any{"extra_field": "value"}

	err := ValidateBadRequestFieldsError(nil, nil, "User", unknownFields)

	assert.IsType(t, ValidationUnknownFieldsError{}, err)
	valErr := err.(ValidationUnknownFieldsError)
	assert.Equal(t, constant.ErrUnexpectedFieldsInTheRequest.Error(), valErr.Code)
}

func TestValidateBadRequestFieldsError_RequiredFields(t *testing.T) {
	requiredFields := map[string]string{"name": "required"}

	err := ValidateBadRequestFieldsError(requiredFields, nil, "User", nil)

	assert.IsType(t, ValidationKnownFieldsError{}, err)
	valErr := err.(ValidationKnownFieldsError)
	assert.Equal(t, constant.ErrMissingFieldsInRequest.Error(), valErr.Code)
}

func TestValidateBadRequestFieldsError_KnownInvalidFields(t *testing.T) {
	knownInvalidFields := map[string]string{"email": "invalid format"}

	err := ValidateBadRequestFieldsError(nil, knownInvalidFields, "User", nil)

	assert.IsType(t, ValidationKnownFieldsError{}, err)
	valErr := err.(ValidationKnownFieldsError)
	assert.Equal(t, constant.ErrBadRequest.Error(), valErr.Code)
}

func TestValidateBusinessError_KnownErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		entityType string
		args       []any
	}{
		{
			name:       "ErrInvalidQueryParameter",
			err:        constant.ErrInvalidQueryParameter,
			entityType: "Report",
			args:       []any{"param1"},
		},
		{
			name:       "ErrInvalidDateFormat",
			err:        constant.ErrInvalidDateFormat,
			entityType: "Report",
		},
		{
			name:       "ErrEntityNotFound",
			err:        constant.ErrEntityNotFound,
			entityType: "Template",
			args:       []any{"template"},
		},
		{
			name:       "ErrMissingRequiredFields",
			err:        constant.ErrMissingRequiredFields,
			entityType: "Report",
		},
		{
			name:       "ErrInvalidFileFormat",
			err:        constant.ErrInvalidFileFormat,
			entityType: "Template",
		},
		{
			name:       "ErrInvalidOutputFormat",
			err:        constant.ErrInvalidOutputFormat,
			entityType: "Template",
		},
		{
			name:       "ErrScriptTagDetected",
			err:        constant.ErrScriptTagDetected,
			entityType: "Template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateBusinessError(tt.err, tt.entityType, tt.args...)
			assert.NotNil(t, result)
			assert.NotEqual(t, tt.err, result) // Should be wrapped
		})
	}
}

func TestValidateBusinessError_UnknownError(t *testing.T) {
	unknownErr := errors.New("unknown error")

	result := ValidateBusinessError(unknownErr, "Test")

	// Should return the same error
	assert.Equal(t, unknownErr, result)
}
