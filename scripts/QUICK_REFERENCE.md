# PDF Comparison - Quick Reference

## One-Line Commands

```bash
# Run complete comparison
make compare-pdf

# View text report
cat test-reports/comparison-report.txt

# View HTML report
open test-reports/comparison-report.html

# Check truncation count
cat test-reports/comparison-data.json | jq '.pdf_analysis.total_truncations'

# Clean up all reports
make clean
```

## Common Tasks

### Generate Reports Only (No Comparison)
```bash
make test-pdf
```

### Run Scripts Manually
```bash
# Bash comparison
bash scripts/compare-reports.sh

# Python analysis
python3 scripts/extract-pdf-issues.py
```

### Debug Specific Issues
```bash
# View truncated entries
grep '\.\.\.' test-reports/pdf-extracted-text.txt

# Compare CSV files manually
diff test-reports/TestReport_reference.csv test-reports/TestReport.csv

# Extract specific field from JSON
cat test-reports/comparison-data.json | jq '.pdf_analysis.patterns'
```

### CI/CD Integration
```yaml
# Add to .github/workflows/test.yml
- name: Compare PDF reports
  run: make compare-pdf
  continue-on-error: true

- name: Upload results
  uses: actions/upload-artifact@v3
  with:
    name: comparison-report
    path: test-reports/comparison-report.html
```

## File Locations

| What | Where |
|------|-------|
| Scripts | `scripts/` |
| Reports | `test-reports/` |
| Reference | `test/pdf/TestReport.ods` |
| PDF code | `report/pdf.go` |
| Excel code | `report/excel.go` |

## Current Issue

**Problem:** 1,868 text truncations in PDF
**Location:** `report/pdf.go:306-310`
**Fix:** Replace `colWidth / 2.0` with `p.pdf.GetStringWidth(cellText)`

## Exit Codes

- `0` - All checks passed
- `1` - Issues detected (check reports)

## Dependencies

Install if missing:
```bash
# Ubuntu/Debian
sudo apt-get install libreoffice-calc poppler-utils

# macOS
brew install --cask libreoffice
brew install poppler
```

## Support

- Full docs: `scripts/README.md`
- Implementation: `test-reports/COMPARISON_SUMMARY.md`
- Project guide: `CLAUDE.md`
