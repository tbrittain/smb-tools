package smb_connection

import (
	"bytes"
	"compress/zlib"
	"io"
	"testing"
)

// MockFileReader is a mock implementation of FileReader
type MockFileReader struct {
	Content []byte
	Err     error
}

func (m MockFileReader) ReadFile(filePath string) (io.ReadCloser, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return io.NopCloser(bytes.NewReader(m.Content)), nil
}

func TestUnzipLeagueSaveGame(t *testing.T) {
	// Create compressed test data
	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	_, err := writer.Write([]byte("test data"))
	if err != nil {
		t.Fatalf("Failed to write compressed data: %v", err)
	}
	writer.Close()

	mockReader := MockFileReader{Content: compressed.Bytes()}
	result, err := unzipLeagueSaveGame("mock/path", mockReader)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := "test data"
	if string(result) != expected {
		t.Errorf("Expected %q, got %q", expected, string(result))
	}
}
