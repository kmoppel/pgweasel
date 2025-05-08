package logparser

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/kmoppel/pgweasel/internal/pglog"
	"github.com/rs/zerolog/log"
)

type PgLogCsvEntry struct {
	LogTime              string // Column 1
	UserName             string // Column 2
	DatabaseName         string // Column 3
	ProcessID            string // Column 4
	ConnectionFrom       string // Column 5
	SessionID            string // Column 6
	SessionLineNum       string // Column 7
	CommandTag           string // Column 8
	SessionStartTime     string // Column 9
	VirtualTransactionID string // Column 10
	TransactionID        string // Column 11
	ErrorSeverity        string // Column 12
	SQLStateCode         string // Column 13
	Message              string // Column 14
	Detail               string // Column 15
	Hint                 string // Column 16
	InternalQuery        string // Column 17
	InternalQueryPos     string // Column 18
	Context              string // Column 19
	Query                string // Column 20
	QueryPos             string // Column 21
	Location             string // Column 22
	ApplicationName      string // Column 23
	BackendType          string // Column 24
	// (Columns beyond this are optional or reserved)
}

func ShowErrorsCsv(filePath string, minLvl string, extraFilters []string) error {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	r := csv.NewReader(file)
	r.FieldsPerRecord = -1 // Allow variable fields

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading record:", err)
			continue
		}

		if len(record) < 24 {
			fmt.Println("Skipping incomplete record")
			continue
		}

		e := PgLogCsvEntry{
			LogTime:              record[0],
			UserName:             record[1],
			DatabaseName:         record[2],
			ProcessID:            record[3],
			ConnectionFrom:       record[4],
			SessionID:            record[5],
			SessionLineNum:       record[6],
			CommandTag:           record[7],
			SessionStartTime:     record[8],
			VirtualTransactionID: record[9],
			TransactionID:        record[10],
			ErrorSeverity:        record[11],
			SQLStateCode:         record[12],
			Message:              record[13],
			Detail:               record[14],
			Hint:                 record[15],
			InternalQuery:        record[16],
			InternalQueryPos:     record[17],
			Context:              record[18],
			Query:                record[19],
			QueryPos:             record[20],
			Location:             record[21],
			ApplicationName:      record[22],
			BackendType:          record[23],
		}

		fullLine := strings.Join(record, ",")
		log.Debug().Msgf("Parsed CSV entry: %+v", e)

		userFiltersSatisfied := 0
		if len(extraFilters) > 0 {
			for _, userFilter := range extraFilters {
				m, err := regexp.MatchString(userFilter, fullLine) // TODO compile and cache the regex
				if err != nil {
					log.Fatal().Err(err).Msgf("Error matching user provided filter %s on line: %s", userFilter, fullLine)
					continue
				}
				if m {
					userFiltersSatisfied += 1
					break
				}
			}
		}

		if pglog.SeverityToNum(e.ErrorSeverity) >= pglog.SeverityToNum(minLvl) && userFiltersSatisfied == len(extraFilters) {
			fmt.Println(fullLine)
		}
	}
	return nil
}
