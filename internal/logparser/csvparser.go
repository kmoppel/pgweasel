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
	ch := make(chan pglog.LogEntry)
	go func() {
		defer close(ch)
		file, err := os.Open(filePath)
		if err != nil {
			log.Error().Err(err).Msgf("Error opening file: %s", filePath)
			return
		}
		defer file.Close()

		var reader io.Reader = file

		if strings.HasSuffix(filePath, ".gz") {
			gzReader, err := gzip.NewReader(file)
			if err != nil {
				log.Error().Err(err).Msgf("Error creating gzip reader for file: %s", filePath)
				return
			}
			defer gzReader.Close()
			reader = gzReader
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

			if len(record) < 24 {
				fmt.Println("Skipping incomplete record")
				continue
			}

			// e := pglog.CsvEntry{
			// 	LogTime:              record[0],
			// 	UserName:             record[1],
			// 	DatabaseName:         record[2],
			// 	ProcessID:            record[3],
			// 	ConnectionFrom:       record[4],
			// 	SessionID:            record[5],
			// 	SessionLineNum:       record[6],
			// 	CommandTag:           record[7],
			// 	SessionStartTime:     record[8],
			// 	VirtualTransactionID: record[9],
			// 	TransactionID:        record[10],
			// 	ErrorSeverity:        record[11],
			// 	SQLStateCode:         record[12],
			// 	Message:              record[13],
			// 	Detail:               record[14],
			// 	Hint:                 record[15],
			// 	InternalQuery:        record[16],
			// 	InternalQueryPos:     record[17],
			// 	Context:              record[18],
			// 	Query:                record[19],
			// 	QueryPos:             record[20],
			// 	Location:             record[21],
			// 	ApplicationName:      record[22],
			// 	BackendType:          record[23],
			// }
			e := pglog.LogEntry{
				LogTime:       record[0],
				ErrorSeverity: record[11],
				Message:       record[13],
				Lines:         record,
			}

			e.Lines = record
			// log.Debug().Msgf("Parsed CSV entry: %+v", e)
			ch <- e

		}
	}()
	return ch
}
