//go:build unix

package research

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

func secureFileAccessSupported() bool {
	return true
}

func openAllowedReadFile(path string) (*os.File, error) {
	base, rel, err := allowedTarget(path, pathAccessRead)
	if err != nil {
		return nil, err
	}
	parts := splitCleanRel(rel)
	if len(parts) == 0 {
		return nil, fmt.Errorf("读取目标不能是允许目录本身")
	}

	dirFD, err := unix.Open(base, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, fmt.Errorf("无法打开允许读目录: %w", err)
	}
	currentFD := dirFD
	defer func() {
		if currentFD >= 0 {
			_ = unix.Close(currentFD)
		}
	}()

	for _, part := range parts[:len(parts)-1] {
		if part == "" || part == "." || part == ".." {
			return nil, fmt.Errorf("非法路径片段 %q", part)
		}
		nextFD, err := openChildDirNoFollow(currentFD, part)
		if err != nil {
			return nil, err
		}
		_ = unix.Close(currentFD)
		currentFD = nextFD
	}

	fileName := parts[len(parts)-1]
	if fileName == "" || fileName == "." || fileName == ".." {
		return nil, fmt.Errorf("非法文件名 %q", fileName)
	}
	fileFD, err := unix.Openat(currentFD, fileName, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件: %w", err)
	}
	file := os.NewFile(uintptr(fileFD), path)
	if file == nil {
		_ = unix.Close(fileFD)
		return nil, fmt.Errorf("无法创建文件句柄")
	}
	return file, nil
}

func openAllowedWriteFile(path string, mode string) (*os.File, string, error) {
	base, rel, absPath, err := allowedWriteTarget(path)
	if err != nil {
		return nil, "", fmt.Errorf("无法解析文件路径: %w", err)
	}
	parts := splitCleanRel(rel)
	if len(parts) == 0 {
		return nil, "", fmt.Errorf("写入目标不能是允许目录本身")
	}

	dirFD, err := unix.Open(base, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, "", fmt.Errorf("无法打开允许写目录: %w", err)
	}
	currentFD := dirFD
	defer func() {
		if currentFD >= 0 {
			_ = unix.Close(currentFD)
		}
	}()

	for _, part := range parts[:len(parts)-1] {
		if part == "" || part == "." || part == ".." {
			return nil, "", fmt.Errorf("非法路径片段 %q", part)
		}
		nextFD, err := openOrCreateChildDirNoFollow(currentFD, part)
		if err != nil {
			return nil, "", err
		}
		_ = unix.Close(currentFD)
		currentFD = nextFD
	}

	fileName := parts[len(parts)-1]
	if fileName == "" || fileName == "." || fileName == ".." {
		return nil, "", fmt.Errorf("非法文件名 %q", fileName)
	}
	flags := unix.O_WRONLY | unix.O_CREAT | unix.O_CLOEXEC | unix.O_NOFOLLOW
	if mode == "append" {
		flags |= unix.O_APPEND
	} else {
		flags |= unix.O_TRUNC
	}
	fileFD, err := unix.Openat(currentFD, fileName, flags, 0o644)
	if err != nil {
		return nil, "", fmt.Errorf("无法打开文件: %w", err)
	}
	file := os.NewFile(uintptr(fileFD), absPath)
	if file == nil {
		_ = unix.Close(fileFD)
		return nil, "", fmt.Errorf("无法创建文件句柄")
	}
	return file, absPath, nil
}

func openChildDirNoFollow(parentFD int, name string) (int, error) {
	fd, err := unix.Openat(parentFD, name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return -1, fmt.Errorf("无法打开目录 %q: %w", name, err)
	}
	return fd, nil
}

func openOrCreateChildDirNoFollow(parentFD int, name string) (int, error) {
	fd, err := openChildDirNoFollow(parentFD, name)
	if err == nil {
		return fd, nil
	}
	if !errors.Is(err, unix.ENOENT) {
		return -1, err
	}
	if err := unix.Mkdirat(parentFD, name, 0o755); err != nil && !errors.Is(err, unix.EEXIST) {
		return -1, fmt.Errorf("无法创建目录 %q: %w", name, err)
	}
	return openChildDirNoFollow(parentFD, name)
}

func allowedWriteTarget(path string) (base string, rel string, resolved string, err error) {
	base, rel, err = allowedTarget(path, pathAccessWrite)
	if err != nil {
		return "", "", "", err
	}
	resolved = filepath.Join(base, rel)
	return base, rel, resolved, nil
}

func splitCleanRel(rel string) []string {
	rel = filepath.Clean(rel)
	if rel == "." {
		return nil
	}
	return strings.Split(rel, string(os.PathSeparator))
}
