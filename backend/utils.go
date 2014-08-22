package backend

import (
	//"path"
	"path/filepath"
)

func AdaptPathForPrison(rootPath string, initialPath string) string {
	if !filepath.IsAbs(initialPath) {
		initialPath = filepath.Join(rootPath, initialPath)
	}

	initialPath = filepath.FromSlash(initialPath)
	return initialPath
}
