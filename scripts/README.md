# Report Comparison Scripts

This directory contains automated tools for comparing generated PDF and Excel reports against reference files.

## Overview

The comparison workflow identifies differences between generated reports and reference outputs, with special focus on:

- **Excel accuracy**: Byte-level comparison of data integrity
- **PDF rendering**: Detection of text truncation and formatting issues
- **Detailed analysis**: HTML and JSON reports for visual inspection

## Quick Start

Run the complete comparison workflow:

```bash
make compare-pdf
```

This will:
1. Build the PDF printer tool
2. Generate test reports (Excel + PDF)
3. Compare outputs against reference file `test/pdf/TestReport.ods`
4. Create detailed analysis reports in `test-reports/`

## Scripts

### compare-reports.sh

Main comparison script that:
- Converts ODS, XLSX, and PDF to comparable formats
- Performs byte-level CSV comparison
- Analyzes PDF text for truncation patterns
- Generates comprehensive text report

**Usage:**
```bash
bash scripts/compare-reports.sh
```

**Environment Variables:**
- `REFERENCE_ODS`: Path to reference ODS file (default: `test/pdf/TestReport.ods`)
- `GENERATED_XLSX`: Path to generated Excel (default: `test-reports/TestReport.xlsx`)
- `GENERATED_PDF`: Path to generated PDF (default: `test-reports/TestReport.pdf`)
- `OUTPUT_DIR`: Output directory for reports (default: `test-reports`)
- `CLEANUP`: Set to `yes` to remove temporary files (default: `no`)

**Example:**
```bash
CLEANUP=yes bash scripts/compare-reports.sh
```

### extract-pdf-issues.py

Python script for deep PDF analysis:
- Identifies truncation patterns and frequencies
- Performs CSV line-by-line diff analysis
- Generates interactive HTML report with visual diff
- Exports JSON data for further processing

**Usage:**
```bash
python3 scripts/extract-pdf-issues.py
```

**Requirements:**
- Python 3.6+
- No external dependencies (uses stdlib only)

**Outputs:**
- `test-reports/comparison-report.html`: Visual HTML report
- `test-reports/comparison-data.json`: Machine-readable analysis data

## Output Files

After running `make compare-pdf`, the following files are created in `test-reports/`:

| File | Description |
|------|-------------|
| `comparison-report.txt` | Human-readable text summary with root cause analysis |
| `comparison-report.html` | Interactive HTML report with visual diff |
| `comparison-data.json` | Structured JSON data for programmatic analysis |
| `pdf-extracted-text.txt` | Raw text extracted from PDF for inspection |
| `excel-diff.txt` | Line-by-line CSV diff (only if differences found) |
| `TestReport.xlsx` | Generated Excel report |
| `TestReport.pdf` | Generated PDF report |
| `.tmp/` | Temporary conversion files (can be deleted) |

## Viewing Results

### Text Report
```bash
cat test-reports/comparison-report.txt
```

### HTML Report
```bash
# Linux
xdg-open test-reports/comparison-report.html

# macOS
open test-reports/comparison-report.html

# Or simply open in browser:
firefox test-reports/comparison-report.html
```

### JSON Data
```bash
cat test-reports/comparison-data.json | jq '.pdf_analysis.total_truncations'
```

## Current Findings

### ✅ Excel Output - PASS
The Excel generator works perfectly. Generated XLSX files match the reference ODS file exactly when normalized to CSV format.

### ⚠️ PDF Output - TEXT TRUNCATION DETECTED

**Issue:** 1,868 instances of truncated text in PDF output

**Root Cause:**
```go
// File: report/pdf.go, Lines: 306-310
maxLen := int(colWidth / 2.0)  // Assumes ~2mm per character
if len(cellText) > maxLen {
    cellText = cellText[:maxLen-3] + "..."
}
```

**Problem:**
- Uses approximate formula `colWidth / 2.0` instead of actual font metrics
- Doesn't account for variable-width fonts
- Cuts text too aggressively

**Recommended Fix:**
```go
// Use gofpdf's GetStringWidth() for accurate measurement
textWidth := p.pdf.GetStringWidth(cellText)
availableWidth := colWidth - 2*PDF_CELL_PADDING

if textWidth > availableWidth {
    suffix := "..."
    suffixWidth := p.pdf.GetStringWidth(suffix)

    for len(cellText) > 0 {
        if p.pdf.GetStringWidth(cellText) + suffixWidth <= availableWidth {
            cellText = cellText + suffix
            break
        }
        cellText = cellText[:len(cellText)-1]
    }
}
```

## Integration with CI/CD

The comparison workflow can be integrated into CI pipelines:

```yaml
# Example GitHub Actions workflow
- name: Run PDF comparison
  run: make compare-pdf

- name: Upload comparison results
  uses: actions/upload-artifact@v3
  with:
    name: pdf-comparison-report
    path: |
      test-reports/comparison-report.html
      test-reports/comparison-report.txt
      test-reports/comparison-data.json
```

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make test-pdf` | Generate test reports only (no comparison) |
| `make compare-pdf` | Generate reports and run full comparison |
| `make clean` | Remove all generated files including reports |

## Dependencies

### System Requirements
- **LibreOffice** (headless mode): For ODS/XLSX to CSV conversion
  ```bash
  # Ubuntu/Debian
  sudo apt-get install libreoffice-calc

  # macOS
  brew install --cask libreoffice
  ```

- **pdftotext** (poppler-utils): For PDF text extraction
  ```bash
  # Ubuntu/Debian
  sudo apt-get install poppler-utils

  # macOS
  brew install poppler
  ```

- **Python 3**: For detailed analysis script
  ```bash
  python3 --version  # Should be 3.6+
  ```

## Troubleshooting

### "LibreOffice conversion failed"
Ensure LibreOffice is installed and accessible in PATH:
```bash
which libreoffice
libreoffice --version
```

### "pdftotext not found"
Install poppler-utils (see Dependencies section above)

### "Permission denied" on scripts
Make scripts executable:
```bash
chmod +x scripts/compare-reports.sh
chmod +x scripts/extract-pdf-issues.py
```

### Empty or missing comparison files
Ensure reports are generated before comparison:
```bash
make test-pdf    # Generate reports first
make compare-pdf # Then run comparison
```

## Contributing

When modifying comparison scripts:

1. **Test thoroughly**: Run against known-good reference files
2. **Update documentation**: Keep this README in sync with changes
3. **Preserve backward compatibility**: Don't break existing workflows
4. **Add error handling**: Scripts should fail gracefully

## Future Enhancements

Potential improvements for the comparison workflow:

- [ ] Visual diff of PDF layout and positioning
- [ ] Color comparison for styled cells
- [ ] Font size and style verification
- [ ] Performance metrics (generation time, file size)
- [ ] Automated regression testing on commits
- [ ] Support for multiple reference files
- [ ] Threshold-based pass/fail criteria

## License

These scripts are part of the go-qlik project and follow the same license terms.
