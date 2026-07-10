package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/NITTC-Robosemi/stcm-viewer/src/model"
)

// removeTrailingComma removes a trailing comma that appears before the final
// closing bracket of the JSON array, matching the original Python behaviour.
func removeTrailingComma(data string) string {
	trimmed := strings.TrimSpace(data)
	if len(trimmed) < 2 {
		return data
	}
	// Find the last non-whitespace character before the final closing bracket.
	endIdx := len(trimmed) - 1
	// Look for a comma before the closing bracket, ignoring whitespace.
	for i := endIdx - 1; i >= 0; i-- {
		c := trimmed[i]
		if c == ',' {
			return trimmed[:i] + trimmed[endIdx:]
		}
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			// Found a non-whitespace, non-comma character: no trailing comma.
			break
		}
	}
	// Preserve original whitespace if no trailing comma was found.
	return data
}

// ParseSTCMFile parses an STCM file and returns the parsed data.
func ParseSTCMFile(filepath string) (model.ParsedData, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	dataString := removeTrailingComma(string(content))
	dataString = strings.TrimSpace(dataString)

	var entries []model.STCMEntry
	if err := json.Unmarshal([]byte(dataString), &entries); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	allData := make(model.ParsedData)
	for _, entry := range entries {
		group, ok := allData[entry.GroupName]
		if !ok {
			group = make(map[string]*model.VariableData)
			allData[entry.GroupName] = group
		}

		variable, ok := group[entry.VariableName]
		if !ok {
			variable = &model.VariableData{}
			group[entry.VariableName] = variable
		}

		for _, p := range entry.VariableData {
			variable.X = append(variable.X, p.X)
			variable.Y = append(variable.Y, p.Y)
		}
	}

	return allData, nil
}
