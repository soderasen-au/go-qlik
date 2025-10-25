package report

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/qlik-oss/enigma-go/v4"
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/util"

	"github.com/soderasen-au/go-qlik/qlik/engine"
)

const (
	PDF_MARGIN_LEFT   = 10.0
	PDF_MARGIN_TOP    = 10.0
	PDF_MARGIN_RIGHT  = 10.0
	PDF_FONT_SIZE     = 9.0
	PDF_HEADER_SIZE   = 10.0
	PDF_LINE_HEIGHT   = 5.0
	PDF_CELL_PADDING  = 1.0
	PDF_PAGE_WIDTH    = 297.0 // A4 landscape width in mm (reference only - actual dimensions set per orientation)
	PDF_PAGE_HEIGHT   = 210.0 // A4 landscape height in mm (reference only - actual dimensions set per orientation)
	PDF_MAX_COL_WIDTH = 80.0  // Increased from 50.0 to reduce text truncation
	PDF_MIN_COL_WIDTH = 15.0
)

type PdfReportPrinter struct {
	ReportPrinterBase
	pdf              *gofpdf.Fpdf
	colWidths        []float64
	displayColWidths []float64                  // Column widths in display order
	cube2report      map[int]int                // cube column index => report column index
	displayColInfo   map[int]*engine.ColumnInfo // display column index => ColumnInfo
	printedRows      int
	currentY         float64 // Track current Y position for multi-object layouts
	pageWidth        float64 // Actual page width based on orientation
	pageHeight       float64 // Actual page height based on orientation
}

func NewPdfReportPrinter() *PdfReportPrinter {
	p := &PdfReportPrinter{}
	p.ReportResults = make(map[string]*ReportResult)
	return p
}

// Calculate optimal column widths based on content
func (p *PdfReportPrinter) calculateColumnWidths(layout *engine.ObjectLayoutEx) {
	DimCnt := len(layout.HyperCube.DimensionInfo)
	ColumnOrder := layout.HyperCube.ColumnOrder
	if ColumnOrder == nil || len(ColumnOrder) == 0 {
		ColumnOrder = layout.HyperCube.EffectiveInterColumnSortOrder
	}

	// Build widths only for valid (non-error) columns
	p.colWidths = make([]float64, 0, len(ColumnOrder))
	availableWidth := p.pageWidth - PDF_MARGIN_LEFT - PDF_MARGIN_RIGHT

	for _, colIx := range ColumnOrder {
		var colInfo *engine.ColumnInfo
		expIx := colIx - DimCnt
		if colIx < DimCnt {
			dim := layout.HyperCube.DimensionInfo[colIx]
			if dim.Error != nil {
				continue
			}
			colInfo = engine.NewColumnInfoFromDimension(dim)
		} else {
			exp := layout.HyperCube.MeasureInfo[expIx]
			if exp.Error != nil {
				continue
			}
			colInfo = engine.NewColumnInfoFromMeasure(exp)
		}

		// Calculate width based on ApprMaxGlyphCount
		// Use average character width from font metrics
		avgCharWidth := p.pdf.GetStringWidth("M") / 1.6 // M is widest, average is ~60% of M width

		// Calculate widths from header and content
		headerWidth := p.pdf.GetStringWidth(colInfo.FallbackTitle)
		contentWidth := float64(colInfo.ApprMaxGlyphCount) * avgCharWidth

		// Add 20% safety margin since ApprMaxGlyphCount can underestimate actual text width
		// and we use exact GetStringWidth() when rendering cells
		contentWidth *= 1.2

		// Use the larger of header or content width
		width := headerWidth
		if contentWidth > width {
			width = contentWidth
		}

		if width < PDF_MIN_COL_WIDTH {
			width = PDF_MIN_COL_WIDTH
		}
		if width > PDF_MAX_COL_WIDTH {
			width = PDF_MAX_COL_WIDTH
		}
		p.colWidths = append(p.colWidths, width)
	}

	// Normalize widths to fit page
	// First, try reducing max column width gradually before scaling
	maxWidthLimit := PDF_MAX_COL_WIDTH
	for maxWidthLimit >= PDF_MIN_COL_WIDTH {
		totalWidth := 0.0
		for _, w := range p.colWidths {
			width := w
			if width > maxWidthLimit {
				width = maxWidthLimit
			}
			totalWidth += width
		}

		if totalWidth <= availableWidth {
			// Fits! Apply the max width cap
			for i := range p.colWidths {
				if p.colWidths[i] > maxWidthLimit {
					p.colWidths[i] = maxWidthLimit
				}
			}
			return
		}

		// Still too wide, try smaller max
		maxWidthLimit -= 5.0
	}

	// Last resort: proportional scaling
	// When scaling is needed, we need extra padding because cells will be tighter
	totalWidth := 0.0
	for _, w := range p.colWidths {
		totalWidth += w
	}
	if totalWidth > availableWidth {
		// Scale to slightly less than available width to leave breathing room
		scale := (availableWidth * 0.98) / totalWidth
		//logger.Debug().Float64("scale", scale).Float64("totalWidth", totalWidth).Float64("availableWidth", availableWidth).Msg("Scaling columns to fit page")
		for i := range p.colWidths {
			//oldWidth := p.colWidths[i]
			p.colWidths[i] *= scale
			//logger.Debug().Int("col", i).Float64("oldWidth", oldWidth).Float64("newWidth", p.colWidths[i]).Msg("Column scaled")
		}
	}
}

// Apply cell styling from Qlik attributes
func (p *PdfReportPrinter) applyCellStyle(cell *enigma.NxCell, logger *zerolog.Logger) {
	if cell.AttrExps == nil || cell.AttrExps.Values == nil || len(cell.AttrExps.Values) == 0 {
		return
	}

	attrs := cell.AttrExps.Values
	if len(attrs) > 0 {
		bgAttr := attrs[0]
		bgColor, res := NewARGBColorFromQlikAttr(bgAttr)
		if res == nil && bgColor != nil {
			// Set fill color
			p.pdf.SetFillColor(bgColor.R, bgColor.G, bgColor.B)

			// Set text color based on background luminance or explicit font color
			if len(attrs) > 1 {
				fontAttr := attrs[1]
				fontColor, res := NewARGBColorFromQlikAttr(fontAttr)
				if res == nil && fontColor != nil {
					p.pdf.SetTextColor(fontColor.R, fontColor.G, fontColor.B)
				}
			} else {
				// Auto-contrast: use white text on dark bg, black on light bg
				luminance := (0.299*float64(bgColor.R) + 0.587*float64(bgColor.G) + 0.114*float64(bgColor.B)) / 255.0
				if luminance > 0.5 {
					p.pdf.SetTextColor(0, 0, 0)
				} else {
					p.pdf.SetTextColor(255, 255, 255)
				}
			}
		}
	}
}

// Reset to default cell styling
func (p *PdfReportPrinter) resetCellStyle() {
	p.pdf.SetFillColor(255, 255, 255) // White background
	p.pdf.SetTextColor(0, 0, 0)       // Black text
	p.pdf.SetFont("Arial", "", PDF_FONT_SIZE)
}

// Print table header
func (p *PdfReportPrinter) printObjectHeader(layout *engine.ObjectLayoutEx, r Report, logger *zerolog.Logger) *util.Result {
	if layout == nil {
		return util.MsgError("printObjectHeader", "nil layout")
	}

	logger.Debug().Msg("printing header")
	DimCnt := len(layout.HyperCube.DimensionInfo)
	ColumnOrder := layout.HyperCube.ColumnOrder
	if ColumnOrder == nil || len(ColumnOrder) == 0 {
		ColumnOrder = layout.HyperCube.EffectiveInterColumnSortOrder
	}

	printTotals := layout.Totals != nil && layout.Totals.Show && len(layout.HyperCube.GrandTotalRow) > 0

	p.pdf.SetFont("Arial", "B", PDF_HEADER_SIZE)
	p.pdf.SetFillColor(240, 240, 240) // Light gray background for header
	p.pdf.SetTextColor(0, 0, 0)

	layout.ColumnInfos = make([]*engine.ColumnInfo, 0)
	p.cube2report = make(map[int]int)                      // Initialize cube to report column mapping
	p.displayColInfo = make(map[int]*engine.ColumnInfo)    // Initialize display column info mapping
	p.displayColWidths = make([]float64, len(p.colWidths)) // Initialize display column widths
	ColCnt := 0
	ExpCnt := 0
	for _, colIx := range ColumnOrder {
		var colInfo *engine.ColumnInfo
		expIx := colIx - DimCnt

		if colIx < DimCnt {
			dim := layout.HyperCube.DimensionInfo[colIx]
			if dim.Error != nil {
				logger.Warn().Msgf("dim[%d] %s has error, skipping", colIx, dim.FallbackTitle)
				continue
			}
			colInfo = engine.NewColumnInfoFromDimension(dim)
		} else {
			exp := layout.HyperCube.MeasureInfo[expIx]
			if exp.Error != nil {
				logger.Warn().Msgf("exp[%d] %s has error, skipping", expIx, exp.FallbackTitle)
				continue
			}
			colInfo = engine.NewColumnInfoFromMeasure(exp)
		}

		layout.ColumnInfos = append(layout.ColumnInfos, colInfo)
		cellText := colInfo.FallbackTitle

		// Build cube to report column mapping: data column position -> display column index
		p.cube2report[ColCnt] = ColCnt
		if r.ColumnHeaderFormats != nil {
			if colHeaderFmt, ok := r.ColumnHeaderFormats[cellText]; ok {
				logger.Debug().Msgf("data[%d]:%s => report[%d]", ColCnt, cellText, colHeaderFmt.Order)
				p.cube2report[ColCnt] = colHeaderFmt.Order
			}
		}
		repIdx := p.cube2report[ColCnt]

		// Store ColumnInfo and width at display position
		p.displayColInfo[repIdx] = colInfo
		if ColCnt < len(p.colWidths) && repIdx < len(p.displayColWidths) {
			p.displayColWidths[repIdx] = p.colWidths[ColCnt]
		}

		// Apply custom header formatting if configured
		if r.ColumnHeaderFormats != nil {
			if colHeaderFmt, ok := r.ColumnHeaderFormats[cellText]; ok {
				if colHeaderFmt.Label != "" {
					cellText = colHeaderFmt.Label
				}
				// Apply custom header colors
				if colHeaderFmt.BgColor != "" {
					if bgColor, res := NewARGBFromQlikColor(colHeaderFmt.BgColor); res == nil && bgColor != nil {
						p.pdf.SetFillColor(bgColor.R, bgColor.G, bgColor.B)
					}
				}
				if colHeaderFmt.FgColor != "" {
					if fgColor, res := NewARGBFromQlikColor(colHeaderFmt.FgColor); res == nil && fgColor != nil {
						p.pdf.SetTextColor(fgColor.R, fgColor.G, fgColor.B)
					}
				}
			}
		}

		if repIdx < len(p.displayColWidths) {
			p.pdf.CellFormat(p.displayColWidths[repIdx], PDF_LINE_HEIGHT, cellText, "1", 0, "C", true, 0, "")
		}

		// Reset colors for next header cell
		p.pdf.SetFillColor(240, 240, 240)
		p.pdf.SetTextColor(0, 0, 0)

		ColCnt++
		if expIx >= 0 {
			ExpCnt++
		}
	}

	p.pdf.Ln(-1)
	p.printedRows++

	// Print totals row if configured (matches Excel behavior)
	if printTotals {
		p.pdf.SetFont("Arial", "", PDF_FONT_SIZE) // Use normal font, not bold
		ExpCnt := 0
		ColCnt := 0
		for _, colIx := range ColumnOrder {
			expIx := colIx - DimCnt
			totalText := ""

			// Only print totals for measure columns (expIx >= 0)
			if expIx >= 0 && ExpCnt < len(layout.HyperCube.GrandTotalRow) && layout.HyperCube.GrandTotalRow[ExpCnt] != nil {
				totalText = layout.HyperCube.GrandTotalRow[ExpCnt].Text
				ExpCnt++
			}

			repIdx, ok := p.cube2report[ColCnt]
			if !ok {
				repIdx = ColCnt
			}
			if repIdx < len(p.displayColWidths) {
				p.pdf.CellFormat(p.displayColWidths[repIdx], PDF_LINE_HEIGHT, totalText, "1", 0, "", false, 0, "")
			}
			ColCnt++
		}
		p.pdf.Ln(-1)
		p.printedRows++
	}

	// Add static columns from ColumnHeaderFormats
	if r.ColumnHeaderFormats != nil {
		for _, colHeaderFmt := range r.ColumnHeaderFormats {
			if colHeaderFmt.ColumnType == StaticColumnType {
				// Note: Static columns are handled at the data row level in printStackObject
				// Here we just need to ensure we have the column info
				logger.Debug().Msgf("Static column detected: %s", colHeaderFmt.Label)
			}
		}
	}

	return nil
}

// Print a single data cell
func (p *PdfReportPrinter) printCell(cell *enigma.NxCell, colWidth float64, colInfo *engine.ColumnInfo, logger *zerolog.Logger) {
	p.resetCellStyle()
	p.applyCellStyle(cell, logger)

	cellText := cell.Text
	cellNum := float64(cell.Num)
	isNum := !math.IsNaN(cellNum)

	// Use numeric value if available and not explicitly marked as text
	if isNum && colInfo != nil && colInfo.NumFormat != nil && colInfo.NumFormat.Type != "U" {
		// For numbers, use the text representation from Qlik (already formatted)
		cellText = cell.Text
	}

	// Truncate long text to fit cell using actual font metrics
	textWidth := p.pdf.GetStringWidth(cellText)
	// gofpdf's CellFormat actually fits text within the specified width with internal padding
	// We need to account for approximately 1mm total internal margin based on empirical testing
	availableWidth := colWidth - 1.0

	if textWidth > availableWidth {
		// Iteratively reduce text until it fits with "..."
		suffix := "..."
		suffixWidth := p.pdf.GetStringWidth(suffix)

		for len(cellText) > 0 {
			if p.pdf.GetStringWidth(cellText)+suffixWidth <= availableWidth {
				cellText = cellText + suffix
				break
			}
			// Remove last character (handle multi-byte UTF-8 properly)
			runes := []rune(cellText)
			if len(runes) > 0 {
				cellText = string(runes[:len(runes)-1])
			} else {
				break
			}
		}
	}

	p.pdf.CellFormat(colWidth, PDF_LINE_HEIGHT, cellText, "1", 0, "", true, 0, "")
}

// Print current selection (field filters)
func (p *PdfReportPrinter) printCurrentSelection(r Report, doc *enigma.Doc, logger *zerolog.Logger) *util.Result {
	logger.Info().Msg("printing current selection")

	selObj, res := engine.GetCurrentSelection(doc, "$")
	if res != nil {
		return res.With("GetCurrentSelection")
	}

	// Title
	p.pdf.SetFont("Arial", "B", PDF_HEADER_SIZE)
	p.pdf.Cell(0, PDF_LINE_HEIGHT, "Current Selection")
	p.pdf.Ln(-1)
	p.printedRows++

	// Get dimension label map
	dimLabelMap := make(map[string]string)
	dimList, res := engine.GetDimensionList(doc)
	if res == nil {
		for _, dimItem := range dimList {
			dimObj, _ := doc.GetDimension(engine.ConnCtx, dimItem.Info.Id)
			dimLayout, _ := dimObj.GetLayout(engine.ConnCtx)
			dim := dimLayout.Dim
			if len(dim.FieldDefs) < 1 {
				continue
			}
			dimDef := dim.FieldDefs[0]
			dimLabel := dim.LabelExpression
			if len(dim.FieldLabels) > 0 {
				dimLabel = dim.FieldLabels[0]
			}
			dimLabelMap[dimDef] = dimLabel
		}
	}

	// Print selections
	p.pdf.SetFont("Arial", "", PDF_FONT_SIZE)
	for _, sel := range selObj.Selections {
		fname := sel.Field

		// Check if hidden
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

		// Use dimension label if available
		if strings.HasPrefix(fname, "=") {
			if dname, ok := dimLabelMap[fname]; ok {
				fname = dname
			}
		}

		// Check page break
		if p.pdf.GetY() > p.pageHeight-PDF_MARGIN_TOP-PDF_LINE_HEIGHT*2 {
			p.pdf.AddPage()
		}

		// Print field name and selected values
		p.pdf.CellFormat(50, PDF_LINE_HEIGHT, fname, "1", 0, "", false, 0, "")
		p.pdf.CellFormat(0, PDF_LINE_HEIGHT, sel.Selected, "1", 0, "", false, 0, "")
		p.pdf.Ln(-1)
		p.printedRows++
	}

	p.pdf.Ln(3) // Add spacing
	return nil
}

// Print custom headers
func (p *PdfReportPrinter) printCustomHeaders(doc *enigma.Doc, headers []CustomHeader, logger *zerolog.Logger) *util.Result {
	logger.Info().Msg("printing custom headers")

	for _, header := range headers {
		// Check page break
		if p.pdf.GetY() > p.pageHeight-PDF_MARGIN_TOP-PDF_LINE_HEIGHT*2 {
			p.pdf.AddPage()
		}

		text := header.Text
		if t := strings.TrimSpace(text); strings.HasPrefix(t, "=") {
			dual, err := doc.EvaluateEx(engine.ConnCtx, header.Text)
			if err != nil {
				logger.Err(err).Msg("EvaluateEx")
				return util.Error("EvaluateEx", err)
			}
			text = dual.Text
			if text == "" && dual.IsNumeric {
				text = fmt.Sprintf("%v", dual.Number)
			}
		}

		p.pdf.SetFont("Arial", "B", PDF_FONT_SIZE)
		p.pdf.CellFormat(50, PDF_LINE_HEIGHT, header.Label, "1", 0, "", false, 0, "")
		p.pdf.SetFont("Arial", "", PDF_FONT_SIZE)
		p.pdf.CellFormat(0, PDF_LINE_HEIGHT, text, "1", 0, "", false, 0, "")
		p.pdf.Ln(-1)
		p.printedRows++
	}

	p.pdf.Ln(3) // Add spacing
	return nil
}

// Print sheet header (combines current selection and custom headers)
func (p *PdfReportPrinter) printSheetHeader(r Report, doc *enigma.Doc, logger *zerolog.Logger) *util.Result {
	if r.OutputCurrentSelection {
		if res := p.printCurrentSelection(r, doc, logger); res != nil {
			return res.With("printCurrentSelection")
		}
	}

	if len(r.Headers) > 0 {
		if res := p.printCustomHeaders(doc, r.Headers, logger); res != nil {
			return res.With("printCustomHeaders")
		}
	}

	return nil
}

// Print container object (object with children)
func (p *PdfReportPrinter) printContainer(r Report, objId string, logger *zerolog.Logger) *util.Result {
	logger.Info().Msgf("printing container object: %s", objId)

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

	if objLayout.ChildList == nil || len(objLayout.ChildList.Items) == 0 {
		logger.Warn().Msg("container has no children")
		return nil
	}

	// Print container title if available
	prop := engine.ObjectPropeties{Info: objLayout.Info}
	rawProp, err := obj.GetPropertiesRaw(engine.ConnCtx)
	if err == nil {
		prop.Properties = rawProp
		title, _ := engine.GetTitle(objLayout.Info, &prop, logger)
		if title != nil && strings.TrimSpace(*title) != "" {
			p.pdf.SetFont("Arial", "B", PDF_HEADER_SIZE+2)
			p.pdf.Cell(0, PDF_LINE_HEIGHT*1.5, *title)
			p.pdf.Ln(-1)
			p.pdf.Ln(2)
			p.printedRows++
		}
	}

	// Build child map
	childArray := make(map[string]string) // refId -> objectId
	for ci, entry := range objLayout.ChildList.Items {
		info := &engine.ContainerChildInfo{}
		err := json.Unmarshal(entry.Data, info)
		if err != nil {
			logger.Warn().Int("child", ci).Err(err).Msg("failed to unmarshal child info")
			continue
		}
		childArray[info.ContainerChildId] = entry.Info.Id
		childArray[info.QExtendsId] = entry.Info.Id
	}

	// Print each child
	for ci, child := range objLayout.Children {
		childID, ok := childArray[child.RefId]
		if !ok {
			logger.Warn().Int("child", ci).Str("refId", child.RefId).Msg("child not found in child list")
			continue
		}

		// Print child label if available
		if len(child.Label) > 0 {
			p.pdf.SetFont("Arial", "B", PDF_FONT_SIZE+1)
			p.pdf.Cell(0, PDF_LINE_HEIGHT, child.Label)
			p.pdf.Ln(-1)
			p.printedRows++
		}

		// Print child object
		childLogger := logger.With().Int("child", ci).Str("id", childID).Logger()
		if res := p.printObject(r, childID, &childLogger); res != nil {
			logger.Err(res).Msgf("failed to print child %s", childID)
			return res.With("printObject")
		}

		// Add spacing between children
		if ci < len(objLayout.Children)-1 {
			p.pdf.Ln(3)
		}
	}

	return nil
}

// Print object (dispatcher for different object types)
func (p *PdfReportPrinter) printObject(r Report, objId string, logger *zerolog.Logger) *util.Result {
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

	// Container objects
	if objLayout.Info.Type == "container" {
		return p.printContainer(r, objId, logger)
	}

	// Pivot tables - currently not fully supported in PDF, print as stack
	if objLayout.HyperCube != nil && (objLayout.HyperCube.Mode == "P" || objLayout.HyperCube.Mode == "K") {
		logger.Warn().Msg("Pivot tables are not fully supported in PDF, printing as standard table")
		return p.printStackObject(r, objId, logger)
	}

	// Standard stack objects
	return p.printStackObject(r, objId, logger)
}

// Print multiple objects
func (p *PdfReportPrinter) printObjects(r Report, logger *zerolog.Logger) *util.Result {
	if len(r.TargetIDs) < 1 {
		logger.Warn().Msg("no objects to print")
		return nil
	}

	for i, objId := range r.TargetIDs {
		objLogger := logger.With().Int("object", i).Str("id", objId).Logger()
		if res := p.printObject(r, objId, &objLogger); res != nil {
			objLogger.Err(res).Msg("printObject failed")
			return res.With("printObject")
		}

		// Add page break between objects (except for last one)
		if i < len(r.TargetIDs)-1 {
			p.pdf.AddPage()
		}
	}

	return nil
}

// Print entire sheet
func (p *PdfReportPrinter) printSheet(r Report, logger *zerolog.Logger) *util.Result {
	if len(r.TargetIDs) != 1 {
		return util.MsgError("printSheet", "exactly one sheet ID required")
	}

	sheetId := r.TargetIDs[0]
	logger.Info().Msgf("printing sheet: %s", sheetId)

	sheet, err := r.Doc.GetObject(engine.ConnCtx, sheetId)
	if err != nil {
		return util.Error("GetSheet", err)
	}

	children, err := sheet.GetChildInfos(engine.ConnCtx)
	if err != nil {
		return util.Error("GetChildInfos", err)
	}

	for i, child := range children {
		childLogger := logger.With().Int("child", i).Str("id", child.Id).Logger()
		if res := p.printObject(r, child.Id, &childLogger); res != nil {
			childLogger.Err(res).Msg("printObject failed")
			return res.With("printObject")
		}

		// Add page break between sheet children (except for last one)
		if i < len(children)-1 {
			p.pdf.AddPage()
		}
	}

	return nil
}

// Print stack object (standard table)
func (p *PdfReportPrinter) printStackObject(r Report, objId string, logger *zerolog.Logger) *util.Result {
	logger.Info().Msgf("printing stack object: %s", objId)

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
		logger.Warn().Msg("no hypercube, skipping")
		return nil
	}

	logger.Info().Msgf("Hypercube size: %d x %d", objLayout.HyperCube.Size.Cx, objLayout.HyperCube.Size.Cy)

	// Check for hypercube errors
	if objLayout.HyperCube.Error != nil {
		cubeErr := objLayout.HyperCube.Error
		return util.MsgError("HyperCubeError", fmt.Sprintf("code: %d, %s", cubeErr.ErrorCode, cubeErr.ExtendedMessage))
	}

	// Calculate column widths
	p.calculateColumnWidths(objLayout)

	// Print header
	if res := p.printObjectHeader(objLayout, r, logger); res != nil {
		return res.With("printObjectHeader")
	}

	// Get data
	dataPages, res := engine.GetHyperCubeData(obj, *objLayout.HyperCube.Size)
	if res != nil {
		return res.With("GetHyperCubeData")
	}

	logger.Info().Msgf("Got %d data pages", len(dataPages))

	// Reorganize pages into a row-based structure to handle column pagination
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

			for _ci, cell := range rowCells {
				cubeColIx := page.Area.Left + _ci
				rowData[absoluteRowIdx][cubeColIx] = cell
			}
		}
	}

	// Print data rows in order
	for rowIdx := 0; rowIdx <= maxRow; rowIdx++ {
		cellsByCol, exists := rowData[rowIdx]
		if !exists {
			continue
		}

		// Check if we need a new page
		if p.pdf.GetY() > p.pageHeight-PDF_MARGIN_TOP-PDF_LINE_HEIGHT {
			p.pdf.AddPage()
			// Re-print header on new page
			if res := p.printObjectHeader(objLayout, r, logger); res != nil {
				return res.With("printObjectHeader")
			}
		}

		// Create a temporary array to hold cells in the correct display order
		reorderedCells := make([]*enigma.NxCell, len(p.displayColWidths))

		for cubeColIx, cell := range cellsByCol {
			// Look up the display column index for this cube column
			repIdx, ok := p.cube2report[cubeColIx]
			if !ok {
				repIdx = cubeColIx // Fallback: use data position if not in map
				logger.Warn().Msgf("No mapping for cubeColIx %d, using cubeColIx", cubeColIx)
			}
			if repIdx < len(reorderedCells) {
				reorderedCells[repIdx] = cell
			} else {
				logger.Warn().Msgf("repIdx %d >= reorderedCells size %d, cell dropped", repIdx, len(reorderedCells))
			}
		}

		// Print cells in display order
		for ci, cell := range reorderedCells {
			if ci >= len(p.displayColWidths) {
				break
			}

			// Get ColumnInfo for this display column
			colInfo := p.displayColInfo[ci]

			if cell == nil {
				// Print empty cell to maintain column alignment
				p.pdf.CellFormat(p.displayColWidths[ci], PDF_LINE_HEIGHT, "", "1", 0, "", false, 0, "")
			} else {
				cellLogger := logger.With().Int("row", rowIdx).Int("col", ci).Logger()
				p.printCell(cell, p.displayColWidths[ci], colInfo, &cellLogger)
			}
		}
		p.pdf.Ln(-1)
		p.printedRows++
	}

	return nil
}

// Main print method
func (p *PdfReportPrinter) Print(r Report) *util.Result {
	if !r.IsValid() {
		return util.MsgError("Print", "invalid report")
	}

	rResult, res := NewReportResult(r)
	if res != nil {
		return res.With("NewReportResult")
	}
	p.ReportResults[util.MaybeNil(r.ID)] = rResult
	logger := rResult.Logger.With().Str("report", *r.ID).Logger()

	if r.Doc == nil {
		return util.MsgError("CheckDoc", "doc is not opened")
	}

	// Determine PDF orientation and page dimensions
	orientation := "L" // default landscape
	if r.OutputPDFOrientation != nil && *r.OutputPDFOrientation == PDF_ORIENTATION_PORTRAIT {
		orientation = "P"
		p.pageWidth = 210.0  // A4 portrait width in mm
		p.pageHeight = 297.0 // A4 portrait height in mm
	} else {
		p.pageWidth = 297.0  // A4 landscape width in mm
		p.pageHeight = 210.0 // A4 landscape height in mm
	}

	// Initialize PDF
	p.pdf = gofpdf.New(orientation, "mm", "A4", "")
	p.pdf.SetMargins(PDF_MARGIN_LEFT, PDF_MARGIN_TOP, PDF_MARGIN_RIGHT)
	p.pdf.SetAutoPageBreak(true, PDF_MARGIN_TOP)
	p.pdf.AddPage()
	p.pdf.SetFont("Arial", "", PDF_FONT_SIZE)

	// Print sheet header (current selection + custom headers)
	if res := p.printSheetHeader(r, r.Doc, &logger); res != nil {
		logger.Err(res).Msg("printSheetHeader failed")
		return res.With("printSheetHeader")
	}

	// Normalize target
	r.Target = strings.ToLower(r.Target)

	// Route to appropriate print method based on target
	if r.Target == TARGET_OBJECTS {
		if len(r.TargetIDs) < 1 {
			return util.MsgError("Print", "no target objects specified")
		}
		res = p.printObjects(r, &logger)
	} else if r.Target == TARGET_SHEET {
		res = p.printSheet(r, &logger)
	} else {
		return util.MsgError("Print", fmt.Sprintf("PDF printer does not support target '%s'", r.Target))
	}

	if res != nil {
		logger.Err(res).Msg("print failed")
		return res
	}

	// Save PDF
	err := p.pdf.OutputFileAndClose(*rResult.ReportFile)
	if err != nil {
		return util.Error("SavePDF", err)
	}

	rResult.PrintedRows = p.printedRows
	logger.Info().Msgf("PDF saved to %s (%d rows)", *rResult.ReportFile, p.printedRows)

	return nil
}

func (p PdfReportPrinter) GetReportResult(id string) (*ReportResult, *util.Result) {
	result, ok := p.ReportResults[id]
	if !ok {
		return nil, util.MsgError("ReportFiles", "report id doesn't exist")
	}
	return result, nil
}
