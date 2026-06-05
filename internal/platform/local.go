// Local execution detection (e.g. sam local sets AWS_SAM_LOCAL).
package platform

import "os"

const samLocalEnvVar = "AWS_SAM_LOCAL"

// LocalMode reports whether the Lambda is running under SAM local (AWS_SAM_LOCAL is set).
func LocalMode() bool {
	switch os.Getenv(samLocalEnvVar) {
	case "true", "1":
		return true
	default:
		return false
	}
}
