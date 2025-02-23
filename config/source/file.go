package source

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const FileSourceName = "file"

type File struct {
	path string
}

// NewFile creates a new File instance from the provided paths.
// It returns an error if no paths are provided or if none of the paths point to an existing file.
// The function checks each path and returns the first one that exists.
//
// Parameters:
// - paths: a variadic list of file paths to check.
//
// Returns:
// - *File: a pointer to the created File instance.
// - error: an error if no valid file path is found.
func NewFile(paths ...string) (*File, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("file path is required")
	}

	existPath := ""
	for _, path := range paths {
		if path == "" {
			continue
		}

		if absPath, ok := isFileExist(path); ok {
			existPath = absPath
			break
		}
	}

	if existPath == "" {
		return nil, fmt.Errorf("file not found")
	}

	return &File{path: existPath}, nil
}

func (f *File) Name() string {
	return FileSourceName
}

func (f *File) Load(v *viper.Viper) error {
	ext := strings.TrimLeft(strings.ToLower(path.Ext(f.path)), ".")
	// viper.SupportedExts

	check := false
	for _, e := range viper.SupportedExts {
		if e == ext {
			check = true
			break
		}
	}

	if !check {
		return fmt.Errorf("unsupported file type: %s", ext)
	}

	v.SetConfigFile(f.path)
	v.SetConfigType(ext)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	return nil
}

func isFileExist(path string) (absPath string, ok bool) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return
	}

	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return
	}
	if info.IsDir() {
		return
	}

	return absPath, true
}
