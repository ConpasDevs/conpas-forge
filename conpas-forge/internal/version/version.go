package version

// Version is injected at build time via -ldflags "-X ...version.Version=vX.Y.Z".
// Defaults to "dev" for local builds without ldflags injection.
var Version = "dev"
