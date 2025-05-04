package smb_connection

import (
	"errors"
	"github.com/Microsoft/go-winio/pkg/guid"
	"path/filepath"
	"smb-tools/backend/config"
	"strings"
	"time"
)

const (
	unixTimestampFormat = "2006-01-02_15-04-05"
)

func HandleExportSaveGame(filePath string) (string, error) {
	if filePath == "" {
		return "", errors.New("file path cannot be empty")
	}

	fileName := filepath.Base(filePath)

	if !strings.HasSuffix(fileName, ".sav") {
		return "", errors.New("file name must end with .sav")
	}

	if !strings.HasPrefix(fileName, "league-") {
		return "", errors.New("file name must start with 'league-'")
	}

	guidString := strings.TrimSuffix(strings.TrimPrefix(fileName, "league-"), ".sav")
	parsed, err := guid.FromString(guidString)
	if err != nil {
		return "", errors.New("invalid GUID in file name: " + err.Error())
	}

	result, err := unzipLeagueSaveGame(filePath, OSFileReader{})
	if err != nil {
		return "", err
	}

	baseAppDataDir, err := config.GetSmbToolsDir()
	if err != nil {
		return "", err
	}

	now := time.Now()

	// the target directory for this file should be
	// baseAppDataDir/league-save-games/<guid>/<unix_timestamp>.db
	fileWriter := OSFileWriter{}
	saveGameDir := filepath.Join(baseAppDataDir, "league-save-games", parsed.String())
	if err := fileWriter.MkdirAll(saveGameDir, DirPermissions); err != nil {
		return "", errors.New("failed to create save game directory: " + err.Error())
	}

	saveGameFilePath := filepath.Join(saveGameDir, now.Format(unixTimestampFormat)+".db")
	if err := persistLeagueSaveGame(result, saveGameFilePath, fileWriter); err != nil {
		return "", errors.New("failed to persist league save game: " + err.Error())
	}

	return saveGameFilePath, nil
}
