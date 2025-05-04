package smb_connection

import (
	"errors"
	"os"
	"testing"
)

// MockFileWriter is a mock implementation of FileWriter for testing
type MockFileWriter struct {
	MkdirAllCalled  bool
	MkdirAllPath    string
	MkdirAllError   error
	WriteFileCalled bool
	WriteFilePath   string
	WriteFileData   []byte
	WriteFileError  error
}

func (m *MockFileWriter) MkdirAll(path string, perm os.FileMode) error {
	m.MkdirAllCalled = true
	m.MkdirAllPath = path
	return m.MkdirAllError
}

func (m *MockFileWriter) WriteFile(filename string, data []byte, perm os.FileMode) error {
	m.WriteFileCalled = true
	m.WriteFilePath = filename
	m.WriteFileData = data
	return m.WriteFileError
}

func TestPersistLeagueSaveGame(t *testing.T) {
	testCases := []struct {
		name          string
		data          []byte
		filePath      string
		mkdirError    error
		writeError    error
		expectedPath  string
		expectedError bool
	}{
		{
			name:          "Success with .db extension",
			data:          []byte("test data"),
			filePath:      "/path/to/file.db",
			expectedPath:  "/path/to/file.db",
			expectedError: false,
		},
		{
			name:          "Success adding .db extension",
			data:          []byte("test data"),
			filePath:      "/path/to/file",
			expectedPath:  "/path/to/file.db",
			expectedError: false,
		},
		{
			name:          "MkdirAll error",
			data:          []byte("test data"),
			filePath:      "/path/to/file.db",
			mkdirError:    errors.New("mkdir error"),
			expectedError: true,
		},
		{
			name:          "WriteFile error",
			data:          []byte("test data"),
			filePath:      "/path/to/file.db",
			writeError:    errors.New("write error"),
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockWriter := &MockFileWriter{
				MkdirAllError:  tc.mkdirError,
				WriteFileError: tc.writeError,
			}

			err := persistLeagueSaveGame(tc.data, tc.filePath, mockWriter)

			if tc.expectedError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tc.expectedError && err != nil {
				t.Errorf("Expected no error but got %v", err)
			}

			if !tc.expectedError {
				if !mockWriter.MkdirAllCalled {
					t.Error("MkdirAll was not called")
				}

				if !mockWriter.WriteFileCalled {
					t.Error("WriteFile was not called")
				}

				if mockWriter.WriteFilePath != tc.expectedPath {
					t.Errorf("Expected path %s, got %s", tc.expectedPath, mockWriter.WriteFilePath)
				}

				if string(mockWriter.WriteFileData) != string(tc.data) {
					t.Errorf("Expected data %s, got %s", string(tc.data), string(mockWriter.WriteFileData))
				}
			}
		})
	}
}
