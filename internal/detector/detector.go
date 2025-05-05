package detector

import (
	"os"
	"path/filepath"
	"time"

	"github.com/kmoppel/pgweasel/internal/postgres"
	"github.com/rs/zerolog/log"
)

// Functions to detect the location of Postgres log files automatically

// func ScanForPostgresInstances() ([]string, error) {
// }

const DEFAULT_LOG_DIR = "/var/log/postgresql"
const DEFAULT_LOG_LINE_PREFIX = "%m [%p] %q%u@%d "

// Returns the most recent file in the specified folder, plus it's parent folder
func FindMostRecentFileInFolder(folder string) (latestFile string, latestFileParentFolder string, err error) {
	// var latestFile string
	var latestModTime time.Time

	err = filepath.Walk(DEFAULT_LOG_DIR, func(path string, info os.FileInfo, err error) error {
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
		return "", "", err
	}

	if latestFile == "" {
		return "", "", os.ErrNotExist
	}

	return latestFile, filepath.Dir(latestFile), nil
}

func GetLatestLogFileAndFolder(fileOrFolder []string, pgConnstr string) (string, string, error) {
	var logFolder, logFile string

	if len(fileOrFolder) == 0 && pgConnstr == "" {
		log.Debug().Msgf("No log file or folder specified, using default log directory %s ...", DEFAULT_LOG_DIR)
		return FindMostRecentFileInFolder(DEFAULT_LOG_DIR)
	}

	if len(fileOrFolder) > 0 {
		_, err := os.Stat(fileOrFolder[0])
		if !os.IsNotExist(err) {
			logFile = fileOrFolder[0]
		}
		logFolder = filepath.Dir(logFile)
		return logFile, logFolder, nil
	}

	if pgConnstr != "" {
		log.Debug().Msg("Using --connstr to determine logs location ...")
		logSettings, err := postgres.GetLogSettings(pgConnstr)
		if err != nil {
			log.Fatal().Err(err).Msg("Error getting log directory and prefix from DB")
		}
		log.Debug().Msgf("Postgres logSettings: %v", logSettings)
		if logSettings.LogDestination == "stderr" {
			logFolder = DEFAULT_LOG_DIR // TODO a list of few other common standard locations
		}
	} else {
		logFolder = DEFAULT_LOG_DIR
	}
	return FindMostRecentFileInFolder(logFolder)
}
