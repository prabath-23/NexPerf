package version

import "runtime"

var (
	Version = "0.3.0"
	Commit  = "unknown"
)

func Info() map[string]string {
	return map[string]string{
		"version": Version,
		"commit":  Commit,
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
	}
}
