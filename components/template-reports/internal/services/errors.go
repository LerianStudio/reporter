package services

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"plugin-template-engine/pkg"
	"plugin-template-engine/pkg/constant"
)

// ErrDatabaseItemNotFound is throws an item informed was not found
var ErrDatabaseItemNotFound = errors.New("errDatabaseItemNotFound")

// ValidatePGError validate pgError and return business error
func ValidatePGError(pgErr *pgconn.PgError, entityType string) error {
	switch pgErr.ConstraintName {
	case "example_parent_example_id_fkey":
		return pkg.ValidateBusinessError(constant.ErrParentExampleIDNotFound, entityType)
	default:
		return pgErr
	}
}
