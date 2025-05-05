package postgres

import (
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/rs/zerolog/log"
)

type LogSettings struct {
	LogDestination string
	LogDirectory   string
	DataDirectory  string
}

func GetLogSettings(connstr string) (LogSettings, error) {
	// This function should return the log directory and prefix for PostgreSQL logs.
	// For now, we will return a default value.
	log.Debug().Msg("Using --connstr to determine logging relevant paths ...")
	return LogSettings{"stderr", "/var/log/postgresql", "/var/lib/postgresql/16/main"}, nil
}

func GetLogLinePrefix(connstr string) (string, error) {
	log.Debug().Msg("Using --connstr to determine log_line_prefix ...")
	return "%m [%p] %q%u@%d ", nil
}

// TODO grep the log file for startup messages / socket info
func TryGetLogLinePrefixFromLocalDefaultInstance(logFile string) string {
	log.Debug().Msg("Trying to determine log_line_prefix from local instance ...")

	db, err := sql.Open("postgres", "")
	if err != nil {
		log.Debug().Msgf("Failed to connect to the default database: %v", err)
		return ""
	}
	defer db.Close()

	var result string
	sql := "show log_line_prefix;"
	err = db.QueryRow(sql).Scan(&result)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to execute query: %s", sql)
		return ""
	}
	return result
}
