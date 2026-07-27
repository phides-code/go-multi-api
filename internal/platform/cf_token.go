// API token validation via the X-CF-Token request header.
package platform

import (
	"crypto/subtle"
	"os"
	"strings"
)

const CFTTokenHeader = "X-CF-Token"

// CFTTokenEnvVar is the process env key for the expected API token (SAM maps AwsCfToken here).
const CFTTokenEnvVar = "AWS_CF_TOKEN"

func ExpectedCFTToken() string {
	return os.Getenv(CFTTokenEnvVar)
}

func ValidCFTToken(expected string, headers map[string]string) bool {
	if expected == "" {
		return false
	}

	provided := HeaderValue(headers, CFTTokenHeader)
	if provided == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(expected), []byte(provided)) == 1
}

func HeaderValue(headers map[string]string, name string) string {
	for key, value := range headers {
		if strings.EqualFold(key, name) {
			return value
		}
	}
	return ""
}
