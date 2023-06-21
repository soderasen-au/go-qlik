package qrs

import (
	"encoding/csv"
	"io"
	"os"
	"time"

	"github.com/Click-CI/common/util"
	"github.com/gocarina/gocsv"
)

type (
	TraceLogTimeStamp struct {
		time.Time
	}

	TraceLogEntry struct {
		Timestamp           TraceLogTimeStamp `csv:"Timestamp"`
		Level               string            `csv:"Level"`
		Hostname            string            `csv:"Hostname"`
		Logger              string            `csv:"Logger"`
		Thread              string            `csv:"Thread"`
		Id                  string            `csv:"Id"`
		ServiceUser         string            `csv:"ServiceUser"`
		Message             string            `csv:"Message"`
		ProxySessionId      string            `csv:"ProxySessionId"`
		Action              string            `csv:"Action"`
		ActiveUserDirectory string            `csv:"ActiveUserDirectory"`
		ActiveUserId        string            `csv:"ActiveUserId"`
		ResourceId          string            `csv:"ResourceId"`
		Checksum            string            `csv:"Checksum"`
	}
)

func (date *TraceLogTimeStamp) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.Parse("20060102T150405.000-0700", csv)
	return err
}

func ReadTraceLogFile(file string) ([]TraceLogEntry, *util.Result) {
	ret := make([]TraceLogEntry, 0)

	logFile, err := os.Open(file)
	if err != nil {
		return nil, util.Error("OpenFile", err)
	}
	defer logFile.Close()

	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.LazyQuotes = true
		return r
	})

	if err = gocsv.UnmarshalFile(logFile, &ret); err != nil {
		return nil, util.Error("UnmarshalCsvFile", err)
	}
	return ret, nil
}
