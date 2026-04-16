package installer

import (
	"context"
	"runtime"

	"github.com/conpasDEVS/conpas-forge/internal/config"
)

type Result struct {
	ModuleName       string
	Success          bool
	PathsWritten     []string
	Warnings         []string
	Err              error
	InstalledVersion string
}

type Module interface {
	Name() string
	Install(ctx context.Context, opts *InstallOptions, progress func(ProgressEvent)) Result
}

type InstallOptions struct {
	Config   *config.Config
	Persona  string
	Models   map[string]string
	Platform Platform
	HomeDir  string
}

type Platform struct {
	OS   string
	Arch string
}

type ProgressEvent struct {
	Module  string
	Status  string
	Detail  string
	Percent int
}

func DetectPlatform() Platform {
	return Platform{OS: runtime.GOOS, Arch: runtime.GOARCH}
}
