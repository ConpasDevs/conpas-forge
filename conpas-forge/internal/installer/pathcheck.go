package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func CheckPathContains(dir string) bool {
	pathEnv := os.Getenv("PATH")
	sep := string(os.PathListSeparator)
	dirs := strings.Split(pathEnv, sep)
	dirAbs, _ := filepath.Abs(dir)

	for _, d := range dirs {
		dAbs, _ := filepath.Abs(d)
		if runtime.GOOS == "windows" {
			if strings.EqualFold(dAbs, dirAbs) {
				return true
			}
		} else {
			if dAbs == dirAbs {
				return true
			}
		}
	}
	return false
}

func PathWarning(binDir string, goos string) string {
	if goos == "windows" {
		return fmt.Sprintf(
			"WARNING: %s is not in your PATH.\n"+
				"Add it permanently with PowerShell:\n"+
				"  [Environment]::SetEnvironmentVariable('PATH', $env:PATH + ';%s', 'User')\n"+
				"Then restart your terminal.",
			binDir, binDir)
	}
	return fmt.Sprintf(
		"WARNING: %s is not in your PATH.\n"+
			"Add to your shell profile (~/.bashrc or ~/.zshrc):\n"+
			"  export PATH=\"%s:$PATH\"\n"+
			"Then: source ~/.bashrc",
		binDir, binDir)
}
