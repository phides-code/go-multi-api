// Local execution detection (e.g. sam local sets AWS_SAM_LOCAL).
package platform

import "os"

// SAMLocalEnvVar is the process env key SAM sets for local API emulation.
const SAMLocalEnvVar = "AWS_SAM_LOCAL"

// LocalMode reports whether the Lambda is running under SAM local
// (AWS_SAM_LOCAL is "true" or "1").
func LocalMode() bool {
	switch os.Getenv(SAMLocalEnvVar) {
	case "true", "1":
		return true
	default:
		return false
	}
}
