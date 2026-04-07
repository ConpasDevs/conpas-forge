package config

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	resolvedHomeDir string
	homeDirOnce     sync.Once
)

func HomeDir() string {
	homeDirOnce.Do(func() {
		dir, err := os.UserHomeDir()
		if err != nil {
			panic("conpas-forge: cannot determine home directory: " + err.Error())
		}
		resolvedHomeDir = dir
	})
	return resolvedHomeDir
}

func OverrideHomeDir(dir string) {
	resolvedHomeDir = dir
	homeDirOnce.Do(func() {})
}

func ForgeDir() string   { return filepath.Join(HomeDir(), ".conpas-forge") }
func ConfigPath() string { return filepath.Join(ForgeDir(), "config.yaml") }
func BinDir() string     { return filepath.Join(ForgeDir(), "bin") }

func EngramBinary() string {
	name := "engram"
	if runtime.GOOS == "windows" {
		name = "engram.exe"
	}
	return filepath.Join(BinDir(), name)
}

func ClaudeJSON() string          { return filepath.Join(HomeDir(), ".claude.json") }
func ClaudeDir() string           { return filepath.Join(HomeDir(), ".claude") }
func ClaudeMD() string            { return filepath.Join(ClaudeDir(), "CLAUDE.md") }
func SettingsJSON() string        { return filepath.Join(ClaudeDir(), "settings.json") }
func SettingsJSONBak() string     { return filepath.Join(ClaudeDir(), "settings.json.bak") }
func OutputStylesDir() string     { return filepath.Join(ClaudeDir(), "output-styles") }
func SkillsDir() string           { return filepath.Join(ClaudeDir(), "skills") }
func SkillDir(name string) string { return filepath.Join(SkillsDir(), name) }
func SharedSkillsDir() string     { return filepath.Join(SkillsDir(), "_shared") }
