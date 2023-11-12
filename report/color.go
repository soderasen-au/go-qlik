package report

import (
	"fmt"
	"github.com/soderasen-au/go-qlik/qlik/engine"
	"strconv"
	"strings"

	"github.com/qlik-oss/enigma-go/v4"
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/util"
	"github.com/xuri/excelize/v2"
)

type ARGBColor struct {
	A int
	R int
	G int
	B int
}

var (
	QlikPredefinedColorMap = map[string]*ARGBColor{
		"yellow":       {R: 255, G: 255, B: 0},
		"white":        {R: 255, G: 255, B: 255},
		"red":          {R: 128, G: 0, B: 0},
		"magenta":      {R: 128, G: 0, B: 128},
		"lightred":     {R: 255, G: 0, B: 0},
		"lightmagenta": {R: 255, G: 0, B: 255},
		"lightgreen":   {R: 0, G: 255, B: 0},
		"lightgray":    {R: 192, G: 192, B: 192},
		"lightcyan":    {R: 0, G: 255, B: 255},
		"lightblue":    {R: 0, G: 0, B: 255},
		"green":        {R: 0, G: 238, B: 0},
		"darkgray":     {R: 128, G: 128, B: 128},
		"cyan":         {R: 0, G: 128, B: 128},
		"brown":        {R: 128, G: 128, B: 0},
		"blue":         {R: 0, G: 0, B: 128},
		"black ":       {R: 0, G: 0, B: 0},
	}
)

func (c ARGBColor) AssignBgStyle(excelStyle *excelize.Style) {
	code := fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
	excelStyle.Fill.Type = "pattern"
	excelStyle.Fill.Pattern = 1
	excelStyle.Fill.Color = make([]string, 1)
	excelStyle.Fill.Color[0] = code

	luminance := (0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)) / 255.0
	d := 255
	if luminance > 0.5 {
		d = 0
	}
	textCode := fmt.Sprintf("#%02X%02X%02X", d, d, d)
	excelStyle.Font = &excelize.Font{Color: textCode}
}

func (c ARGBColor) AssignLuminanceFont(excelStyle *excelize.Style) {
	luminance := (0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)) / 255.0
	d := 255
	if luminance > 0.5 {
		d = 0
	}
	textCode := fmt.Sprintf("#%02X%02X%02X", d, d, d)
	excelStyle.Font = &excelize.Font{Color: textCode}
}

func (c ARGBColor) AssignFontStyle(excelStyle *excelize.Style) {
	textCode := fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
	excelStyle.Font = &excelize.Font{Color: textCode}
}

func NewARGBFromQlikColor(t string) (*ARGBColor, *util.Result) {
	t = strings.ToUpper(t)
	t = strings.ReplaceAll(t, " ", "")

	var ret *ARGBColor
	if strings.HasPrefix(t, "ARGB") {
		argb := t[4:]
		argb = argb[:len(argb)-1]
		cv := strings.Split(argb, ",")
		if len(cv) != 4 {
			return nil, util.MsgError("ParseColor", "invalid color sections")
		}
		r, err := strconv.Atoi(cv[1])
		if err != nil {
			return nil, util.MsgError("ParseColor", "invalid red color code")
		}
		g, err := strconv.Atoi(cv[2])
		if err != nil {
			return nil, util.MsgError("ParseColor", "invalid green color code")
		}
		b, err := strconv.Atoi(cv[3])
		if err != nil {
			return nil, util.MsgError("ParseColor", "invalid blue color code")
		}
		ret = &ARGBColor{
			R: r,
			G: g,
			B: b,
		}
	} else if strings.HasPrefix(t, "RGB") {
		argb := t[4:]
		argb = argb[:len(argb)-1]
		cv := strings.Split(argb, ",")
		if len(cv) != 3 {
			return nil, util.MsgError("ParseColor", "invalid RGB color sections")
		}
		r, err := strconv.Atoi(cv[0])
		if err != nil {
			return nil, util.MsgError("ParseColor", "invalid red color code")
		}
		g, err := strconv.Atoi(cv[1])
		if err != nil {
			return nil, util.MsgError("ParseColor", "invalid green color code")
		}
		b, err := strconv.Atoi(cv[2])
		if err != nil {
			return nil, util.MsgError("ParseColor", "invalid blue color code")
		}
		ret = &ARGBColor{
			R: r,
			G: g,
			B: b,
		}
	} else {
		colorText := strings.ToLower(t)
		cv := strings.Split(colorText, "(")
		colorCode := cv[0]
		if argb, ok := QlikPredefinedColorMap[colorCode]; ok {
			ret = argb
		}
	}

	return ret, nil
}

func NewARGBColorFromQlikAttr(attr *enigma.NxSimpleValue) (*ARGBColor, *util.Result) {
	if attr == nil {
		return nil, util.MsgError("NewARGBColorFromQlikAttr", "nil attr")
	}

	return NewARGBFromQlikColor(attr.Text)
}

func GetCellColorFont(attrs []*enigma.NxSimpleValue, excelStyle *excelize.Style, cellLogger *zerolog.Logger) *util.Result {
	if excelStyle == nil {
		excelStyle = &excelize.Style{}
	}

	bgAttr := attrs[0]
	cellLogger.Trace().Msgf("bg attr: %s", bgAttr.Text)
	bgColor, res := NewARGBColorFromQlikAttr(bgAttr)
	if res != nil {
		return res.With("NewARGBColorFromQlikAttr(bgAttr)")
	}
	if bgColor != nil {
		bgColor.AssignBgStyle(excelStyle)

		if len(attrs) > 1 {
			fontAttr := attrs[1]
			cellLogger.Trace().Msgf("font attr: %s", fontAttr.Text)
			fontColor, res := NewARGBColorFromQlikAttr(fontAttr)
			if res != nil {
				return res.With("NewARGBColorFromQlikAttr(fontAttr)")
			}
			if fontColor != nil {
				fontColor.AssignFontStyle(excelStyle)
			}
		} else {
			bgColor.AssignLuminanceFont(excelStyle)
		}
	}

	return nil
}

func (f ColumnHeaderFormat) GetHeaderCellStyle(colInfo *engine.ColumnInfo, cellLogger *zerolog.Logger) (*excelize.Style, *util.Result) {
	excelStyle := &excelize.Style{}
	cellLogger.Debug().Msgf("bg color: %s", f.BgColor)
	bgColor, res := NewARGBFromQlikColor(f.BgColor)
	if res != nil {
		return nil, res.LogWith(cellLogger, "NewARGBFrom(BgColor)")
	}
	if bgColor != nil {
		bgColor.AssignBgStyle(excelStyle)
	}

	cellLogger.Debug().Msgf("fg color: %s", f.FgColor)
	fgColor, res := NewARGBFromQlikColor(f.FgColor)
	if res != nil {
		return nil, res.LogWith(cellLogger, "NewARGBFrom(FgColor)")
	}
	if fgColor != nil {
		fgColor.AssignFontStyle(excelStyle)
	} else if bgColor != nil {
		bgColor.AssignLuminanceFont(excelStyle)
	}

	if f.DateFmt != "" {
		cellLogger.Debug().Msgf("num fmt: %s", f.DateFmt)
		excelStyle.CustomNumFmt = &f.DateFmt
	} else if f.NumFmt != "" {
		cellLogger.Debug().Msgf("num fmt: %s", f.NumFmt)
		excelStyle.CustomNumFmt = &f.NumFmt
	}
	cellLogger.Debug().Msgf("excel custom num fmt: %s", util.MaybeNil(excelStyle.CustomNumFmt))

	return excelStyle, nil
}

func GetStackCellStyle(cell *enigma.NxCell, cellLogger *zerolog.Logger) (*excelize.Style, *util.Result) {
	excelStyle := excelize.Style{}
	if cell.AttrExps != nil && cell.AttrExps.Values != nil && len(cell.AttrExps.Values) > 0 {
		res := GetCellColorFont(cell.AttrExps.Values, &excelStyle, cellLogger)
		if res != nil {
			return nil, res.With("GetCellColorFont")
		}
	}
	return &excelStyle, nil
}

func GetPivotCellStyle(cell *enigma.NxPivotValuePoint, cellLogger *zerolog.Logger) (*excelize.Style, *util.Result) {
	excelStyle := excelize.Style{}
	if cell.AttrExps != nil && cell.AttrExps.Values != nil && len(cell.AttrExps.Values) > 0 {
		res := GetCellColorFont(cell.AttrExps.Values, &excelStyle, cellLogger)
		if res != nil {
			return nil, res.With("GetCellColorFont")
		}
	}
	if cell.Type == "T" {
		if excelStyle.Font == nil {
			excelStyle.Font = &excelize.Font{}
		}
		excelStyle.Font.Bold = true
	}
	return &excelStyle, nil
}
