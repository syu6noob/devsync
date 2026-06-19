package core

import (
	"errors"
	"path"
	"path/filepath"
	"strings"
)

func NormalizeRelativePath(p string) string {
	p = filepath.ToSlash(p)
	p = strings.TrimSpace(p)
	p = strings.TrimPrefix(p, "./")
	p = strings.TrimPrefix(p, "/")
	p = path.Clean(p)
	if p == "." {
		return ""
	}
	return p
}

func IsSafeRelativePath(p string) bool {
	p = NormalizeRelativePath(p)
	if p == "" {
		return false
	}
	if strings.HasPrefix(p, "../") || p == ".." {
		return false
	}
	if strings.Contains(p, "\x00") {
		return false
	}
	// Avoid Windows drive-like paths when received by the Linux server.
	if strings.Contains(strings.Split(p, "/")[0], ":") {
		return false
	}
	return true
}

func SafeJoin(root, rel string) (string, error) {
	rel = NormalizeRelativePath(rel)
	if !IsSafeRelativePath(rel) {
		return "", errors.New("unsafe relative path")
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	targetAbs, err := filepath.Abs(filepath.Join(rootAbs, filepath.FromSlash(rel)))
	if err != nil {
		return "", err
	}
	if targetAbs != rootAbs && !strings.HasPrefix(targetAbs, rootAbs+string(filepath.Separator)) {
		return "", errors.New("path escapes workspace")
	}
	return targetAbs, nil
}
