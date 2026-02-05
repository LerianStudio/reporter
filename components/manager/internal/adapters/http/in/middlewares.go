// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package in

import (
	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/net/http"

	"github.com/LerianStudio/lib-commons/v2/commons"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var UUIDPathParameter = "id"

// ParsePathParametersUUID convert and validate if the path parameter is UUID
func ParsePathParametersUUID(c *fiber.Ctx) error {
	pathParam := c.Params(UUIDPathParameter)

	if commons.IsNilOrEmpty(&pathParam) {
		err := pkg.ValidateBusinessError(constant.ErrInvalidPathParameter, "", UUIDPathParameter)
		return http.WithError(c, err)
	}

	parsedPathUUID, errPath := uuid.Parse(pathParam)
	if errPath != nil {
		err := pkg.ValidateBusinessError(constant.ErrInvalidPathParameter, "", UUIDPathParameter)
		return http.WithError(c, err)
	}

	c.Locals(UUIDPathParameter, parsedPathUUID)

	return c.Next()
}
