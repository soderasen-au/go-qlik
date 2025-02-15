package ss

import (
	"testing"
	"time"
)

// TestRender runs unit tests for the RenderTpl function
func TestRender(t *testing.T) {
	// Setup variables in TplVarMap
	TplVarMap = TplVarMapType{
		"var1": func(input string) string { return "value1" },
		"var2": func(input string) string { return "value2" },
	}

	// Setup a fixed time for `today` function
	fixedTime := time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)
	timeNowFunc := func() time.Time { return fixedTime }
	timeNow := timeNowFunc // Mock the time.Now() function if necessary

	TplFuncMap["today"] = func(params []string) string {
		fmt := "2006-01-02"
		if len(params) > 0 {
			fmt = params[0]
		}
		return timeNow().Format(fmt)
	}
	TplFuncMap["yesterday"] = func(params []string) string {
		fmt := "2006-01-02"
		if len(params) > 0 {
			fmt = params[0]
		}
		yesterday := timeNow().AddDate(0, 0, -1)
		return yesterday.Format(fmt)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Function replacement - today with default format",
			input:    "$today()",
			expected: "2023-10-01",
		},
		{
			name:     "Function replacement - yesterday with default format",
			input:    "$yesterday()",
			expected: "2023-09-30",
		},
		{
			name:     "Function replacement - today with custom format",
			input:    "$today(02-Jan-2006)",
			expected: "01-Oct-2023",
		},
		{
			name:     "Function replacement - yesterday with custom format",
			input:    "$yesterday(02-Jan-2006)",
			expected: "30-Sep-2023",
		},
		{
			name:     "Variable replacement",
			input:    "$(var1)",
			expected: "value1",
		},
		{
			name:     "Unknown function",
			input:    "$unknown()",
			expected: "$unknown()",
		},
		{
			name:     "Unknown variable",
			input:    "$(var_unknown)",
			expected: "$(var_unknown)",
		},
		{
			name:     "Function replacement with invalid time format parameters",
			input:    "$yesterday(invalid-format)",
			expected: "invalid-format",
		},
		{
			name:     "Mixed content",
			input:    "Today is $today() and var is $(var1)",
			expected: "Today is 2023-10-01 and var is value1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderTpl(tt.input)
			if result != tt.expected {
				t.Errorf("RenderTpl(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}
