package smb_connection

import (
	"bytes"
	"compress/zlib"
	"io"
	"os"
)

// FileReader defines an interface for reading files
type FileReader interface {
	ReadFile(filePath string) (io.ReadCloser, error)
}

// OSFileReader is a concrete implementation of FileReader using the OS
type OSFileReader struct{}

func (OSFileReader) ReadFile(filePath string) (io.ReadCloser, error) {
	return os.Open(filePath)
}

// unzipLeagueSaveGame decompresses a file using zlib compression
func unzipLeagueSaveGame(filePath string, reader FileReader) ([]byte, error) {
	file, err := reader.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		return nil, err
	}

	zlibReader, err := zlib.NewReader(&buf)
	if err != nil {
		return nil, err
	}
	defer zlibReader.Close()

	var decompressed bytes.Buffer
	_, err = io.Copy(&decompressed, zlibReader)
	if err != nil {
		return nil, err
	}

	return decompressed.Bytes(), nil
}
