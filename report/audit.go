package report

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/soderasen-au/go-common/util"
)

type AuditRecord struct {
	Timestamp       time.Time
	UserDir         string
	UserId          string
	IpAddr          string
	AppId           string
	Cmd             string
	Ids             []string
	ReportFileName  string
	ReportFileSize  int
	ReportTotalRows int
}

func (f AuditRecord) GetCSVLine() []string {
	return []string{
		f.Timestamp.Format(time.RFC3339),
		f.UserDir,
		f.UserId,
		f.IpAddr,
		f.Cmd,
		fmt.Sprintf("%v", f.Ids),
		f.ReportFileName,
		fmt.Sprintf("%d", f.ReportFileSize),
		fmt.Sprintf("%d", f.ReportTotalRows),
	}
}
func GetCSVHeader() string {
	return fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%v,%v\n", "Timestamp", "UserDir", "UserId", "IpAddr", "AppId", "Cmd", "Ids", "FileName", "FileSize", "TotalRows")
}

type AuditLog struct {
	fileName string
	fd       *os.File
	writer   *csv.Writer
	mu       sync.Mutex
}

func (audit *AuditLog) Close() {
	if audit.fd != nil {
		audit.fd.Close()
	}
}

func (audit *AuditLog) Record(r AuditRecord) *util.Result {
	audit.mu.Lock()
	defer audit.mu.Unlock()

	err := audit.writer.Write(r.GetCSVLine())
	if err != nil {
		return util.Error("WriteRecord", err)
	}
	return nil
}

func (audit *AuditLog) OpenFile(fn string) *util.Result {
	f, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return util.Error("OpenFile", err)
	}

	f.Seek(0, 0)
	scanner := bufio.NewScanner(audit.fd)
	ok := scanner.Scan()
	if err := scanner.Err(); !ok && err != nil {
		return util.Error("Scan", err)
	}
	firstLine := strings.TrimSpace(scanner.Text())
	if firstLine == "" {
		f.Seek(0, 0)
		_, err := f.WriteString(GetCSVHeader())
		if err != nil {
			return util.Error("WriteHeader", err)
		}
	}
	f.Seek(0, 2) //to the end of file

	audit.fileName = fn
	audit.fd = f
	audit.writer = csv.NewWriter(audit.fd)

	return nil
}

func NewAuditLog(fn string) (*AuditLog, *util.Result) {
	auditLog := &AuditLog{}
	res := auditLog.OpenFile(fn)
	if res != nil {
		return nil, res.With("OpenFile")
	}

	return auditLog, nil
}
