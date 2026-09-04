package version

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	// Version is the current semantic version of the NodePhone CLI.
	Version = "v1.0.0"
)

// Info holds version details.
type Info struct {
	Version   string
	GoVersion string
	OS        string
	Arch      string
}

// GetInfo returns version information structure.
func GetInfo() Info {
	goVer := runtime.Version()
	// Strip "go" prefix if present for clean "1.xx" display as requested in PRD
	goVerClean := strings.TrimPrefix(goVer, "go")

	return Info{
		Version:   Version,
		GoVersion: goVerClean,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String returns formatted version summary.
func (i Info) String() string {
	var sb strings.Builder
	sb.WriteString("NodePhone CLI\n\n")
	sb.WriteString(fmt.Sprintf("Version : %s\n", i.Version))
	sb.WriteString(fmt.Sprintf("Go      : %s\n", i.GoVersion))
	sb.WriteString(fmt.Sprintf("OS      : %s", i.OS))
	return sb.String()
}
