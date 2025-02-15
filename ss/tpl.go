package ss

import (
	"regexp"
	"strings"
	"time"

	"github.com/soderasen-au/go-common/util"
)

type (

	// TplVarMapFunc defines a function type that transform var-name to var-value.
	TplVarMapFunc func(string) string

	// TplVarMapType defines a map where keys are var-name, and values are function that generate var-value.
	TplVarMapType map[string]TplVarMapFunc

	// TplFuncMapFunc defines a function type that processes a slice of strings as parameter list and returns a single string as function result.
	TplFuncMapFunc func([]string) string

	// TplFuncMapType defines a map where keys are function names, and values are functions processing string params to return a string result.
	TplFuncMapType map[string]TplFuncMapFunc
)

var TplFuncMap = TplFuncMapType{
	"today": func(params []string) string {
		if len(params) == 0 {
			return time.Now().Format("2006-01-02") // Default format
		}
		return time.Now().Format(util.FromISO8601(params[0])) // Use custom format
	},
	"yesterday": func(params []string) string {
		yesterday := time.Now().AddDate(0, 0, -1)
		if len(params) == 0 {
			return yesterday.Format("2006-01-02")
		}
		return yesterday.Format(util.FromISO8601(params[0]))
	},
}

var TplVarMap = TplVarMapType{}

func RenderTpl(input string) string {
	// Regex pattern to match `$func(...)` or `$(var)`
	pattern := `\$(\w+)?\(([^)]*)\)`
	re := regexp.MustCompile(pattern)

	// Replace each match
	result := re.ReplaceAllStringFunc(input, func(match string) string {
		matches := re.FindStringSubmatch(match)
		if len(matches) < 3 {
			return match // No valid match found
		}

		funcName := matches[1]                  // Function name or variable (can be empty)
		params := strings.TrimSpace(matches[2]) // Extract parameters

		// Case 1: Variable replacement (e.g., `$(var)`)
		if funcName == "" {
			if varFunc, exists := TplVarMap[params]; exists {
				return varFunc(params) // Replace with variable value
			}
			return match // Variable not found, return as is
		}

		// Case 2: Function replacement (e.g., `$func(param1, param2)`)
		paramList := []string{}
		if params != "" {
			paramList = strings.Split(params, ",")
			for i, param := range paramList {
				paramList[i] = strings.TrimSpace(param) // Trim spaces
			}
		}

		// Check if function exists
		if function, exists := TplFuncMap[funcName]; exists {
			return function(paramList) // Call function with parameters
		}

		return match // Function not found, return original match
	})

	return result
}
