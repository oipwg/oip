package version

import "strings"

var (
	GitCommitHash = "No Commit Hash"
	BuildDate     = "No Build Date"
	BuiltBy       = "Someone"
	GoVersion     = "Unknown"
)

func init() {
	// `go version` includes redundant prefix, strip if exists
	// go version go1.10.3 windows/amd64
	GoVersion = strings.TrimPrefix(GoVersion, "go version ")
}
