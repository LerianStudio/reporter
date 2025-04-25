package in

import (
	"github.com/LerianStudio/lib-commons/commons"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"plugin-template-engine/pkg"
	"plugin-template-engine/pkg/constant"
	"plugin-template-engine/pkg/net/http"
)

var (
	OrgIDHeaderParameter = "X-Organization-Id"
)

// ParseHeaderParameters convert and validate if the header parameters is UUID
func ParseHeaderParameters(c *fiber.Ctx) error {
	headerParam := c.Get(OrgIDHeaderParameter)

	if commons.IsNilOrEmpty(&headerParam) {
		err := pkg.ValidateBusinessError(constant.ErrInvalidHeaderParameter, "", OrgIDHeaderParameter)
		return http.WithError(c, err)
	}

	parsedHeaderUUID, errHeader := uuid.Parse(headerParam)
	if errHeader != nil {
		err := pkg.ValidateBusinessError(constant.ErrInvalidHeaderParameter, "", OrgIDHeaderParameter)
		return http.WithError(c, err)
	}

	c.Locals(OrgIDHeaderParameter, parsedHeaderUUID)

	return c.Next()
}
