package smb_connection

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DirPermissions represents read, write, execute permissions for owner and read, execute for group and others (0755)
	DirPermissions os.FileMode = 0755
	// FilePermissions represents read, write permissions for owner and read for group and others (0644)
	FilePermissions os.FileMode = 0644
)

// FileWriter defines an interface for writing files
type FileWriter interface {
	MkdirAll(path string, perm os.FileMode) error
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

// OSFileWriter is a concrete implementation of FileWriter using the OS
type OSFileWriter struct{}

func (OSFileWriter) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (OSFileWriter) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func persistLeagueSaveGame(data []byte, filePath string, writer FileWriter) error {
	if !strings.HasSuffix(filePath, ".db") {
		filePath = filePath + ".db"
	}

	dir := filepath.Dir(filePath)
	if err := writer.MkdirAll(dir, DirPermissions); err != nil {
		return errors.New("failed to create directory: " + err.Error())
	}

	err := writer.WriteFile(filePath, data, FilePermissions)
	if err != nil {
		return errors.New("failed to write file: " + err.Error())
	}

	return nil
}
