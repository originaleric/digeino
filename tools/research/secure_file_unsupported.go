//go:build !unix

package research

import (
	"fmt"
	"os"
)

func secureFileAccessSupported() bool {
	return false
}

func openAllowedReadFile(path string) (*os.File, error) {
	return nil, fmt.Errorf("secure file access is not supported on this platform")
}

func openAllowedWriteFile(path string, mode string) (*os.File, string, error) {
	return nil, "", fmt.Errorf("secure file access is not supported on this platform")
}
