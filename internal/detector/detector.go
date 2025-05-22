package detector

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

var DEFAULT_LOG_LOCATIONS = []string{"/var/log/postgresql", "/var/lib/pgsql"}

const DEFAULT_LOGFILE_SUFFIX = ".log"

// Returns the most recent file in the specified folder, plus it's parent folder
func FindMostRecentFileInFolders(foldersToScan []string) (latestFile string, latestFileParentFolder string, err error) {
	var latestModTime time.Time

	for _, folder := range foldersToScan {
		err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && info.ModTime().After(latestModTime) && strings.HasSuffix(path, DEFAULT_LOGFILE_SUFFIX) && strings.Contains(path, "postgresql") {
				latestFile = path
				latestModTime = info.ModTime()
			}
			return nil
		})
		if err != nil {
			continue
		}
	}

	if latestFile == "" {
		return "", "", os.ErrNotExist
	}

	return latestFile, filepath.Dir(latestFile), nil
}

// Folder will be used for tailing in the future
func DiscoverLatestLogFileAndFolder(fileOrFolder []string) (string, string, error) {
	var logFile string

	if len(fileOrFolder) == 0 {
		log.Debug().Msgf("No log file or folder specified, using default log locations: %v", DEFAULT_LOG_LOCATIONS)
		return FindMostRecentFileInFolders(DEFAULT_LOG_LOCATIONS)
	}

	if len(fileOrFolder) > 0 {
		_, err := os.Stat(fileOrFolder[0])
		if !os.IsNotExist(err) {
			logFile = fileOrFolder[0]
		}
		return logFile, filepath.Dir(logFile), nil
	}

	return FindMostRecentFileInFolders(DEFAULT_LOG_LOCATIONS)
}
