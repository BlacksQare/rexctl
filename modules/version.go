package modules

import (
	"fmt"
	"rexctl/config"
	"runtime/debug"
)

// GetBuildCommitHash returns the commit hash of the current build.
// Priority:
// 1. config.CommitHash if injected via ldflags (and != "dev")
// 2. VCS revision from Go runtime build info (debug.ReadBuildInfo)
// 3. Fallback to config.CommitHash ("dev")
func GetBuildCommitHash() string {
	if config.CommitHash != "" && config.CommitHash != "dev" {
		return config.CommitHash
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		var rev string
		var dirty bool
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				rev = setting.Value
			}
			if setting.Key == "vcs.modified" && setting.Value == "true" {
				dirty = true
			}
		}
		if rev != "" {
			if len(rev) > 7 {
				rev = rev[:7]
			}
			if dirty {
				rev += "-dirty"
			}
			return rev
		}
	}

	return config.CommitHash
}

// CmdVersion prints the commit hash of the current build.
func CmdVersion() {
	fmt.Println(GetBuildCommitHash())
}
