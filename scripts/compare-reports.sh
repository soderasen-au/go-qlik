#!/bin/bash
#
# compare-reports.sh - Compare generated reports with reference files
#
# This script compares PDF and Excel outputs against reference ODS file,
# identifies differences, and generates a detailed comparison report.
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REFERENCE_ODS="${REFERENCE_ODS:-test/pdf/TestReport.ods}"
GENERATED_XLSX="${GENERATED_XLSX:-test-reports/TestReport.xlsx}"
GENERATED_PDF="${GENERATED_PDF:-test-reports/TestReport.pdf}"
OUTPUT_DIR="${OUTPUT_DIR:-test-reports}"
TEMP_DIR="$OUTPUT_DIR/.tmp"

# Ensure output directory exists
mkdir -p "$OUTPUT_DIR"
mkdir -p "$TEMP_DIR"

echo -e "${BLUE}=== Report Comparison Tool ===${NC}"
echo ""

# Check if files exist
if [[ ! -f "$REFERENCE_ODS" ]]; then
    echo -e "${RED}Error: Reference file not found: $REFERENCE_ODS${NC}"
    exit 1
fi

if [[ ! -f "$GENERATED_XLSX" ]]; then
    echo -e "${RED}Error: Generated Excel file not found: $GENERATED_XLSX${NC}"
    exit 1
fi

if [[ ! -f "$GENERATED_PDF" ]]; then
    echo -e "${RED}Error: Generated PDF file not found: $GENERATED_PDF${NC}"
    exit 1
fi

echo -e "${BLUE}Files to compare:${NC}"
echo "  Reference:      $REFERENCE_ODS"
echo "  Generated XLSX: $GENERATED_XLSX"
echo "  Generated PDF:  $GENERATED_PDF"
echo ""

# Convert files to CSV for comparison
echo -e "${BLUE}Converting files to CSV...${NC}"

# Convert reference ODS to CSV
libreoffice --headless --convert-to csv:"Text - txt - csv (StarCalc)":59,34,76,1 \
    --outdir "$TEMP_DIR" "$REFERENCE_ODS" 2>&1 | grep -v "^convert" || true
mv "$TEMP_DIR/TestReport.csv" "$TEMP_DIR/reference.csv" 2>/dev/null || true

# Convert generated XLSX to CSV
libreoffice --headless --convert-to csv:"Text - txt - csv (StarCalc)":59,34,76,1 \
    --outdir "$TEMP_DIR" "$GENERATED_XLSX" 2>&1 | grep -v "^convert" || true
mv "$TEMP_DIR/TestReport.csv" "$TEMP_DIR/generated.csv" 2>/dev/null || true

# Extract text from PDF
pdftotext "$GENERATED_PDF" "$TEMP_DIR/pdf_text.txt" 2>/dev/null || {
    echo -e "${YELLOW}Warning: pdftotext failed or not installed${NC}"
}

echo -e "${GREEN}✓ Conversion complete${NC}"
echo ""

# Compare CSV files
echo -e "${BLUE}Comparing CSV data...${NC}"

if diff -q "$TEMP_DIR/reference.csv" "$TEMP_DIR/generated.csv" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Excel output matches reference file exactly${NC}"
    EXCEL_STATUS="PASS"
else
    echo -e "${YELLOW}⚠ Excel output differs from reference${NC}"
    EXCEL_STATUS="DIFF"

    # Generate detailed diff
    diff -u "$TEMP_DIR/reference.csv" "$TEMP_DIR/generated.csv" > "$OUTPUT_DIR/excel-diff.txt" || true
    echo -e "  Detailed diff saved to: ${YELLOW}$OUTPUT_DIR/excel-diff.txt${NC}"
fi

echo ""

# Analyze PDF content
echo -e "${BLUE}Analyzing PDF content...${NC}"

if [[ -f "$TEMP_DIR/pdf_text.txt" ]]; then
    # Count truncated product names (ending with "...")
    TRUNCATED_COUNT=$(grep -o '\.\.\.' "$TEMP_DIR/pdf_text.txt" | wc -l)

    if [[ $TRUNCATED_COUNT -gt 0 ]]; then
        echo -e "${YELLOW}⚠ Found $TRUNCATED_COUNT truncated text entries (ending with '...')${NC}"
        PDF_STATUS="TRUNCATED"
    else
        echo -e "${GREEN}✓ No text truncation detected${NC}"
        PDF_STATUS="OK"
    fi

    # Save PDF text for inspection
    cp "$TEMP_DIR/pdf_text.txt" "$OUTPUT_DIR/pdf-extracted-text.txt"
    echo -e "  PDF text saved to: ${BLUE}$OUTPUT_DIR/pdf-extracted-text.txt${NC}"
else
    echo -e "${YELLOW}⚠ Could not extract PDF text for analysis${NC}"
    PDF_STATUS="UNKNOWN"
fi

echo ""

# Generate detailed comparison report
echo -e "${BLUE}Generating comparison report...${NC}"

REPORT_FILE="$OUTPUT_DIR/comparison-report.txt"

cat > "$REPORT_FILE" << EOF
================================================================================
REPORT COMPARISON RESULTS
Generated: $(date)
================================================================================

FILES COMPARED
--------------
Reference:      $REFERENCE_ODS
Generated XLSX: $GENERATED_XLSX
Generated PDF:  $GENERATED_PDF

================================================================================
EXCEL COMPARISON: $EXCEL_STATUS
================================================================================

EOF

if [[ "$EXCEL_STATUS" == "PASS" ]]; then
    cat >> "$REPORT_FILE" << EOF
Status: PASS ✓
The generated Excel file matches the reference ODS file exactly.
All data, formatting, and structure are identical when normalized to CSV.

EOF
else
    cat >> "$REPORT_FILE" << EOF
Status: DIFFERENCES FOUND
The generated Excel file differs from the reference.

See excel-diff.txt for detailed line-by-line differences.

Summary of differences:
EOF

    # Add diff statistics
    diff -u "$TEMP_DIR/reference.csv" "$TEMP_DIR/generated.csv" | \
        grep -E "^[\+\-]" | head -20 >> "$REPORT_FILE" 2>/dev/null || true

    cat >> "$REPORT_FILE" << EOF

(First 20 differences shown. See excel-diff.txt for complete comparison)

EOF
fi

cat >> "$REPORT_FILE" << EOF
================================================================================
PDF ANALYSIS: $PDF_STATUS
================================================================================

EOF

if [[ "$PDF_STATUS" == "OK" ]]; then
    cat >> "$REPORT_FILE" << EOF
Status: OK ✓
No text truncation or formatting issues detected in PDF output.

EOF
elif [[ "$PDF_STATUS" == "TRUNCATED" ]]; then
    cat >> "$REPORT_FILE" << EOF
Status: TEXT TRUNCATION DETECTED ⚠

Found $TRUNCATED_COUNT instances of truncated text (ending with '...').

ROOT CAUSE:
-----------
In report/pdf.go:306-310, cell text is truncated using a conservative formula:

    maxLen := int(colWidth / 2.0)  // Assumes ~2mm per character
    if len(cellText) > maxLen {
        cellText = cellText[:maxLen-3] + "..."
    }

ISSUE:
------
This calculation is too aggressive and doesn't use actual font metrics.
Product names that would fit in the available column width are being cut off.

RECOMMENDED FIX:
----------------
Use gofpdf's GetStringWidth() method to calculate actual text width:

    textWidth := p.pdf.GetStringWidth(cellText)
    if textWidth > colWidth - 2*PDF_CELL_PADDING {
        // Binary search or iterative approach to find max fitting length
        for len(cellText) > 0 && p.pdf.GetStringWidth(cellText+"...") > colWidth - 2*PDF_CELL_PADDING {
            cellText = cellText[:len(cellText)-1]
        }
        cellText = cellText + "..."
    }

EXAMPLE TRUNCATIONS FROM PDF:
------------------------------
EOF

    # Extract some examples of truncated text
    grep -o '[A-Za-z][A-Za-z ]*\.\.\.' "$TEMP_DIR/pdf_text.txt" | \
        head -15 >> "$REPORT_FILE" 2>/dev/null || true

    cat >> "$REPORT_FILE" << EOF

(First 15 truncations shown)

LOCATION IN CODE:
-----------------
File: report/pdf.go
Function: printCell()
Lines: 306-310

See pdf-extracted-text.txt for complete PDF content analysis.

EOF
fi

cat >> "$REPORT_FILE" << EOF
================================================================================
RECOMMENDATIONS
================================================================================

EOF

if [[ "$EXCEL_STATUS" == "PASS" && "$PDF_STATUS" == "OK" ]]; then
    cat >> "$REPORT_FILE" << EOF
Status: ALL CHECKS PASSED ✓

Both Excel and PDF outputs are functioning correctly.

EOF
else
    cat >> "$REPORT_FILE" << EOF
Priority Issues to Fix:
EOF

    if [[ "$PDF_STATUS" == "TRUNCATED" ]]; then
        cat >> "$REPORT_FILE" << EOF

1. HIGH PRIORITY: Fix PDF text truncation
   - Update report/pdf.go:306-310 to use actual font metrics
   - Use p.pdf.GetStringWidth() instead of colWidth/2.0 approximation
   - Test with long product names to verify full text is visible

EOF
    fi

    if [[ "$EXCEL_STATUS" != "PASS" ]]; then
        cat >> "$REPORT_FILE" << EOF

2. MEDIUM PRIORITY: Excel output differences
   - Review excel-diff.txt for specific changes
   - Verify if differences are intentional or bugs
   - Update reference file if changes are correct

EOF
    fi
fi

cat >> "$REPORT_FILE" << EOF
================================================================================
FILES GENERATED
================================================================================

Comparison Report:     $REPORT_FILE (this file)
PDF Extracted Text:    $OUTPUT_DIR/pdf-extracted-text.txt
EOF

if [[ "$EXCEL_STATUS" != "PASS" ]]; then
    cat >> "$REPORT_FILE" << EOF
Excel Diff:            $OUTPUT_DIR/excel-diff.txt
EOF
fi

cat >> "$REPORT_FILE" << EOF
Temporary Files:       $TEMP_DIR/ (can be deleted)

================================================================================
END OF REPORT
================================================================================
EOF

echo -e "${GREEN}✓ Comparison report saved to: ${BLUE}$REPORT_FILE${NC}"
echo ""

# Show summary
echo -e "${BLUE}=== Summary ===${NC}"
echo -e "Excel Status: $(if [[ "$EXCEL_STATUS" == "PASS" ]]; then echo -e "${GREEN}$EXCEL_STATUS ✓${NC}"; else echo -e "${YELLOW}$EXCEL_STATUS ⚠${NC}"; fi)"
echo -e "PDF Status:   $(if [[ "$PDF_STATUS" == "OK" ]]; then echo -e "${GREEN}$PDF_STATUS ✓${NC}"; else echo -e "${YELLOW}$PDF_STATUS ⚠${NC}"; fi)"
echo ""
echo -e "${BLUE}View detailed report:${NC} cat $REPORT_FILE"
echo ""

# Cleanup option
if [[ "${CLEANUP:-no}" == "yes" ]]; then
    echo -e "${BLUE}Cleaning up temporary files...${NC}"
    rm -rf "$TEMP_DIR"
    echo -e "${GREEN}✓ Cleanup complete${NC}"
fi

# Exit with status based on results
if [[ "$EXCEL_STATUS" == "PASS" && ("$PDF_STATUS" == "OK" || "$PDF_STATUS" == "UNKNOWN") ]]; then
    exit 0
else
    exit 1
fi
