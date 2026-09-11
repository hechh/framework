package base

import (
	"path/filepath"

	"github.com/hechh/framework/library/fileutil"
)

func EnsureDir(path string) (string, error) {
	abspath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if err := fileutil.EnsureDir(abspath); err != nil {
		return "", err
	}
	return abspath, nil
}
