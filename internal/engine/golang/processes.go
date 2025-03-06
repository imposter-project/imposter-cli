package golang

import (
	"strconv"
	"strings"

	"gatehill.io/imposter/internal/engine/procutil"
)

var matcher = procutil.ProcessMatcher{
	ProcessName:    "imposter-go",
	CommandPattern: "imposter-go$",
	GetPort: func(cmdline []string, env []string) int {
		// Check environment variables
		for _, e := range env {
			if !strings.HasPrefix(e, "IMPOSTER_PORT=") {
				continue
			}
			portRaw := strings.TrimPrefix(e, "IMPOSTER_PORT=")
			if port, err := strconv.Atoi(portRaw); err == nil {
				return port
			}
		}
		return 8080
	},
}
