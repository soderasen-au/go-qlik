//go:build windows

package report

import (
	"fmt"
	"path/filepath"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/rs/zerolog"
	"github.com/soderasen-au/go-common/util"
)

// Excel constants for COM automation
const (
	// PageSetup constants
	xlPaperA4       = 9  // A4 paper (210mm x 297mm)
	xlLandscape     = 2  // Landscape orientation
	xlPortrait      = 1  // Portrait orientation

	// Export format
	xlTypePDF       = 0  // PDF file format

	// Quality
	xlQualityStandard = 0
)

// ExcelToPDFWinConfig holds configuration for Excel-to-PDF conversion
type ExcelToPDFWinConfig struct {
	// Input/Output
	InputExcelPath   string  // Path to input Excel file
	OutputPDFPath    string  // Path to output PDF file (for single sheet) or directory (for multiple sheets)
	Password         string  // Password for protected Excel files (optional)

	// Sheet selection
	SheetNames       []string // Specific sheet names to export (empty = all sheets)
	SheetIndices     []int    // Specific sheet indices to export (1-based, empty = all sheets)

	// Page setup
	PaperSize        int     // Excel paper size constant (default: xlPaperA4)
	Orientation      int     // Excel orientation constant (default: xlLandscape)
	FitToWidth       int     // Fit to N pages wide (0 = no fit, 1 = fit to 1 page wide)
	FitToHeight      int     // Fit to N pages tall (0 = no fit)

	// Margins (in inches, Excel default: 0.75)
	LeftMargin       float64 // Left margin in inches
	RightMargin      float64 // Right margin in inches
	TopMargin        float64 // Top margin in inches
	BottomMargin     float64 // Bottom margin in inches

	// Print area (optional)
	PrintArea        string  // Excel range notation (e.g., "A1:Z100"), empty = use default

	// Export options
	ExportMultiplePDFs bool  // If true, export one PDF per sheet (OutputPDFPath becomes directory)
	IncludeDocProperties bool // Include Excel document properties in PDF
	OpenAfterPublish   bool  // Open PDF after export (default: false)

	// Logger
	Logger           *zerolog.Logger
}

// SetDefaults applies sensible defaults to unset config values
func (c *ExcelToPDFWinConfig) SetDefaults() {
	if c.PaperSize == 0 {
		c.PaperSize = xlPaperA4
	}
	if c.Orientation == 0 {
		c.Orientation = xlLandscape
	}
	if c.FitToWidth == 0 {
		c.FitToWidth = 1 // Fit to 1 page wide by default
	}
	if c.LeftMargin == 0 {
		c.LeftMargin = 0.5 // Smaller than Excel default
	}
	if c.RightMargin == 0 {
		c.RightMargin = 0.5
	}
	if c.TopMargin == 0 {
		c.TopMargin = 0.75
	}
	if c.BottomMargin == 0 {
		c.BottomMargin = 0.75
	}
}

// Validate checks configuration validity
func (c *ExcelToPDFWinConfig) Validate() *util.Result {
	if c.InputExcelPath == "" {
		return util.MsgError("Validate", "InputExcelPath is required")
	}
	if c.OutputPDFPath == "" {
		return util.MsgError("Validate", "OutputPDFPath is required")
	}

	// Check for conflicting sheet selection
	if len(c.SheetNames) > 0 && len(c.SheetIndices) > 0 {
		return util.MsgError("Validate", "cannot specify both SheetNames and SheetIndices")
	}

	return nil
}

// ExcelToPDFWin handles Windows-specific Excel-to-PDF conversion using COM automation
type ExcelToPDFWin struct {
	config   ExcelToPDFWinConfig
	logger   *zerolog.Logger
	excel    *ole.IDispatch
	workbook *ole.IDispatch
}

// NewExcelToPDFWin creates a new Excel-to-PDF converter with the given configuration
func NewExcelToPDFWin(config ExcelToPDFWinConfig) (*ExcelToPDFWin, *util.Result) {
	config.SetDefaults()

	if res := config.Validate(); res != nil {
		return nil, res.With("Validate")
	}

	logger := config.Logger
	if logger == nil {
		defaultLogger := zerolog.Nop()
		logger = &defaultLogger
	}

	return &ExcelToPDFWin{
		config: config,
		logger: logger,
	}, nil
}

// Convert performs the Excel-to-PDF conversion
func (e *ExcelToPDFWin) Convert() *util.Result {
	logger := e.logger.With().Str("input", e.config.InputExcelPath).Str("output", e.config.OutputPDFPath).Logger()
	logger.Info().Msg("starting Excel-to-PDF conversion")

	// Initialize COM
	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		logger.Err(err).Msg("CoInitializeEx failed")
		return util.Error("CoInitializeEx", err)
	}
	defer ole.CoUninitialize()

	// Open Excel
	if res := e.openExcel(); res != nil {
		return res.With("openExcel")
	}
	defer e.cleanup()

	// Open workbook
	if res := e.openWorkbook(); res != nil {
		return res.With("openWorkbook")
	}

	// Get sheets to export
	sheets, res := e.getSheetsToExport()
	if res != nil {
		return res.With("getSheetsToExport")
	}

	logger.Info().Msgf("found %d sheet(s) to export", len(sheets))

	// Export based on configuration
	if e.config.ExportMultiplePDFs {
		return e.exportMultiplePDFs(sheets)
	}

	return e.exportSinglePDF(sheets)
}

// openExcel initializes Excel.Application COM object
func (e *ExcelToPDFWin) openExcel() *util.Result {
	e.logger.Debug().Msg("creating Excel.Application COM object")

	unknown, err := oleutil.CreateObject("Excel.Application")
	if err != nil {
		e.logger.Err(err).Msg("CreateObject failed - ensure MS Excel is installed")
		return util.Error("CreateObject", err)
	}

	excel, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		unknown.Release()
		e.logger.Err(err).Msg("QueryInterface failed")
		return util.Error("QueryInterface", err)
	}

	e.excel = excel

	// Set Excel to invisible (no UI)
	if _, err := oleutil.PutProperty(e.excel, "Visible", false); err != nil {
		e.logger.Err(err).Msg("failed to set Excel.Visible")
		return util.Error("SetVisible", err)
	}

	// Disable alerts (prompts)
	if _, err := oleutil.PutProperty(e.excel, "DisplayAlerts", false); err != nil {
		e.logger.Err(err).Msg("failed to set Excel.DisplayAlerts")
		return util.Error("SetDisplayAlerts", err)
	}

	e.logger.Debug().Msg("Excel.Application initialized")
	return nil
}

// openWorkbook opens the Excel workbook with optional password
func (e *ExcelToPDFWin) openWorkbook() *util.Result {
	e.logger.Debug().Msgf("opening workbook: %s", e.config.InputExcelPath)

	// Get absolute path
	absPath, err := filepath.Abs(e.config.InputExcelPath)
	if err != nil {
		e.logger.Err(err).Msg("filepath.Abs failed")
		return util.Error("filepath.Abs", err)
	}

	workbooks := oleutil.MustGetProperty(e.excel, "Workbooks").ToIDispatch()
	defer workbooks.Release()

	// Open workbook with password support
	// Workbooks.Open(FileName, UpdateLinks, ReadOnly, Format, Password, ...)
	var workbook *ole.VARIANT
	if e.config.Password != "" {
		e.logger.Debug().Msg("opening password-protected workbook")
		workbook, err = oleutil.CallMethod(workbooks, "Open",
			absPath,           // FileName
			0,                 // UpdateLinks
			false,             // ReadOnly
			nil,               // Format
			e.config.Password, // Password
		)
	} else {
		workbook, err = oleutil.CallMethod(workbooks, "Open", absPath)
	}

	if err != nil {
		e.logger.Err(err).Msg("Workbooks.Open failed")
		return util.Error("Workbooks.Open", err)
	}

	e.workbook = workbook.ToIDispatch()
	e.logger.Debug().Msg("workbook opened successfully")
	return nil
}

// sheetInfo holds sheet metadata
type sheetInfo struct {
	Name  string
	Index int // 1-based
	Sheet *ole.IDispatch
}

// getSheetsToExport returns list of sheets to export based on config
func (e *ExcelToPDFWin) getSheetsToExport() ([]sheetInfo, *util.Result) {
	sheets := oleutil.MustGetProperty(e.workbook, "Sheets").ToIDispatch()
	defer sheets.Release()

	sheetCount := int(oleutil.MustGetProperty(sheets, "Count").Val)
	e.logger.Debug().Msgf("workbook contains %d sheets", sheetCount)

	var result []sheetInfo

	// If specific sheet names requested
	if len(e.config.SheetNames) > 0 {
		for _, name := range e.config.SheetNames {
			sheet, err := oleutil.CallMethod(sheets, "Item", name)
			if err != nil {
				e.logger.Warn().Msgf("sheet '%s' not found, skipping", name)
				continue
			}
			index := int(oleutil.MustGetProperty(sheet.ToIDispatch(), "Index").Val)
			result = append(result, sheetInfo{
				Name:  name,
				Index: index,
				Sheet: sheet.ToIDispatch(),
			})
		}
		return result, nil
	}

	// If specific sheet indices requested
	if len(e.config.SheetIndices) > 0 {
		for _, idx := range e.config.SheetIndices {
			if idx < 1 || idx > sheetCount {
				e.logger.Warn().Msgf("sheet index %d out of range [1-%d], skipping", idx, sheetCount)
				continue
			}
			sheet := oleutil.MustGetProperty(sheets, "Item", idx).ToIDispatch()
			name := oleutil.MustGetProperty(sheet, "Name").ToString()
			result = append(result, sheetInfo{
				Name:  name,
				Index: idx,
				Sheet: sheet,
			})
		}
		return result, nil
	}

	// Export all sheets
	for i := 1; i <= sheetCount; i++ {
		sheet := oleutil.MustGetProperty(sheets, "Item", i).ToIDispatch()
		name := oleutil.MustGetProperty(sheet, "Name").ToString()
		result = append(result, sheetInfo{
			Name:  name,
			Index: i,
			Sheet: sheet,
		})
	}

	return result, nil
}

// setupPageSettings applies page setup (A4, landscape, margins, etc.) to a sheet
func (e *ExcelToPDFWin) setupPageSettings(sheet *ole.IDispatch) *util.Result {
	pageSetup := oleutil.MustGetProperty(sheet, "PageSetup").ToIDispatch()
	defer pageSetup.Release()

	e.logger.Debug().Msg("applying page setup")

	// Set paper size
	if _, err := oleutil.PutProperty(pageSetup, "PaperSize", e.config.PaperSize); err != nil {
		e.logger.Err(err).Msg("failed to set PaperSize")
		return util.Error("SetPaperSize", err)
	}

	// Set orientation
	if _, err := oleutil.PutProperty(pageSetup, "Orientation", e.config.Orientation); err != nil {
		e.logger.Err(err).Msg("failed to set Orientation")
		return util.Error("SetOrientation", err)
	}

	// Set margins (convert to points: 1 inch = 72 points, but Excel uses Application.InchesToPoints)
	application := oleutil.MustGetProperty(sheet, "Application").ToIDispatch()
	defer application.Release()

	leftPoints := oleutil.MustCallMethod(application, "InchesToPoints", e.config.LeftMargin).Val
	if _, err := oleutil.PutProperty(pageSetup, "LeftMargin", leftPoints); err != nil {
		e.logger.Err(err).Msg("failed to set LeftMargin")
		return util.Error("SetLeftMargin", err)
	}

	rightPoints := oleutil.MustCallMethod(application, "InchesToPoints", e.config.RightMargin).Val
	if _, err := oleutil.PutProperty(pageSetup, "RightMargin", rightPoints); err != nil {
		e.logger.Err(err).Msg("failed to set RightMargin")
		return util.Error("SetRightMargin", err)
	}

	topPoints := oleutil.MustCallMethod(application, "InchesToPoints", e.config.TopMargin).Val
	if _, err := oleutil.PutProperty(pageSetup, "TopMargin", topPoints); err != nil {
		e.logger.Err(err).Msg("failed to set TopMargin")
		return util.Error("SetTopMargin", err)
	}

	bottomPoints := oleutil.MustCallMethod(application, "InchesToPoints", e.config.BottomMargin).Val
	if _, err := oleutil.PutProperty(pageSetup, "BottomMargin", bottomPoints); err != nil {
		e.logger.Err(err).Msg("failed to set BottomMargin")
		return util.Error("SetBottomMargin", err)
	}

	// Set fit-to-width/height
	if e.config.FitToWidth > 0 {
		// Disable Zoom when using FitToPages
		if _, err := oleutil.PutProperty(pageSetup, "Zoom", false); err != nil {
			e.logger.Err(err).Msg("failed to disable Zoom")
			return util.Error("DisableZoom", err)
		}

		if _, err := oleutil.PutProperty(pageSetup, "FitToPagesWide", e.config.FitToWidth); err != nil {
			e.logger.Err(err).Msg("failed to set FitToPagesWide")
			return util.Error("SetFitToPagesWide", err)
		}

		if e.config.FitToHeight > 0 {
			if _, err := oleutil.PutProperty(pageSetup, "FitToPagesTall", e.config.FitToHeight); err != nil {
				e.logger.Err(err).Msg("failed to set FitToPagesTall")
				return util.Error("SetFitToPagesTall", err)
			}
		} else {
			// Allow unlimited height
			if _, err := oleutil.PutProperty(pageSetup, "FitToPagesTall", false); err != nil {
				e.logger.Err(err).Msg("failed to disable FitToPagesTall")
				return util.Error("DisableFitToPagesTall", err)
			}
		}
	}

	// Set print area if specified
	if e.config.PrintArea != "" {
		e.logger.Debug().Msgf("setting print area: %s", e.config.PrintArea)
		if _, err := oleutil.PutProperty(pageSetup, "PrintArea", e.config.PrintArea); err != nil {
			e.logger.Err(err).Msg("failed to set PrintArea")
			return util.Error("SetPrintArea", err)
		}
	}

	e.logger.Debug().Msg("page setup complete")
	return nil
}

// exportSinglePDF exports selected sheets to a single PDF
func (e *ExcelToPDFWin) exportSinglePDF(sheets []sheetInfo) *util.Result {
	logger := e.logger.With().Str("mode", "single-pdf").Logger()
	logger.Info().Msgf("exporting %d sheet(s) to single PDF", len(sheets))

	// Apply page setup to all sheets
	for i, sheet := range sheets {
		logger.Debug().Msgf("configuring sheet %d/%d: %s", i+1, len(sheets), sheet.Name)
		if res := e.setupPageSettings(sheet.Sheet); res != nil {
			return res.With(fmt.Sprintf("setupPageSettings[%s]", sheet.Name))
		}
	}

	// Get absolute output path
	absOutputPath, err := filepath.Abs(e.config.OutputPDFPath)
	if err != nil {
		logger.Err(err).Msg("filepath.Abs failed")
		return util.Error("filepath.Abs", err)
	}

	// Export workbook to PDF
	logger.Info().Msgf("exporting to: %s", absOutputPath)
	if _, err := oleutil.CallMethod(e.workbook, "ExportAsFixedFormat",
		xlTypePDF,                      // Type
		absOutputPath,                  // Filename
		xlQualityStandard,              // Quality
		e.config.IncludeDocProperties,  // IncludeDocProperties
		false,                          // IgnorePrintAreas
		nil,                            // From (nil = all)
		nil,                            // To (nil = all)
		e.config.OpenAfterPublish,      // OpenAfterPublish
	); err != nil {
		logger.Err(err).Msg("ExportAsFixedFormat failed")
		return util.Error("ExportAsFixedFormat", err)
	}

	logger.Info().Msg("export complete")
	return nil
}

// exportMultiplePDFs exports each sheet to a separate PDF file
func (e *ExcelToPDFWin) exportMultiplePDFs(sheets []sheetInfo) *util.Result {
	logger := e.logger.With().Str("mode", "multi-pdf").Logger()
	logger.Info().Msgf("exporting %d sheet(s) to separate PDF files", len(sheets))

	// OutputPDFPath is treated as directory
	outputDir := e.config.OutputPDFPath

	for i, sheet := range sheets {
		sheetLogger := logger.With().Str("sheet", sheet.Name).Int("index", sheet.Index).Logger()
		sheetLogger.Debug().Msgf("processing sheet %d/%d", i+1, len(sheets))

		// Apply page setup
		if res := e.setupPageSettings(sheet.Sheet); res != nil {
			return res.With(fmt.Sprintf("setupPageSettings[%s]", sheet.Name))
		}

		// Activate sheet (required for export)
		if _, err := oleutil.CallMethod(sheet.Sheet, "Activate"); err != nil {
			sheetLogger.Err(err).Msg("Activate failed")
			return util.Error("Activate", err)
		}

		// Generate output filename
		// Sanitize sheet name for filesystem
		safeName := sanitizeFilename(sheet.Name)
		outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.pdf", safeName))
		absOutputPath, err := filepath.Abs(outputPath)
		if err != nil {
			sheetLogger.Err(err).Msg("filepath.Abs failed")
			return util.Error("filepath.Abs", err)
		}

		// Export sheet to PDF
		sheetLogger.Info().Msgf("exporting to: %s", absOutputPath)
		if _, err := oleutil.CallMethod(sheet.Sheet, "ExportAsFixedFormat",
			xlTypePDF,                      // Type
			absOutputPath,                  // Filename
			xlQualityStandard,              // Quality
			e.config.IncludeDocProperties,  // IncludeDocProperties
			false,                          // IgnorePrintAreas
			nil,                            // From
			nil,                            // To
			false,                          // OpenAfterPublish (never open intermediate PDFs)
		); err != nil {
			sheetLogger.Err(err).Msg("ExportAsFixedFormat failed")
			return util.Error("ExportAsFixedFormat", err)
		}

		sheetLogger.Debug().Msg("sheet export complete")
	}

	logger.Info().Msg("all sheets exported successfully")
	return nil
}

// cleanup releases COM objects in reverse order
func (e *ExcelToPDFWin) cleanup() {
	logger := e.logger.With().Str("stage", "cleanup").Logger()
	logger.Debug().Msg("starting cleanup")

	// Close workbook without saving
	if e.workbook != nil {
		logger.Debug().Msg("closing workbook")
		if _, err := oleutil.CallMethod(e.workbook, "Close", false); err != nil {
			logger.Warn().Err(err).Msg("Workbook.Close failed")
		}
		e.workbook.Release()
		e.workbook = nil
	}

	// Quit Excel
	if e.excel != nil {
		logger.Debug().Msg("quitting Excel application")
		if _, err := oleutil.CallMethod(e.excel, "Quit"); err != nil {
			logger.Warn().Err(err).Msg("Excel.Quit failed")
		}
		e.excel.Release()
		e.excel = nil
	}

	logger.Debug().Msg("cleanup complete")
}

// sanitizeFilename removes characters that are invalid in Windows filenames
func sanitizeFilename(name string) string {
	// Windows forbidden characters: < > : " / \ | ? *
	replacer := map[rune]rune{
		'<': '_', '>': '_', ':': '_', '"': '_', '/': '_',
		'\\': '_', '|': '_', '?': '_', '*': '_',
	}

	runes := []rune(name)
	for i, r := range runes {
		if replacement, found := replacer[r]; found {
			runes[i] = replacement
		}
	}

	return string(runes)
}
