package detector

import (
	"os"
	"path/filepath"
	"time"
)

// Functions to detect the location of Postgres log files automatically

// func ScanForPostgresInstances() ([]string, error) {
// }

const DEFAULT_LOG_DIR = "/var/log/postgresql"

func DetectLatestPostgresLogFile() (string, error) {
	var latestFile string
	var latestModTime time.Time

	err := filepath.Walk(DEFAULT_LOG_DIR, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.ModTime().After(latestModTime) {
			latestFile = path
			latestModTime = info.ModTime()
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if latestFile == "" {
		return "", os.ErrNotExist
	}

	return latestFile, nil
}
