package report

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/soderasen-au/go-common/util"

	"github.com/soderasen-au/go-qlik/qlik/engine"
)

type CsvReportPrinter struct {
	ReportPrinterBase
	ObjId          string
	ObjLayout      *engine.ObjectLayoutEx
	Writer         *csv.Writer
	ColCnt         int
	RowCnt         int
	Cube2report    map[int]int
	Cube2CustomFmt map[int]*ColumnHeaderFormat
}

func NewCsvReportPrinter() *CsvReportPrinter {
	p := &CsvReportPrinter{}
	p.ReportResults = make(map[string]*ReportResult)
	return p
}

func (p *CsvReportPrinter) printObjectHeader() *util.Result {
	if p.ObjLayout == nil {
		return util.MsgError("printObjectHeader", "nil p.ObjLayout")
	}
	if p.Writer == nil {
		return util.MsgError("printObjectHeader", "nil Writer")
	}

	logger := p.Logger.With().Str("print", "header").Logger()
	DimCnt := len(p.ObjLayout.HyperCube.DimensionInfo)

	ColumnOrder := p.ObjLayout.HyperCube.ColumnOrder
	if ColumnOrder == nil || len(ColumnOrder) == 0 {
		ColumnOrder = p.ObjLayout.HyperCube.EffectiveInterColumnSortOrder
	}

	var colInfo *engine.ColumnInfo
	p.ObjLayout.ColumnInfos = make([]*engine.ColumnInfo, 0)
	record := make([]string, len(ColumnOrder))
	p.Cube2report = make(map[int]int) // cube column index => report column index
	p.Cube2CustomFmt = make(map[int]*ColumnHeaderFormat)
	for ci, colIx := range ColumnOrder {
		expIx := colIx - DimCnt
		if colIx < DimCnt {
			dim := p.ObjLayout.HyperCube.DimensionInfo[colIx]
			if dim.Error != nil {
				logger.Warn().Msgf("dim[%d] %s has error: (%d) [%s] %s, ignore.", colIx, dim.FallbackTitle, dim.Error.ErrorCode, dim.Error.Context, dim.Error.ExtendedMessage)
				continue
			}
			colInfo = engine.NewColumnInfoFromDimension(dim)
		} else {
			exp := p.ObjLayout.HyperCube.MeasureInfo[expIx]
			if exp.Error != nil {
				logger.Warn().Msgf("exp[%d] %s has error: (%d) [%s] %s, ignore.", expIx, exp.FallbackTitle, exp.Error.ErrorCode, exp.Error.Context, exp.Error.ExtendedMessage)
				continue
			}
			colInfo = engine.NewColumnInfoFromMeasure(exp)
		}
		p.ObjLayout.ColumnInfos = append(p.ObjLayout.ColumnInfos, colInfo)
		cellText := colInfo.FallbackTitle

		p.Cube2report[p.ColCnt] = p.ColCnt
		if p.R.ColumnHeaderFormats != nil {
			if colHeaderFmt, ok := p.R.ColumnHeaderFormats[cellText]; ok {
				logger.Info().Msgf(" - sense[%d]:%s => report[%d]", ci, cellText, colHeaderFmt.Order)
				p.Cube2report[p.ColCnt] = colHeaderFmt.Order
				p.Cube2CustomFmt[p.ColCnt] = &colHeaderFmt
				if colHeaderFmt.Label != "" {
					cellText = colHeaderFmt.Label
				}
			}
		}

		repIdx := p.Cube2report[p.ColCnt]
		logger.Info().Msgf("header col: %d => %d: %s", ci, colIx, cellText)
		record[repIdx] = cellText
		p.ColCnt++
	}

	err := p.Writer.Write(record)
	if err != nil {
		logger.Err(err).Msg("write header")
		return util.Error("WriteCSV", err)
	}
	p.Writer.Flush()
	if p.ColCnt != p.ObjLayout.HyperCube.Size.Cx && p.ObjLayout.HyperCube.Mode == "S" {
		logger.Warn().Msgf("Col(%d) != hypercube.x(%d)", p.ColCnt, p.ObjLayout.HyperCube.Size.Cx)
	}

	return nil
}

// rect [in] rect.Top, rect.Left set the start offset posistion of the table;
// rect* [out] rect.Top, rect.Left, rect.Width, rect.Height to indicate result table area;
func (p *CsvReportPrinter) printStackObject() *util.Result {
	logger := p.Logger.With().Str("Stack", p.ObjId).Logger()
	logger.Info().Msg("start to print")

	obj, err := p.Doc.GetObject(engine.ConnCtx, p.ObjId)
	if err != nil {
		logger.Err(err).Msg("GetObject failed")
		return util.Error("GetObject", err)
	}
	if obj.Handle == 0 {
		logger.Error().Msgf("can't get object %s, save your app properly and make sure object exists", p.ObjId)
		return util.MsgError("GetObject", fmt.Sprintf("can't get object %s, save your app properly and make sure object exists", p.ObjId))
	}
	logger.Info().Msgf("got object: %s/%s", obj.GenericType, obj.GenericId)

	if p.ObjLayout.HyperCube == nil {
		logger.Warn().Msgf("can't get hypercube for object: %s/%s, ignore", obj.GenericType, obj.GenericId)
		return nil
	}
	logger.Info().Msgf("Hypercube size: %d, %d", p.ObjLayout.HyperCube.Size.Cx, p.ObjLayout.HyperCube.Size.Cy)

	if p.ObjLayout.HyperCube.Error != nil {
		cubeErr := p.ObjLayout.HyperCube.Error
		if cubeErr.ErrorCode == 7005 {
			cubeErr.ExtendedMessage = "Evaluation condition is not met. Calculation is failed."
			cubeErr.Context = "HypercubeVerification"
		}
		errMsg := fmt.Sprintf("hypercube has error: code: %d, context: %s, message: %s", cubeErr.ErrorCode, cubeErr.Context, cubeErr.ExtendedMessage)
		logger.Warn().Msg(errMsg)
		return util.MsgError("CheckHyperCube", errMsg)
	}

	res := p.printObjectHeader()
	if res != nil {
		logger.Err(res).Msg("printObjectHeader failed")
		return res.With("printObjectHeader")
	}
	if p.ObjLayout.HyperCube.Size.Cx != p.ColCnt {
		logger.Warn().Msgf("printed data cells[%d] != header cells[%d]", p.ObjLayout.HyperCube.Size.Cx, p.ColCnt)
	}

	dataPages, res := engine.GetHyperCubeData(obj, *p.ObjLayout.HyperCube.Size, engine.PivotPaging)
	if res != nil {
		logger.Err(res).Msg("GetHyperCubeData failed")
		return res.With("GetHyperCubeData")
	}
	logger.Info().Msgf("Hypercube: %d", len(dataPages))

	for pi, page := range dataPages {
		if page.Area.Height < 1 {
			logger.Warn().Msgf("page: [%d] is empty, ignore ...", pi)
			continue
		}

		pageLogger := logger.With().Str(fmt.Sprintf("page[%d]", pi), fmt.Sprintf("(%d, %d)", p.RowCnt, p.ColCnt)).Logger()
		pageLogger.Debug().Msg("start")

		records := make([][]string, len(page.Matrix))
		for ri, row := range page.Matrix {
			records[ri] = make([]string, len(row))
			for cubeColIx, cell := range row {
				ci := p.Cube2report[cubeColIx]
				if colFmt, ok := p.Cube2CustomFmt[cubeColIx]; ok {
					if colFmt != nil {
						if colFmt.NumFmt != "" {
							txt, res := FormatNum(float64(cell.Num), colFmt.NumFmt)
							if res != nil {
								pageLogger.Error().Msgf("FormatNum: %s", res.Error())
								records[ri][ci] = cell.Text
							} else {
								records[ri][ci] = txt
							}
						} else if colFmt.DateFmt != "" {
							records[ri][ci] = FormatDate(float64(cell.Num), colFmt.DateFmt)
						} else {
							records[ri][ci] = cell.Text
						}
					} else {
						pageLogger.Warn().Msgf("p.Cube2CustomFmt[%d] is empty", cubeColIx)
						records[ri][ci] = cell.Text
					}
				} else {
					pageLogger.Trace().Msgf("p.Cube2CustomFmt has no cube idx: %d", cubeColIx)
					records[ri][ci] = cell.Text
				}
			}
			p.RowCnt++
		}
		err = p.Writer.WriteAll(records)
		if err != nil {
			pageLogger.Err(err).Msg("GetHyperCubeData")
			return res.With("GetHyperCubeData")
		}
		p.Writer.Flush()
		pageLogger.Debug().Msg("end")
	}

	reportResult, res := p.GetReportResult(*p.R.ID)
	if res != nil {
		return res.With("GetReportResult")
	}
	reportResult.PrintedRows = p.RowCnt

	logger.Info().Msgf("finish printing total rows: %d", reportResult.PrintedRows)
	return nil
}

// rect [in] rect.Top, rect.Left set the start offset posistion of the table;
// rect* [out] rect.Top, rect.Left, rect.Width, rect.Height to indicate result table area;
func (p *CsvReportPrinter) printObject() *util.Result {
	obj, err := p.Doc.GetObject(engine.ConnCtx, p.ObjId)
	if err != nil {
		p.Logger.Err(err).Msg("GetObject failed")
		return util.Error("GetObject", err)
	}
	if obj.Handle == 0 {
		return util.LogMsgError(p.Logger, "GetObject", fmt.Sprintf("can't get object %s, save your app properly and make sure object exists", p.ObjId))
	}
	p.Logger.Info().Msgf("got object: %s/%s", obj.GenericType, obj.GenericId)

	objLayout, res := engine.GetObjectLayoutEx(obj)
	if res != nil {
		return res.LogWith(p.Logger, "GetObjectLayoutEx")
	}

	if objLayout.Info.Type == "container" {
		return util.LogMsgError(p.Logger, "GetObjectType", fmt.Sprintf("can't print csv for objecct type `%s`", objLayout.Info.Type))
	}

	if objLayout.HyperCube != nil && objLayout.HyperCube.Mode == "P" {
		return util.LogMsgError(p.Logger, "GetObjectType", fmt.Sprintf("can't print csv for objecct type `%s`", "pivot"))
	}

	if objLayout.HyperCube != nil && objLayout.HyperCube.Mode == "K" {
		return util.LogMsgError(p.Logger, "GetObjectType", fmt.Sprintf("can't print csv for objecct type `%s`", "pivot_stack"))
	}

	p.ObjLayout = objLayout
	return p.printStackObject()
}

func (p *CsvReportPrinter) Print(r Report) *util.Result {
	if !r.IsValid() {
		return util.MsgError("Print", "invalid report")
	}

	if !r.OutputFormat.IsCsv() {
		return util.MsgError("OutputFormat", "CsvReportPrinter only support csv format")
	}

	rResult, res := NewReportResult(r)
	if res != nil {
		return res.With("NewReportResult")
	}
	p.ReportResults[util.MaybeNil(r.ID)] = rResult
	logger := rResult.Logger.With().Str("report", *r.ID).Logger()
	p.Logger = &logger
	p.R = r

	if r.Doc == nil {
		return util.MsgError("CheckDoc", "doc is not opened")
	}
	p.Doc = r.Doc

	ofs, err := os.OpenFile(util.MaybeNil(rResult.ReportFile), os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return util.Error("OpenFile: "+util.MaybeNil(rResult.ReportFile), err)
	}
	p.Writer = csv.NewWriter(ofs)
	if r.OutputFormat.IsTsv() {
		p.Writer.Comma = '\t'
	}
	defer func() {
		if ofs != nil {
			ofs.Close()
		}
	}()

	r.Target = strings.ToLower(r.Target)
	if r.Target != TARGET_OBJECTS {
		return util.LogMsgError(&logger, "CheckTarget", r.Target+" is not supported. Sense only supports objects")
	}
	if len(r.TargetIDs) != 1 {
		return util.LogMsgError(&logger, "CheckTarget", "Sense only supports report single object")
	}
	p.ObjId = r.TargetIDs[0]

	res = p.printObject()
	if res != nil {
		return res.With("printObject")
	}

	p.Writer.Flush()
	err = ofs.Close()
	if res != nil {
		return util.Error("Close", err)
	}
	ofs = nil

	logger.Info().Msgf("report is saved as [%s]", *rResult.ReportFile)

	return nil
}

func (p CsvReportPrinter) GetReportResult(id string) (*ReportResult, *util.Result) {
	result, ok := p.ReportResults[id]
	if !ok {
		return nil, util.MsgError("ReportFiles", "report id doesn't exists")
	}
	return result, nil
}
