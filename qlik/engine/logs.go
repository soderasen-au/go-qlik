package engine

import (
	"encoding/csv"
	"io"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/soderasen-au/go-common/util"
)

type (
	AuditLogTimeStamp struct {
		time.Time
	}

	AuditLogEntry struct {
		ProductVersion    string            `csv:"ProductVersion"`
		Timestamp         AuditLogTimeStamp `csv:"Timestamp"`
		Hostname          string            `csv:"Hostname"`
		Id                string            `csv:"Id"`
		EngineTimestamp   AuditLogTimeStamp `csv:"EngineTimestamp"`
		EngineVersion     string            `csv:"EngineVersion"`
		Description       string            `csv:"Description"`
		ProxySessionId    string            `csv:"ProxySessionId"`
		ProxyPackageId    string            `csv:"ProxyPackageId"`
		RequestSequenceId string            `csv:"RequestSequenceId"`
		UserDirectory     string            `csv:"UserDirectory"`
		UserId            string            `csv:"UserId"`
		SessionId         string            `csv:"SessionId"`
		ObjectId          string            `csv:"ObjectId"`
		ObjectName        string            `csv:"ObjectName"`
		Service           string            `csv:"Service"`
		Origin            string            `csv:"Origin"`
		Context           string            `csv:"Context"`
		Command           string            `csv:"Command"`
		Result            string            `csv:"Result"`
		Message           string            `csv:"Message"`
		Id2               string            `csv:"Id2"`
	}
)

func (date *AuditLogTimeStamp) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.Parse("20060102T150405.000-0700", csv)
	return err
}

func ReadAuditLogFile(file string) ([]AuditLogEntry, *util.Result) {
	ret := make([]AuditLogEntry, 0)

	logFile, err := os.Open(file)
	if err != nil {
		return nil, util.Error("OpenFile", err)
	}
	defer logFile.Close()

	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		return r
	})

	if err = gocsv.UnmarshalFile(logFile, &ret); err != nil {
		return nil, util.Error("UnmarshalCsvFile", err)
	}
	return ret, nil
}
