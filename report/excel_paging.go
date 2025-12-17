package report

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/qlik-oss/enigma-go/v4"
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/util"
	"github.com/xuri/excelize/v2"

	"github.com/soderasen-au/go-qlik/qlik/engine"
)

// ExcelPagingConfig holds configuration for paginated Excel export
type ExcelPagingConfig struct {
	RowsPerPage       int    `json:"rows_per_page" yaml:"rows_per_page"`
	ReportTitle       string `json:"report_title" yaml:"report_title"`
	TotalRecordsLabel string `json:"total_records_label" yaml:"total_records_label"`
	ShowColumnNumbers bool   `json:"show_column_numbers" yaml:"show_column_numbers"`
	ShowSubtotals     bool   `json:"show_subtotals" yaml:"show_subtotals"`
}

// DefaultExcelPagingConfig returns default configuration
func DefaultExcelPagingConfig() ExcelPagingConfig {
	return ExcelPagingConfig{
		RowsPerPage:       50,
		ReportTitle:       "Paginated Report",
		TotalRecordsLabel: "Total Records Found",
		ShowColumnNumbers: false,
		ShowSubtotals:     false,
	}
}

// ExcelPagingPrinter exports Qlik objects to paginated Excel sheets
type ExcelPagingPrinter struct {
	ReportPrinterBase
	Config ExcelPagingConfig
}

// NewExcelPagingPrinter creates a new paginated Excel printer
func NewExcelPagingPrinter(config ExcelPagingConfig) *ExcelPagingPrinter {
	if config.RowsPerPage <= 0 {
		config.RowsPerPage = 50
	}
	if config.TotalRecordsLabel == "" {
		config.TotalRecordsLabel = "Total Records Found"
	}
	p := &ExcelPagingPrinter{Config: config}
	p.ReportResults = make(map[string]*ReportResult)
	return p
}

// printHorizontalSelection prints current selections in horizontal format
// Field names in one row, field values in the next row
func (p *ExcelPagingPrinter) printHorizontalSelection(r Report, doc *enigma.Doc, sheet string, excel *excelize.File, rect enigma.Rect, _logger *zerolog.Logger) (*enigma.Rect, *util.Result) {
	if excel == nil {
		return nil, util.MsgError("printHorizontalSelection", "nil excel")
	}
	logger := _logger.With().Str("print", "horizontalSelection").Logger()
	resRect := rect
	resRect.Height = 0
	resRect.Width = 0

	selObj, res := engine.GetCurrentSelection(doc, "$")
	if res != nil {
		return nil, res.With("GetCurrentSelection")
	}

	if len(selObj.Selections) == 0 {
		return &resRect, nil
	}

	// Get dimension label mapping
	dimFieldMap := make(map[string]string)
	dimLabelMap := make(map[string]string)
	dimList, res := engine.GetDimensionList(doc)
	if res == nil {
		for _, dimItem := range dimList {
			dimTitle := util.MaybeNil(dimItem.Meta.Title)
			dim := dimItem.Dim
			if len(dim.FieldDefs) < 1 {
				continue
			}
			dimDef := dim.FieldDefs[0]
			dimLabel := dim.LabelExpression
			if len(dim.FieldLabels) > 0 {
				dimLabel = dim.FieldLabels[0]
			}
			dimLabelMap[dimDef] = dimLabel
			dimFieldMap[dimDef] = dimTitle
		}
	}

	// Sort selections by order
	selections := selObj.Selections
	sort.SliceStable(selections, func(i, j int) bool {
		order1, ok := r.CurrentSelectionOrder[selections[i].Field]
		if !ok {
			return false
		}
		order2, ok := r.CurrentSelectionOrder[selections[j].Field]
		if !ok {
			return true
		}
		return order1 < order2
	})

	// Filter hidden fields and map names
	type selectionItem struct {
		name   string
		values string
	}
	items := make([]selectionItem, 0)

	for _, sel := range selections {
		fname := sel.Field
		isHidden := false
		listObj, res := engine.GetListObject(doc, "$", fname)
		if res == nil {
			for _, tag := range listObj.DimensionInfo.Tags {
				if tag == "$hidden" {
					isHidden = true
					break
				}
			}
		}
		if isHidden {
			continue
		}

		// Map field name
		if strings.HasPrefix(fname, "=") {
			if dname, ok := dimLabelMap[fname]; ok {
				fname = dname
			}
		} else if mappedName, ok := dimFieldMap[fname]; ok {
			fname = mappedName
		}

		items = append(items, selectionItem{name: fname, values: sel.Selected})
	}

	if len(items) == 0 {
		return &resRect, nil
	}

	// Print "Current Selection" title
	titleCellName, err := excelize.CoordinatesToCellName(rect.Left, rect.Top)
	if err != nil {
		return nil, util.Error("CoordinatesToCellName", err)
	}
	excel.SetCellStr(sheet, titleCellName, "Current Selection")
	boldStyle := &excelize.Style{Font: &excelize.Font{Bold: true}, Border: []excelize.Border{{Type: "bottom", Color: "000000", Style: 1}}}
	styleId, _ := excel.NewStyle(boldStyle)
	excel.SetCellStyle(sheet, titleCellName, titleCellName, styleId)

	r0 := rect.Top + 1
	c0 := rect.Left

	// Print field names row
	for ci, item := range items {
		cellName, err := excelize.CoordinatesToCellName(c0+ci, r0)
		if err != nil {
			logger.Err(err).Msg("CoordinatesToCellName")
			return nil, util.Error("CoordinatesToCellName", err)
		}
		logger.Debug().Msgf("print field name cell[%s]: %s", cellName, item.name)
		excel.SetCellStr(sheet, cellName, item.name)
		excel.SetCellStyle(sheet, cellName, cellName, styleId)
	}

	// Print field values row
	for ci, item := range items {
		cellName, err := excelize.CoordinatesToCellName(c0+ci, r0+1)
		if err != nil {
			logger.Err(err).Msg("CoordinatesToCellName")
			return nil, util.Error("CoordinatesToCellName", err)
		}
		logger.Debug().Msgf("print field value cell[%s]: %s", cellName, item.values)
		excel.SetCellStr(sheet, cellName, item.values)
	}

	resRect.Height = 3 // title + field names + field values
	resRect.Width = len(items)

	return &resRect, nil
}

// printReportTitle prints the customized report title
func (p *ExcelPagingPrinter) printReportTitle(title string, sheet string, excel *excelize.File, rect enigma.Rect, _logger *zerolog.Logger) (*enigma.Rect, *util.Result) {
	resRect := rect
	resRect.Height = 0
	resRect.Width = 0

	if title == "" {
		return &resRect, nil
	}

	cellName, err := excelize.CoordinatesToCellName(rect.Left, rect.Top)
	if err != nil {
		return nil, util.Error("CoordinatesToCellName", err)
	}

	excel.SetCellStr(sheet, cellName, title)

	// Bold and larger font for title
	titleStyle := &excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 14,
		},
	}
	styleId, _ := excel.NewStyle(titleStyle)
	excel.SetCellStyle(sheet, cellName, cellName, styleId)

	resRect.Height = 1
	resRect.Width = 1

	return &resRect, nil
}

// printTotalRecords prints the total records count
func (p *ExcelPagingPrinter) printTotalRecords(totalRows int, sheet string, excel *excelize.File, rect enigma.Rect, _logger *zerolog.Logger) (*enigma.Rect, *util.Result) {
	resRect := rect
	resRect.Height = 1
	resRect.Width = 2

	labelCell, err := excelize.CoordinatesToCellName(rect.Left, rect.Top)
	if err != nil {
		return nil, util.Error("CoordinatesToCellName", err)
	}
	excel.SetCellStr(sheet, labelCell, fmt.Sprintf("%s:", p.Config.TotalRecordsLabel))

	boldStyle := &excelize.Style{Font: &excelize.Font{Bold: true}}
	styleId, _ := excel.NewStyle(boldStyle)
	excel.SetCellStyle(sheet, labelCell, labelCell, styleId)

	valueCell, err := excelize.CoordinatesToCellName(rect.Left+1, rect.Top)
	if err != nil {
		return nil, util.Error("CoordinatesToCellName", err)
	}
	excel.SetCellInt(sheet, valueCell, int64(totalRows))

	return &resRect, nil
}

// printColumnNumbers prints column sequence numbers (1, 2, 3, ...)
func (p *ExcelPagingPrinter) printColumnNumbers(colCount int, sheet string, excel *excelize.File, rect enigma.Rect, _logger *zerolog.Logger) (*enigma.Rect, *util.Result) {
	resRect := rect
	resRect.Height = 1
	resRect.Width = colCount

	numStyle := &excelize.Style{
		Font:      &excelize.Font{Italic: true},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	}
	styleId, _ := excel.NewStyle(numStyle)

	for ci := 0; ci < colCount; ci++ {
		cellName, err := excelize.CoordinatesToCellName(rect.Left+ci, rect.Top)
		if err != nil {
			return nil, util.Error("CoordinatesToCellName", err)
		}
		excel.SetCellInt(sheet, cellName, int64(ci+1))
		excel.SetCellStyle(sheet, cellName, cellName, styleId)
	}

	return &resRect, nil
}

// printPageSubtotals prints subtotals for numeric columns
func (p *ExcelPagingPrinter) printPageSubtotals(subtotals []float64, isNumeric []bool, sheet string, excel *excelize.File, rect enigma.Rect, r Report, _logger *zerolog.Logger) (*enigma.Rect, *util.Result) {
	resRect := rect
	resRect.Height = 1
	resRect.Width = len(subtotals)

	boldStyle := &excelize.Style{
		Font: &excelize.Font{Bold: true},
		Border: []excelize.Border{
			{Type: "top", Color: "000000", Style: 2},
		},
	}
	if r.AllBorders {
		boldStyle.Border = []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		}
	}
	styleId, _ := excel.NewStyle(boldStyle)

	for ci, subtotal := range subtotals {
		cellName, err := excelize.CoordinatesToCellName(rect.Left+ci, rect.Top)
		if err != nil {
			return nil, util.Error("CoordinatesToCellName", err)
		}

		if ci == 0 {
			excel.SetCellStr(sheet, cellName, "Page Subtotal")
		} else if isNumeric[ci] {
			excel.SetCellFloat(sheet, cellName, subtotal, -1, 64)
		}
		excel.SetCellStyle(sheet, cellName, cellName, styleId)
	}

	return &resRect, nil
}

// printTableHeader prints the object column headers
func (p *ExcelPagingPrinter) printTableHeader(layout *engine.ObjectLayoutEx, sheet string, excel *excelize.File, rect enigma.Rect, r Report, _logger *zerolog.Logger) (*enigma.Rect, map[int]int, *util.Result) {
	if layout == nil {
		return nil, nil, util.MsgError("printTableHeader", "nil layout")
	}
	logger := _logger.With().Str("print", "tableHeader").Logger()

	resRect := rect
	resRect.Height = 1
	resRect.Width = 0

	DimCnt := len(layout.HyperCube.DimensionInfo)
	ColumnOrder := layout.HyperCube.ColumnOrder
	if len(ColumnOrder) == 0 {
		ColumnOrder = make([]int, len(layout.HyperCube.EffectiveInterColumnSortOrder))
		for i := range layout.HyperCube.EffectiveInterColumnSortOrder {
			ColumnOrder[i] = i
		}
	}

	var colInfo *engine.ColumnInfo
	layout.ColumnInfos = make([]*engine.ColumnInfo, 0)
	cube2report := make(map[int]int)
	ColCnt := 0

	boldStyle := &excelize.Style{Font: &excelize.Font{Bold: true}}
	if r.AllBorders {
		boldStyle.Border = []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		}
	}
	styleId, _ := excel.NewStyle(boldStyle)

	for _, colIx := range ColumnOrder {
		expIx := colIx - DimCnt
		if colIx < DimCnt {
			dim := layout.HyperCube.DimensionInfo[colIx]
			if dim.Error != nil {
				logger.Warn().Msgf("dim[%d] %s has error, ignore", colIx, dim.FallbackTitle)
				continue
			}
			colInfo = engine.NewColumnInfoFromDimension(dim)
		} else {
			exp := layout.HyperCube.MeasureInfo[expIx]
			if exp.Error != nil {
				logger.Warn().Msgf("exp[%d] %s has error, ignore", expIx, exp.FallbackTitle)
				continue
			}
			colInfo = engine.NewColumnInfoFromMeasure(exp)
		}
		layout.ColumnInfos = append(layout.ColumnInfos, colInfo)
		cellText := colInfo.FallbackTitle

		cube2report[ColCnt] = ColCnt
		if r.ColumnHeaderFormats != nil {
			if colHeaderFmt, ok := r.ColumnHeaderFormats[cellText]; ok {
				cube2report[ColCnt] = colHeaderFmt.Order
				if colHeaderFmt.Label != "" {
					cellText = colHeaderFmt.Label
				}
			}
		}
		repIdx := cube2report[ColCnt]

		cellName, err := excelize.CoordinatesToCellName(rect.Left+repIdx, rect.Top)
		if err != nil {
			return nil, nil, util.Error("CoordinatesToCellName", err)
		}

		logger.Debug().Msgf("print header cell[%s]: %s", cellName, cellText)
		excel.SetCellStr(sheet, cellName, cellText)
		excel.SetCellStyle(sheet, cellName, cellName, styleId)

		// Set column width
		colName, _, _ := excelize.SplitCellName(cellName)
		w, _ := excel.GetColWidth(sheet, colName)
		if w < float64(colInfo.ApprMaxGlyphCount) && colInfo.ApprMaxGlyphCount < 64 {
			excel.SetColWidth(sheet, colName, colName, float64(colInfo.ApprMaxGlyphCount))
		}

		ColCnt++
	}

	// Handle static columns
	if r.ColumnHeaderFormats != nil {
		for _, colHeaderFmt := range r.ColumnHeaderFormats {
			if colHeaderFmt.ColumnType == StaticColumnType {
				repIdx := colHeaderFmt.Order
				cellName, _ := excelize.CoordinatesToCellName(rect.Left+repIdx, rect.Top)
				excel.SetCellStr(sheet, cellName, colHeaderFmt.Label)
				excel.SetCellStyle(sheet, cellName, cellName, styleId)
				ColCnt++
			}
		}
	}

	resRect.Width = ColCnt
	return &resRect, cube2report, nil
}

// printTableRows prints a subset of rows for the current page
func (p *ExcelPagingPrinter) printTableRows(rows [][]*enigma.NxCell, layout *engine.ObjectLayoutEx, cube2report map[int]int,
	sheet string, excel *excelize.File, rect enigma.Rect, r Report, _logger *zerolog.Logger) ([]float64, []bool, *util.Result) {

	logger := _logger.With().Str("print", "tableRows").Logger()
	colCount := len(layout.ColumnInfos)
	subtotals := make([]float64, colCount)
	isNumeric := make([]bool, colCount)

	// Determine which columns are numeric
	for ci, colInfo := range layout.ColumnInfos {
		if colInfo != nil && colInfo.NumFormat != nil && colInfo.NumFormat.Type != "U" {
			isNumeric[ci] = true
		}
	}

	for ri, rowCells := range rows {
		for ci, cell := range rowCells {
			if ci >= len(cube2report) {
				continue
			}
			if cell == nil {
				continue
			}
			reportColIx := cube2report[ci]
			reportRowIx := rect.Top + ri

			cellName, err := excelize.CoordinatesToCellName(rect.Left+reportColIx, reportRowIx)
			if err != nil {
				logger.Err(err).Msg("CoordinatesToCellName")
				return nil, nil, util.Error("CoordinatesToCellName", err)
			}

			cellNum := float64(cell.Num)
			isNum := !math.IsNaN(cellNum)

			hasColInfo := ci < len(layout.ColumnInfos) && layout.ColumnInfos[ci] != nil
			if isNum && hasColInfo {
				colInfo := layout.ColumnInfos[ci]
				if colInfo.NumFormat != nil && colInfo.NumFormat.Type == "U" {
					isNum = false
				}
			}

			if isNum && hasColInfo {
				excel.SetCellFloat(sheet, cellName, cellNum, -1, 64)
				subtotals[ci] += cellNum
				isNumeric[ci] = true
			} else {
				excel.SetCellStr(sheet, cellName, cell.Text)
			}

			// Apply cell style
			var excelStyle *excelize.Style
			if cell.AttrExps != nil && len(cell.AttrExps.Values) > 0 {
				excelStyle, _ = GetStackCellStyle(cell, &logger)
			}

			if hasColInfo {
				colInfo := layout.ColumnInfos[ci]
				if colInfo.NumFormat != nil && colInfo.NumFormat.Fmt != "" {
					if excelStyle == nil {
						excelStyle = &excelize.Style{}
					}
					excelStyle.CustomNumFmt = &colInfo.NumFormat.Fmt
				}
			}

			if r.AllBorders {
				if excelStyle == nil {
					excelStyle = &excelize.Style{}
				}
				excelStyle.Border = []excelize.Border{
					{Type: "left", Color: "000000", Style: 1},
					{Type: "top", Color: "000000", Style: 1},
					{Type: "right", Color: "000000", Style: 1},
					{Type: "bottom", Color: "000000", Style: 1},
				}
			}

			if excelStyle != nil {
				styleId, _ := excel.NewStyle(excelStyle)
				excel.SetCellStyle(sheet, cellName, cellName, styleId)
			}
		}

		// Handle static columns
		if r.ColumnHeaderFormats != nil {
			for _, colFormat := range r.ColumnHeaderFormats {
				if colFormat.ColumnType == StaticColumnType {
					reportColIx := colFormat.Order
					reportRowIx := rect.Top + ri
					cellName, _ := excelize.CoordinatesToCellName(rect.Left+reportColIx, reportRowIx)
					excel.SetCellStr(sheet, cellName, colFormat.StaticValue)

					if r.AllBorders {
						borderStyle := &excelize.Style{
							Border: []excelize.Border{
								{Type: "left", Color: "000000", Style: 1},
								{Type: "top", Color: "000000", Style: 1},
								{Type: "right", Color: "000000", Style: 1},
								{Type: "bottom", Color: "000000", Style: 1},
							},
						}
						styleId, _ := excel.NewStyle(borderStyle)
						excel.SetCellStyle(sheet, cellName, cellName, styleId)
					}
				}
			}
		}
	}

	return subtotals, isNumeric, nil
}

// printPage prints a single page with all sections
func (p *ExcelPagingPrinter) printPage(pageNum int, rows [][]*enigma.NxCell, totalRows int, layout *engine.ObjectLayoutEx,
	cube2report map[int]int, r Report, doc *enigma.Doc, excel *excelize.File, _logger *zerolog.Logger) *util.Result {

	sheetName := fmt.Sprintf("page-%d", pageNum)
	logger := _logger.With().Str("sheet", sheetName).Int("page", pageNum).Logger()

	excel.NewSheet(sheetName)

	currentRow := 1
	colCount := len(layout.ColumnInfos)

	// 1. Report Title
	if p.Config.ReportTitle != "" {
		rect := enigma.Rect{Top: currentRow, Left: 1}
		titleRect, res := p.printReportTitle(p.Config.ReportTitle, sheetName, excel, rect, &logger)
		if res != nil {
			return res.With("printReportTitle")
		}
		currentRow += titleRect.Height + 1 // +1 for blank row
	}

	// 2. Current Selections (horizontal)
	if r.OutputCurrentSelection {
		rect := enigma.Rect{Top: currentRow, Left: 1}
		selRect, res := p.printHorizontalSelection(r, doc, sheetName, excel, rect, &logger)
		if res != nil {
			return res.With("printHorizontalSelection")
		}
		if selRect.Height > 0 {
			currentRow += selRect.Height + 1 // +1 for blank row
		}
	}

	// 3. Custom Headers
	if len(r.Headers) > 0 {
		rect := enigma.Rect{Top: currentRow, Left: 1}
		for hi, header := range r.Headers {
			labelCell, _ := excelize.CoordinatesToCellName(1, currentRow+hi)
			excel.SetCellStr(sheetName, labelCell, header.Label)

			textCell, _ := excelize.CoordinatesToCellName(2, currentRow+hi)
			if text := strings.TrimSpace(header.Text); strings.HasPrefix(text, "=") {
				dual, err := doc.EvaluateEx(engine.ConnCtx, header.Text)
				if err == nil {
					evalText := dual.Text
					if evalText == "" && dual.IsNumeric {
						evalText = fmt.Sprintf("%v", dual.Number)
					}
					excel.SetCellStr(sheetName, textCell, evalText)
				} else {
					excel.SetCellStr(sheetName, textCell, header.Text)
				}
			} else {
				excel.SetCellStr(sheetName, textCell, header.Text)
			}
		}
		currentRow += len(r.Headers) + 1 // +1 for blank row
		_ = rect
	}

	// 4. Total Records
	{
		rect := enigma.Rect{Top: currentRow, Left: 1}
		_, res := p.printTotalRecords(totalRows, sheetName, excel, rect, &logger)
		if res != nil {
			return res.With("printTotalRecords")
		}
		currentRow += 2 // +1 for the row, +1 for blank row
	}

	// 5. Table Header
	headerRect := enigma.Rect{Top: currentRow, Left: 1}
	{
		hRect, _, res := p.printTableHeader(layout, sheetName, excel, headerRect, r, &logger)
		if res != nil {
			return res.With("printTableHeader")
		}
		currentRow += hRect.Height
	}

	// 6. Column Numbers (optional)
	if p.Config.ShowColumnNumbers {
		rect := enigma.Rect{Top: currentRow, Left: 1}
		numRect, res := p.printColumnNumbers(colCount, sheetName, excel, rect, &logger)
		if res != nil {
			return res.With("printColumnNumbers")
		}
		currentRow += numRect.Height
	}

	// 7. Table Rows
	dataRect := enigma.Rect{Top: currentRow, Left: 1}
	subtotals, isNumeric, res := p.printTableRows(rows, layout, cube2report, sheetName, excel, dataRect, r, &logger)
	if res != nil {
		return res.With("printTableRows")
	}
	currentRow += len(rows)

	// 8. Page Subtotals (optional)
	if p.Config.ShowSubtotals {
		rect := enigma.Rect{Top: currentRow, Left: 1}
		_, res := p.printPageSubtotals(subtotals, isNumeric, sheetName, excel, rect, r, &logger)
		if res != nil {
			return res.With("printPageSubtotals")
		}
	}

	logger.Info().Msgf("page %d printed with %d rows", pageNum, len(rows))
	return nil
}

// Print exports a Qlik object to paginated Excel sheets
func (p *ExcelPagingPrinter) Print(r Report) *util.Result {
	if !r.IsValid() {
		return util.MsgError("Print", "invalid report")
	}

	if r.Name != nil {
		p.Config.ReportTitle = *r.Name
	}
	if r.PaginationConfig != nil {
		if r.PaginationConfig.RowsPerPage > 0 {
			p.Config.RowsPerPage = r.PaginationConfig.RowsPerPage
		}
		if r.PaginationConfig.TotalRecordsLabel != "" {
			p.Config.TotalRecordsLabel = r.PaginationConfig.TotalRecordsLabel
		}
	}

	rResult, res := NewReportResult(r)
	if res != nil {
		return res.With("NewReportResult")
	}
	p.ReportResults[util.MaybeNil(r.ID)] = rResult
	logger := rResult.Logger.With().Str("report", *r.ID).Str("driver", "excel_paging").Logger()

	if r.Doc == nil {
		return util.MsgError("CheckDoc", "doc is not opened")
	}

	if len(r.TargetIDs) != 1 {
		return util.MsgError("CheckTargets", "excel_paging driver supports exactly one object")
	}

	objId := r.TargetIDs[0]
	logger.Info().Msgf("printing object: %s", objId)

	obj, err := r.Doc.GetObject(engine.ConnCtx, objId)
	if err != nil {
		return util.Error("GetObject", err)
	}
	if obj.Handle == 0 {
		return util.MsgError("GetObject", fmt.Sprintf("can't get object %s", objId))
	}

	objLayout, res := engine.GetObjectLayoutEx(obj)
	if res != nil {
		return res.With("GetObjectLayoutEx")
	}

	if objLayout.HyperCube == nil {
		return util.MsgError("CheckHyperCube", "object has no hypercube")
	}

	// Only support stack tables (mode S)
	if objLayout.HyperCube.Mode != "S" && objLayout.HyperCube.Mode != "" {
		return util.MsgError("CheckMode", fmt.Sprintf("excel_paging only supports stack tables (mode S), got mode %s", objLayout.HyperCube.Mode))
	}

	if objLayout.HyperCube.Error != nil {
		cubeErr := objLayout.HyperCube.Error
		return util.MsgError("CheckHyperCube", fmt.Sprintf("hypercube error: %d - %s", cubeErr.ErrorCode, cubeErr.ExtendedMessage))
	}

	totalRows := objLayout.HyperCube.Size.Cy
	logger.Info().Msgf("total rows: %d, rows per page: %d", totalRows, p.Config.RowsPerPage)

	// Fetch all data
	dataPages, res := engine.GetHyperCubeData(obj, *objLayout.HyperCube.Size)
	if res != nil {
		return res.With("GetHyperCubeData")
	}

	// Build cube2report mapping first (need column count for row conversion)
	DimCnt := len(objLayout.HyperCube.DimensionInfo)
	ColumnOrder := objLayout.HyperCube.ColumnOrder
	if len(ColumnOrder) == 0 {
		ColumnOrder = make([]int, len(objLayout.HyperCube.EffectiveInterColumnSortOrder))
		for i := range objLayout.HyperCube.EffectiveInterColumnSortOrder {
			ColumnOrder[i] = i
		}
	}

	objLayout.ColumnInfos = make([]*engine.ColumnInfo, 0)
	cube2report := make(map[int]int)
	ColCnt := 0

	for _, colIx := range ColumnOrder {
		var colInfo *engine.ColumnInfo
		expIx := colIx - DimCnt
		if colIx < DimCnt {
			dim := objLayout.HyperCube.DimensionInfo[colIx]
			if dim.Error != nil {
				continue
			}
			colInfo = engine.NewColumnInfoFromDimension(dim)
		} else {
			exp := objLayout.HyperCube.MeasureInfo[expIx]
			if exp.Error != nil {
				continue
			}
			colInfo = engine.NewColumnInfoFromMeasure(exp)
		}
		objLayout.ColumnInfos = append(objLayout.ColumnInfos, colInfo)

		cube2report[ColCnt] = ColCnt
		cellText := colInfo.FallbackTitle
		if r.ColumnHeaderFormats != nil {
			if colHeaderFmt, ok := r.ColumnHeaderFormats[cellText]; ok {
				cube2report[ColCnt] = colHeaderFmt.Order
			}
		}
		ColCnt++
	}

	// Reorganize pages into a row-based structure to handle column pagination
	// When hypercube has many columns, data is split into multiple pages with different Area.Left offsets
	// Map: rowIndex -> map[colIndex]cell
	rowData := make(map[int]map[int]*enigma.NxCell)
	maxRow := 0

	for _, page := range dataPages {
		if page.Area.Height < 1 {
			continue
		}

		for ri, rowCells := range page.Matrix {
			absoluteRowIdx := page.Area.Top + ri
			if absoluteRowIdx > maxRow {
				maxRow = absoluteRowIdx
			}

			if rowData[absoluteRowIdx] == nil {
				rowData[absoluteRowIdx] = make(map[int]*enigma.NxCell)
			}

			for ci, cell := range rowCells {
				cubeColIx := page.Area.Left + ci
				rowData[absoluteRowIdx][cubeColIx] = cell
			}
		}
	}

	// Convert map structure back to ordered row slices
	colCount := len(objLayout.ColumnInfos)
	allRows := make([][]*enigma.NxCell, 0, maxRow+1)
	for rowIdx := 0; rowIdx <= maxRow; rowIdx++ {
		row := make([]*enigma.NxCell, colCount)
		if cells, ok := rowData[rowIdx]; ok {
			for colIdx, cell := range cells {
				if colIdx < colCount {
					row[colIdx] = cell
				}
			}
		}
		allRows = append(allRows, row)
	}
	logger.Info().Msgf("fetched %d rows from %d data pages", len(allRows), len(dataPages))

	// Create Excel file
	excel := excelize.NewFile()

	// Calculate pages
	pageCount := (len(allRows) + p.Config.RowsPerPage - 1) / p.Config.RowsPerPage
	if pageCount == 0 {
		pageCount = 1 // At least one page even if empty
	}
	logger.Info().Msgf("page count: %d", pageCount)

	// Print each page
	for pageIdx := 0; pageIdx < pageCount; pageIdx++ {
		startRow := pageIdx * p.Config.RowsPerPage
		endRow := startRow + p.Config.RowsPerPage
		if endRow > len(allRows) {
			endRow = len(allRows)
		}

		pageRows := allRows[startRow:endRow]
		if res := p.printPage(pageIdx+1, pageRows, totalRows, objLayout, cube2report, r, r.Doc, excel, &logger); res != nil {
			return res.With(fmt.Sprintf("printPage[%d]", pageIdx+1))
		}
	}

	// Remove default sheet and save
	excel.DeleteSheet("Sheet1")

	if err := excel.SaveAs(*rResult.ReportFile); err != nil {
		return util.Error("SaveWorkBook", err)
	}
	logger.Info().Msgf("report saved as [%s]", *rResult.ReportFile)

	return nil
}

// GetReportResult returns the result for a given report ID
func (p *ExcelPagingPrinter) GetReportResult(id string) (*ReportResult, *util.Result) {
	result, ok := p.ReportResults[id]
	if !ok {
		return nil, util.MsgError("ReportFiles", "report id doesn't exist")
	}
	return result, nil
}
