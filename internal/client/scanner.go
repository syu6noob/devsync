package client

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/syu6noob/devsync/internal/core"
	"github.com/syu6noob/devsync/internal/ignore"
)

type ScanResult struct {
	Files        map[string]core.FileMeta
	IgnoredCount int
}

func Scan(root string, matcher *ignore.Matcher) (ScanResult, error) {
	result := ScanResult{Files: map[string]core.FileMeta{}}
	err := filepath.WalkDir(root, func(abs string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if abs == root {
			return nil
		}
		rel, err := filepath.Rel(root, abs)
		if err != nil {
			return err
		}
		rel = strings.ReplaceAll(rel, "\\", "/")
		rel = core.NormalizeRelativePath(rel)
		isDir := d.IsDir()
		if matcher != nil && matcher.Match(rel, isDir) {
			result.IgnoredCount++
			if isDir {
				return filepath.SkipDir
			}
			return nil
		}
		if isDir {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		hash, err := core.SHA256File(abs)
		if err != nil {
			return err
		}
		result.Files[rel] = core.FileMeta{
			Path:   rel,
			Size:   info.Size(),
			MTime:  info.ModTime().UnixNano(),
			SHA256: hash,
		}
		return nil
	})
	return result, err
}
