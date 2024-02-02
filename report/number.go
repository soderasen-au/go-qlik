package report

import (
	"fmt"
	"github.com/soderasen-au/go-common/util"
	"strconv"
	"strings"
	"time"
)

func ParseExcelDateTime(serialNumber float64) time.Time {
	excelEpoch := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	return excelEpoch.Add(time.Duration(serialNumber * float64(24*time.Hour)))
}

func GetDateFormat(dateFmt string) string {
	goFmt := dateFmt
	goFmt = strings.ReplaceAll(goFmt, "YYYY", "2006")
	goFmt = strings.ReplaceAll(goFmt, "YY", "06")
	goFmt = strings.ReplaceAll(goFmt, "MMMM", "January")
	goFmt = strings.ReplaceAll(goFmt, "MMM", "Jan")
	goFmt = strings.ReplaceAll(goFmt, "MM", "01")
	goFmt = strings.ReplaceAll(goFmt, "M", "1")
	goFmt = strings.ReplaceAll(goFmt, "WWWW", "Monday")
	goFmt = strings.ReplaceAll(goFmt, "WWW", "Mon")
	goFmt = strings.ReplaceAll(goFmt, "W", "Mon")
	goFmt = strings.ReplaceAll(goFmt, "DD", "02")
	goFmt = strings.ReplaceAll(goFmt, "D", "2")

	if strings.Contains(goFmt, "tt") {
		goFmt = strings.ReplaceAll(goFmt, "hh", "03")
		goFmt = strings.ReplaceAll(goFmt, "tt", "PM")
	} else {
		goFmt = strings.ReplaceAll(goFmt, "hh", "15")
	}
	goFmt = strings.ReplaceAll(goFmt, "mm", "04")
	goFmt = strings.ReplaceAll(goFmt, "ss", "05")
	goFmt = strings.ReplaceAll(goFmt, "f", "0")

	return goFmt
}
func FormatDate(num float64, dateFmt string) string {
	return ParseExcelDateTime(num).Format(GetDateFormat(dateFmt))
}

func FormatNum(num float64, fmtStr string) (string, *util.Result) {
	if fmtStr[0] != '#' {
		return "", util.MsgError("fmtStr", "not leading with '#'")
	}
	intPos := strings.Index(fmtStr, "##0")
	if intPos < 0 {
		return "", util.MsgError("fmtStr", "no integer descriptor")
	}
	thousandSep := fmtStr[1:intPos]

	decimalPoint := "."
	decPart := fmtStr[intPos+3:]
	fractionPrecision := len(decPart) - 1

	numString := strconv.FormatFloat(num, 'f', fractionPrecision, 64)
	parts := strings.Split(numString, ".")
	intPart := parts[0]
	intParts := make([]string, len(intPart)/3+1)
	for i := len(intParts) - 1; i > 0; i-- {
		intParts[i] = intPart[len(intPart)-3:]
		intPart = intPart[:len(intPart)-3]
	}
	intParts[0] = intPart

	if len(decPart) > 0 {
		decimalPoint = decPart[0:1]
		return fmt.Sprintf("%s%s%s", strings.Join(intParts, thousandSep), decimalPoint, parts[1]), nil
	} else {
		return strings.Join(intParts, thousandSep), nil
	}
}
