package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type PathType string

const (
	PathFile PathType = "file"
	PathDir  PathType = "directory"
)

type PathState int

const (
	PathStateOk PathState = iota
	PathStateMissing
	PathStateConflict
	PathStateError
)

type PathCheck struct {
	Path         string
	Type         PathType
	ExpectedType PathType
	State        PathState
	Err          error
}

type RequiredPath struct {
	Path string
	Type PathType
}

var Required = [...]RequiredPath{
	{Path: "posts/", Type: PathDir},
	{Path: "index.md", Type: PathFile},
	{Path: "themes/default/index.html", Type: PathFile},
	{Path: "themes/default/post.html", Type: PathFile},
}

func CheckPath(path string, expected PathType) PathCheck {
	info, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			return PathCheck{
				Path:  path,
				State: PathStateMissing,
				Type:  expected,
			}
		}

		return PathCheck{
			Path:  path,
			State: PathStateError,
			Type:  expected,
			Err:   err,
		}
	}

	if info.IsDir() {
		if expected != PathDir {
			return PathCheck{
				Path:         path,
				State:        PathStateConflict,
				Type:         PathDir,
				ExpectedType: expected,
			}
		}
	} else {
		if expected != PathFile {
			return PathCheck{
				Path:         path,
				State:        PathStateConflict,
				Type:         PathFile,
				ExpectedType: expected,
			}
		}
	}

	return PathCheck{
		Path:  path,
		State: PathStateOk,
		Type:  expected,
	}
}

func ValidatePath(path string, expected PathType) (string, error) {
	if strings.TrimSpace(path) == "" {
		switch expected {
		case PathDir:
			return ".", nil
		case PathFile:
			return "", fmt.Errorf("file path can not be empty")
		default:
			return "", fmt.Errorf("unknown PathType: %v", expected)
		}
	}

	cpath := filepath.Clean(path)

	info, err := os.Stat(cpath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("path '%s' does not exist", cpath)
		}
		if errors.Is(err, fs.ErrPermission) {
			return "", fmt.Errorf("permission denied to access '%s'", cpath)
		}

		return "", fmt.Errorf("cannot access path '%s': %w", cpath, err)
	}

	switch expected {
	case PathDir:
		if !info.IsDir() {
			return "", fmt.Errorf("path '%s' is a file, expected a directory", cpath)
		}
	case PathFile:
		if info.IsDir() {
			return "", fmt.Errorf("path '%s' is a directory, expected a file", cpath)
		}
	default:
		return "", fmt.Errorf("unknown PathType: %v", expected)
	}

	return cpath, nil
}

func CopyFile(src, dst string) error {
	s, err := os.OpenFile(src, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	if err != nil {
		return err
	}

	return nil
}

func CopyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("'%s' is a file, expected a directory", src)
	}

	if err := os.CopyFS(dst, os.DirFS(src)); err != nil {
		return err
	}
	return nil
}
