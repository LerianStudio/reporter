package http

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDecodeCursor_ValidCursor(t *testing.T) {
	original := Cursor{
		ID:         "test-id-123",
		PointsNext: true,
	}

	// Encode to JSON then base64
	jsonData, err := json.Marshal(original)
	assert.NoError(t, err)
	encoded := base64.StdEncoding.EncodeToString(jsonData)

	// Decode
	result, err := DecodeCursor(encoded)
	assert.NoError(t, err)
	assert.Equal(t, original.ID, result.ID)
	assert.Equal(t, original.PointsNext, result.PointsNext)
}

func TestDecodeCursor_PointsNextFalse(t *testing.T) {
	original := Cursor{
		ID:         "another-id",
		PointsNext: false,
	}

	jsonData, err := json.Marshal(original)
	assert.NoError(t, err)
	encoded := base64.StdEncoding.EncodeToString(jsonData)

	result, err := DecodeCursor(encoded)
	assert.NoError(t, err)
	assert.Equal(t, original.ID, result.ID)
	assert.False(t, result.PointsNext)
}

func TestDecodeCursor_InvalidBase64(t *testing.T) {
	invalidBase64 := "not-valid-base64!!!"

	_, err := DecodeCursor(invalidBase64)
	assert.Error(t, err)
}

func TestDecodeCursor_InvalidJSON(t *testing.T) {
	// Valid base64 but invalid JSON
	invalidJSON := base64.StdEncoding.EncodeToString([]byte("not json"))

	_, err := DecodeCursor(invalidJSON)
	assert.Error(t, err)
}

func TestDecodeCursor_EmptyString(t *testing.T) {
	_, err := DecodeCursor("")
	assert.Error(t, err)
}

func TestDecodeCursor_UUIDAsID(t *testing.T) {
	id := uuid.New().String()
	original := Cursor{
		ID:         id,
		PointsNext: true,
	}

	jsonData, err := json.Marshal(original)
	assert.NoError(t, err)
	encoded := base64.StdEncoding.EncodeToString(jsonData)

	result, err := DecodeCursor(encoded)
	assert.NoError(t, err)
	assert.Equal(t, id, result.ID)
}

func TestCursor_JSONTags(t *testing.T) {
	cursor := Cursor{
		ID:         "test-id",
		PointsNext: true,
	}

	data, err := json.Marshal(cursor)
	assert.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.Contains(t, result, "id")
	assert.Contains(t, result, "points_next")
	assert.Equal(t, "test-id", result["id"])
	assert.Equal(t, true, result["points_next"])
}
