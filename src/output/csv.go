package output

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NITTC-Robosemi/stcm-viewer/src/model"
)

// safeName replaces characters that are problematic in file names.
func safeName(name string) string {
	name = strings.ReplaceAll(name, ",", "_")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, ":", "_")
	return strings.TrimSpace(name)
}

// WriteCSV writes all parsed data as CSV files into the given directory.
// It returns the directory path created.
func WriteCSV(baseDir string, allData model.ParsedData) (string, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", baseDir, err)
	}

	for groupName, variables := range allData {
		groupDir := filepath.Join(baseDir, safeName(groupName))
		if err := os.MkdirAll(groupDir, 0o755); err != nil {
			return "", fmt.Errorf("failed to create group directory %s: %w", groupDir, err)
		}

		for varName, data := range variables {
			filePath := filepath.Join(groupDir, safeName(varName)+".csv")
			file, err := os.Create(filePath)
			if err != nil {
				return "", fmt.Errorf("failed to create csv file %s: %w", filePath, err)
			}

			writer := csv.NewWriter(file)
			writer.Comma = ';'
			writeErr := func() error {
				if err := writer.Write([]string{"Time", varName}); err != nil {
					return fmt.Errorf("failed to write csv header: %w", err)
				}

				for i := range data.X {
					record := []string{
						fmt.Sprintf("%g", data.X[i]),
						fmt.Sprintf("%g", data.Y[i]),
					}
					if err := writer.Write(record); err != nil {
						return fmt.Errorf("failed to write csv record: %w", err)
					}
				}
				writer.Flush()
				if err := writer.Error(); err != nil {
					return fmt.Errorf("failed to flush csv writer: %w", err)
				}
				return nil
			}()
			closeErr := file.Close()
			if writeErr != nil {
				return "", writeErr
			}
			if closeErr != nil {
				return "", fmt.Errorf("failed to close csv file %s: %w", filePath, closeErr)
			}
		}
	}

	return baseDir, nil
}
