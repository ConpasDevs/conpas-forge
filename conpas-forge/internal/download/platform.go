package download

import "fmt"

func ArchiveExt(goos string) string {
	if goos == "windows" {
		return "zip"
	}
	return "tar.gz"
}

func BinaryName(goos string) string {
	if goos == "windows" {
		return "engram.exe"
	}
	return "engram"
}

func AssetPattern(goos, goarch string) string {
	return fmt.Sprintf("engram_%s_%s", goos, goarch)
}
