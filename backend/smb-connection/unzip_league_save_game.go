package smb_connection

import (
	"bytes"
	"compress/zlib"
	"io"
	"os"
)

func unzipLeagueSaveGame(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
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
