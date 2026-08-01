// Cross-package test constants and small auth-header helpers.
package testutil

import "github.com/phides-code/go-multi-api/internal/platform"

// TestCFTToken is the expected X-CF-Token value for gateway, router, and composition tests.
// Pair with platform.CFTTokenEnvVar via t.Setenv in composition tests.
const TestCFTToken = "test-token"

// CFTokenHeaders returns request headers carrying the given CFT token.
func CFTokenHeaders(token string) map[string]string {
	return map[string]string{platform.CFTTokenHeader: token}
}
