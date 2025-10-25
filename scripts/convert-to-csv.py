#!/usr/bin/env python3
"""
convert-to-csv.py - Convert ODS/XLSX files to CSV

Converts spreadsheet files to CSV format without requiring LibreOffice.
Uses lightweight Python libraries: openpyxl for Excel, odfpy for ODS.
"""

import sys
import csv
from pathlib import Path

def convert_xlsx_to_csv(xlsx_file: Path, csv_file: Path) -> bool:
    """Convert XLSX to CSV using openpyxl."""
    try:
        import openpyxl
    except ImportError:
        print("Error: openpyxl not installed. Run: pip3 install openpyxl", file=sys.stderr)
        return False

    try:
        wb = openpyxl.load_workbook(xlsx_file, data_only=True)
        ws = wb.active

        with open(csv_file, 'w', newline='', encoding='utf-8') as f:
            writer = csv.writer(f)
            for row in ws.iter_rows(values_only=True):
                # Convert None to empty string
                row_data = [str(cell) if cell is not None else '' for cell in row]
                writer.writerow(row_data)

        return True
    except Exception as e:
        print(f"Error converting XLSX to CSV: {e}", file=sys.stderr)
        return False


def convert_ods_to_csv(ods_file: Path, csv_file: Path) -> bool:
    """Convert ODS to CSV using odfpy and pandas."""
    try:
        import pandas as pd
    except ImportError:
        print("Error: pandas not installed. Run: pip3 install pandas", file=sys.stderr)
        return False

    try:
        # pandas can read ODS files if odfpy is installed
        df = pd.read_excel(ods_file, engine='odf')
        df.to_csv(csv_file, index=False, header=False)
        return True
    except ImportError:
        print("Error: odfpy not installed. Run: pip3 install odfpy", file=sys.stderr)
        return False
    except Exception as e:
        print(f"Error converting ODS to CSV: {e}", file=sys.stderr)
        return False


def main():
    if len(sys.argv) != 3:
        print(f"Usage: {sys.argv[0]} <input_file> <output_csv>", file=sys.stderr)
        print("", file=sys.stderr)
        print("Supported formats: .xlsx, .ods", file=sys.stderr)
        sys.exit(1)

    input_file = Path(sys.argv[1])
    output_file = Path(sys.argv[2])

    if not input_file.exists():
        print(f"Error: Input file not found: {input_file}", file=sys.stderr)
        sys.exit(1)

    # Determine file type and convert
    suffix = input_file.suffix.lower()

    if suffix == '.xlsx':
        success = convert_xlsx_to_csv(input_file, output_file)
    elif suffix == '.ods':
        success = convert_ods_to_csv(input_file, output_file)
    else:
        print(f"Error: Unsupported file format: {suffix}", file=sys.stderr)
        print("Supported formats: .xlsx, .ods", file=sys.stderr)
        sys.exit(1)

    if success:
        print(f"✓ Converted {input_file} to {output_file}")
        sys.exit(0)
    else:
        sys.exit(1)


if __name__ == '__main__':
    main()
