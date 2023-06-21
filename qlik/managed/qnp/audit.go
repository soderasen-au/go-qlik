package qnp

import (
	"bytes"
	"encoding/csv"
	"io"
	"time"

	"github.com/Click-CI/common/util"
	"github.com/gocarina/gocsv"
)

type (
	AuditLogTimeStamp struct {
		time.Time
	}

	AuditLogRecord struct {
		EventTime   AuditLogTimeStamp
		Source      string
		Event       string
		Target      string
		TargetId    string
		TargetName  string
		Action      string
		AppId       string
		UserId      string
		UserName    string
		IpAddress   string
		Description string
		Data        string
		DataType    string
	}

	NewsstandReportData struct {
		Title                  string `json:"title,omitempty"`
		Format                 string `json:"format,omitempty"`
		TaskId                 string `json:"taskId,omitempty"`
		FileSize               int64  `json:"fileSize,omitempty"`
		Published              string `json:"published,omitempty"`
		ExecutionId            string `json:"executionId,omitempty"`
		RecipientId            string `json:"recipientId,omitempty"`
		RecipientDomainAccount string `json:"recipientDomainAccount,omitempty"`
	}
)

func (date *AuditLogTimeStamp) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.Parse(time.RFC3339Nano, csv)
	return err
}

func (c *Client) GetAuditLogs() ([]byte, *util.Result) {
	data, res := c.Get("/audit/logs", nil)
	if res != nil {
		return nil, res.With("Get")
	}

	return data, nil
}

func (c *Client) GetAuditLogRecords() ([]AuditLogRecord, *util.Result) {
	data, res := c.GetAuditLogs()
	if res != nil {
		return nil, res.With("GetAuditLogs")
	}
	data = bytes.TrimLeft(data, "\xef\xbb\xbf")

	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.LazyQuotes = true
		return r
	})

	var auditRecords []AuditLogRecord
	err := gocsv.UnmarshalBytes(data, &auditRecords)
	if err != nil {
		return nil, util.Error("ParseAuditLog", err)
	}

	return auditRecords, nil
}
