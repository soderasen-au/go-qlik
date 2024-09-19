package report

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/soderasen-au/go-common/util"
)

func TestFormatNum(t *testing.T) {
	type args struct {
		num    float64
		fmtStr string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 *util.Result
	}{
		{"OnlyInteger", args{12345.6789, "###0"}, "12345", nil},
		{"IntWithThousand", args{12345.6789, "#,##0"}, "12,345", nil},
		{"WithThouAndDec", args{12345.6789, "#.##0,00"}, "12.345,68", nil},
		{"Normal4Decimals", args{12345.678910111213, "#,##0.0000"}, "12,345.6789", nil},
		{"Normal6Decimals", args{12345.678910111213, "#,##0.000000"}, "12,345.678910", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := FormatNum(tt.args.num, tt.args.fmtStr)
			fmt.Println("got: ", got)
			if got != tt.want {
				t.Errorf("FormatNum() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("FormatNum() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	type args struct {
		num     float64
		dateFmt string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"0", args{45223.71234567, "YYYY-MM-DD WWW hh:mm:ss.fff tt"}, "2023-10-24 Tue 05:05:46.665 PM"},
		{"00", args{45223.71234567, "YYYY-MM-DD WWW hh:mm:ss.fff"}, "2023-10-24 Tue 17:05:46.665"},
		{"1", args{45223.71234567, "YYYYMMDD"}, "20231024"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatDate(tt.args.num, tt.args.dateFmt); got != tt.want {
				t.Errorf("FormatDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
