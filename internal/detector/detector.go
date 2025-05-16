package detector

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/kmoppel/pgweasel/internal/postgres"
	"github.com/rs/zerolog/log"
)

// Functions to detect the location of Postgres log files automatically

// func ScanForPostgresInstances() ([]string, error) {
// }

var DEFAULT_LOG_LOCATIONS = []string{"/var/log/postgresql", "/var/lib/pgsql"}

const DEFAULT_LOGFILE_SUFFIX = ".log"
const DEFAULT_LOG_LINE_PREFIX = "%m [%p] %q%u@%d "

// Returns the most recent file in the specified folder, plus it's parent folder
func FindMostRecentFileInFolders(foldersToScan []string) (latestFile string, latestFileParentFolder string, err error) {
	// var latestFile string
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

func DiscoverLatestLogFileAndFolder(fileOrFolder []string, pgConnstr string) (string, string, error) {
	var logFile string

	if len(fileOrFolder) == 0 && pgConnstr == "" {
		log.Debug().Msgf("No log file or folder specified, using default log directory %s ...", DEFAULT_LOG_LOCATIONS)
		return FindMostRecentFileInFolders(DEFAULT_LOG_LOCATIONS)
	}

	if len(fileOrFolder) > 0 {
		_, err := os.Stat(fileOrFolder[0])
		if !os.IsNotExist(err) {
			logFile = fileOrFolder[0]
		}
		return logFile, filepath.Dir(logFile), nil
	}

	if pgConnstr != "" {
		log.Debug().Msg("Using --connstr to determine logs location ...")
		logSettings, err := postgres.GetLogSettings(pgConnstr)
		if err != nil {
			log.Fatal().Err(err).Msg("Error getting log directory and prefix from DB")
		}
		log.Debug().Msgf("Postgres logSettings: %v", logSettings)
		if logSettings.LogDestination != "stderr" {
			return FindMostRecentFileInFolders([]string{path.Join(logSettings.DataDirectory, logSettings.LogDirectory)})
		}
	}
	return FindMostRecentFileInFolders(DEFAULT_LOG_LOCATIONS)
}
