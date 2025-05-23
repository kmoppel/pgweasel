package logparser

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kmoppel/pgweasel/internal/pglog"
	"github.com/rs/zerolog/log"
)

func GetLogRecordsFromCsvFile(filePath string) <-chan pglog.LogEntry {
	log.Debug().Msgf("Looking for log entries from CSV file: %s", filePath)
	ch := make(chan pglog.LogEntry)

	go func() {
		defer close(ch)

		var reader io.Reader

		if filePath == "stdin" {
			reader = os.Stdin
		} else {
			file, err := os.Open(filePath)
			if err != nil {
				log.Error().Err(err).Msgf("Error opening file: %s", filePath)
				return
			}
			defer file.Close()

			reader = file

			if strings.HasSuffix(filePath, ".gz") {
				gzReader, err := gzip.NewReader(file)
				if err != nil {
					log.Error().Err(err).Msgf("Error creating gzip reader for file: %s", filePath)
					return
				}
				defer gzReader.Close()
				reader = gzReader
			}

		}

		r := csv.NewReader(reader)
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

			if len(record) < 23 {
				fmt.Println("Skipping incomplete record")
				continue
			}

			e := pglog.LogEntry{
				LogTime:       record[0],
				ErrorSeverity: record[11],
				Message:       record[13],
				CsvColumns: &pglog.CsvEntry{ // Field order from https://www.postgresql.org/docs/current/file-fdw.html
					CsvColumnCount:       len(record),
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
				},
			}
			//   v13 added backend_type
			//   v14 added leader_pid and query_id
			if len(record) >= 24 {
				e.CsvColumns.BackendType = record[23]
			}
			if len(record) >= 26 {
				e.CsvColumns.LeaderPid = record[24]
				e.CsvColumns.QueryId = record[25]
			}
			ch <- e

		}
	}()
	return ch
}
