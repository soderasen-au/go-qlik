#!/usr/bin/env python3
"""
extract-pdf-issues.py - Detailed PDF issue analysis

This script performs deep analysis of PDF content to identify specific issues:
- Text truncation patterns
- Column width problems
- Formatting inconsistencies
"""

import sys
import re
from pathlib import Path
from typing import List, Dict, Tuple

def analyze_pdf_text(pdf_text_file: Path) -> Dict:
    """Analyze extracted PDF text for issues."""

    if not pdf_text_file.exists():
        return {
            'error': f'PDF text file not found: {pdf_text_file}',
            'truncations': [],
            'total_truncations': 0
        }

    with open(pdf_text_file, 'r', encoding='utf-8') as f:
        content = f.read()

    # Find all truncated entries (ending with ...)
    truncation_pattern = r'([A-Za-z][A-Za-z0-9 \'\-]*)\.\.\.'
    truncations = re.findall(truncation_pattern, content)

    # Analyze truncation patterns
    truncation_info = []
    for text in truncations:
        truncation_info.append({
            'text': text + '...',
            'length': len(text),
            'first_word': text.split()[0] if text.split() else '',
        })

    # Group by first word to identify patterns
    patterns = {}
    for info in truncation_info:
        first_word = info['first_word']
        if first_word not in patterns:
            patterns[first_word] = []
        patterns[first_word].append(info['text'])

    return {
        'truncations': truncation_info,
        'total_truncations': len(truncations),
        'patterns': patterns,
        'unique_prefixes': len(patterns)
    }


def compare_csv_files(reference_csv: Path, generated_csv: Path) -> Dict:
    """Compare two CSV files and identify specific differences."""

    if not reference_csv.exists() or not generated_csv.exists():
        return {
            'error': 'One or both CSV files not found',
            'differences': []
        }

    with open(reference_csv, 'r', encoding='utf-8') as f:
        ref_lines = f.readlines()

    with open(generated_csv, 'r', encoding='utf-8') as f:
        gen_lines = f.readlines()

    differences = []
    max_lines = max(len(ref_lines), len(gen_lines))

    for i in range(max_lines):
        ref_line = ref_lines[i].strip() if i < len(ref_lines) else ''
        gen_line = gen_lines[i].strip() if i < len(gen_lines) else ''

        if ref_line != gen_line:
            differences.append({
                'line': i + 1,
                'reference': ref_line,
                'generated': gen_line
            })

    return {
        'total_differences': len(differences),
        'differences': differences[:50],  # First 50 differences
        'reference_lines': len(ref_lines),
        'generated_lines': len(gen_lines)
    }


def generate_html_report(analysis: Dict, output_file: Path):
    """Generate an HTML report with visual diff."""

    html = """<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>PDF Report Comparison Analysis</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background-color: white;
            padding: 30px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 {
            color: #2c3e50;
            border-bottom: 3px solid #3498db;
            padding-bottom: 10px;
        }
        h2 {
            color: #34495e;
            margin-top: 30px;
            border-left: 4px solid #3498db;
            padding-left: 10px;
        }
        .status-pass {
            color: #27ae60;
            font-weight: bold;
        }
        .status-fail {
            color: #e74c3c;
            font-weight: bold;
        }
        .status-warn {
            color: #f39c12;
            font-weight: bold;
        }
        .truncation {
            background-color: #fff3cd;
            border-left: 4px solid #ffc107;
            padding: 10px;
            margin: 10px 0;
            font-family: monospace;
        }
        .diff-line {
            font-family: monospace;
            padding: 5px;
            margin: 2px 0;
        }
        .diff-removed {
            background-color: #ffdddd;
            color: #d63031;
        }
        .diff-added {
            background-color: #ddffdd;
            color: #00b894;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
        }
        th, td {
            border: 1px solid #ddd;
            padding: 12px;
            text-align: left;
        }
        th {
            background-color: #3498db;
            color: white;
        }
        tr:nth-child(even) {
            background-color: #f2f2f2;
        }
        .code {
            background-color: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 4px;
            padding: 15px;
            font-family: monospace;
            overflow-x: auto;
            margin: 15px 0;
        }
        .recommendation {
            background-color: #e8f4f8;
            border-left: 4px solid #3498db;
            padding: 15px;
            margin: 20px 0;
        }
        .metric {
            display: inline-block;
            background-color: #ecf0f1;
            padding: 10px 20px;
            margin: 5px;
            border-radius: 4px;
        }
        .metric-value {
            font-size: 24px;
            font-weight: bold;
            color: #2c3e50;
        }
        .metric-label {
            font-size: 12px;
            color: #7f8c8d;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>📊 PDF Report Comparison Analysis</h1>
"""

    # Add metrics
    pdf_analysis = analysis.get('pdf', {})
    csv_analysis = analysis.get('csv', {})

    html += f"""
        <div style="margin: 20px 0;">
            <div class="metric">
                <div class="metric-value">{pdf_analysis.get('total_truncations', 0)}</div>
                <div class="metric-label">Truncated Text Entries</div>
            </div>
            <div class="metric">
                <div class="metric-value">{csv_analysis.get('total_differences', 0)}</div>
                <div class="metric-label">CSV Differences</div>
            </div>
            <div class="metric">
                <div class="metric-value">{pdf_analysis.get('unique_prefixes', 0)}</div>
                <div class="metric-label">Unique Truncation Patterns</div>
            </div>
        </div>
"""

    # PDF Analysis Section
    if pdf_analysis.get('total_truncations', 0) > 0:
        html += """
        <h2>🔍 PDF Text Truncation Issues</h2>
        <p class="status-warn">⚠️ WARNING: Text truncation detected in PDF output</p>

        <div class="recommendation">
            <strong>Root Cause:</strong> Aggressive text truncation in <code>report/pdf.go:306-310</code><br>
            <strong>Impact:</strong> Product names and other text fields are cut off with "..."<br>
            <strong>Fix Required:</strong> Use actual font metrics instead of approximate formula
        </div>

        <h3>Truncation Examples</h3>
        <table>
            <tr>
                <th>Truncated Text</th>
                <th>Length</th>
            </tr>
"""

        for trunc in pdf_analysis.get('truncations', [])[:30]:
            html += f"""
            <tr>
                <td class="truncation">{trunc['text']}</td>
                <td>{trunc['length']} chars</td>
            </tr>
"""

        html += """
        </table>

        <h3>Recommended Code Fix</h3>
        <div class="code">
<span style="color: #d63031;">-    maxLen := int(colWidth / 2.0)  // Approximate characters that fit</span>
<span style="color: #d63031;">-    if len(cellText) > maxLen {</span>
<span style="color: #d63031;">-        cellText = cellText[:maxLen-3] + "..."</span>
<span style="color: #d63031;">-    }</span>

<span style="color: #00b894;">+    // Use actual font metrics to determine if text fits</span>
<span style="color: #00b894;">+    textWidth := p.pdf.GetStringWidth(cellText)</span>
<span style="color: #00b894;">+    availableWidth := colWidth - 2*PDF_CELL_PADDING</span>
<span style="color: #00b894;">+    </span>
<span style="color: #00b894;">+    if textWidth > availableWidth {</span>
<span style="color: #00b894;">+        // Iteratively reduce text until it fits with "..."</span>
<span style="color: #00b894;">+        suffix := "..."</span>
<span style="color: #00b894;">+        suffixWidth := p.pdf.GetStringWidth(suffix)</span>
<span style="color: #00b894;">+        </span>
<span style="color: #00b894;">+        for len(cellText) > 0 {</span>
<span style="color: #00b894;">+            if p.pdf.GetStringWidth(cellText) + suffixWidth <= availableWidth {</span>
<span style="color: #00b894;">+                cellText = cellText + suffix</span>
<span style="color: #00b894;">+                break</span>
<span style="color: #00b894;">+            }</span>
<span style="color: #00b894;">+            cellText = cellText[:len(cellText)-1]</span>
<span style="color: #00b894;">+        }</span>
<span style="color: #00b894;">+    }</span>
        </div>
"""
    else:
        html += """
        <h2>✅ PDF Text Analysis</h2>
        <p class="status-pass">✓ No text truncation issues detected</p>
"""

    # CSV Comparison Section
    if csv_analysis.get('total_differences', 0) > 0:
        html += f"""
        <h2>📝 CSV Comparison Results</h2>
        <p class="status-warn">Found {csv_analysis['total_differences']} differences between reference and generated files</p>

        <h3>First 30 Differences</h3>
        <table>
            <tr>
                <th>Line</th>
                <th>Reference</th>
                <th>Generated</th>
            </tr>
"""

        for diff in csv_analysis.get('differences', [])[:30]:
            ref_text = diff['reference'][:100] if diff['reference'] else '(empty)'
            gen_text = diff['generated'][:100] if diff['generated'] else '(empty)'

            html += f"""
            <tr>
                <td>{diff['line']}</td>
                <td class="diff-removed">{ref_text}</td>
                <td class="diff-added">{gen_text}</td>
            </tr>
"""

        html += """
        </table>
"""
    else:
        html += """
        <h2>✅ CSV Comparison Results</h2>
        <p class="status-pass">✓ Excel output matches reference file exactly</p>
"""

    html += """
    </div>
</body>
</html>
"""

    with open(output_file, 'w', encoding='utf-8') as f:
        f.write(html)


def main():
    """Main entry point."""

    # Paths
    temp_dir = Path('test-reports/.tmp')
    output_dir = Path('test-reports')

    print("📊 Analyzing PDF and CSV outputs...")
    print()

    # Analyze PDF
    pdf_text_file = temp_dir / 'pdf_text.txt'
    pdf_analysis = analyze_pdf_text(pdf_text_file)

    print(f"PDF Analysis:")
    print(f"  Truncations found: {pdf_analysis.get('total_truncations', 0)}")
    print(f"  Unique patterns:   {pdf_analysis.get('unique_prefixes', 0)}")
    print()

    # Analyze CSV
    ref_csv = temp_dir / 'reference.csv'
    gen_csv = temp_dir / 'generated.csv'
    csv_analysis = compare_csv_files(ref_csv, gen_csv)

    print(f"CSV Comparison:")
    print(f"  Total differences: {csv_analysis.get('total_differences', 0)}")
    print()

    # Generate HTML report
    html_output = output_dir / 'comparison-report.html'
    generate_html_report({
        'pdf': pdf_analysis,
        'csv': csv_analysis
    }, html_output)

    print(f"✓ HTML report generated: {html_output}")
    print()

    # Generate detailed JSON output
    import json
    json_output = output_dir / 'comparison-data.json'
    with open(json_output, 'w', encoding='utf-8') as f:
        json.dump({
            'pdf_analysis': pdf_analysis,
            'csv_analysis': csv_analysis
        }, f, indent=2)

    print(f"✓ JSON data saved: {json_output}")


if __name__ == '__main__':
    main()
