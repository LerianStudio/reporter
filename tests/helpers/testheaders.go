package helpers

import (
	"fmt"
)

func AuthHeaders(seed string) map[string]string {
	return map[string]string{
		"X-Request-Id": fmt.Sprintf("req-%s", seed),
	}
}

func AuthHeadersWithOrg(seed, orgID string) map[string]string {
	h := AuthHeaders(seed)
	h["X-Organization-Id"] = orgID

	return h
}
