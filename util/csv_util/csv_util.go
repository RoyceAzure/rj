package csv_util

import (
	"bytes"
	"encoding/csv"
	"errors"
)

type CsvUtilError error

var (
	ErrEmptyCSV CsvUtilError = errors.New("CSV file has no data")
)

func CsvToJSON(csvData []byte) ([]map[string]string, error) {
	reader := csv.NewReader(bytes.NewReader(csvData))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 1 {
		return []map[string]string{}, ErrEmptyCSV
	}

	headers := records[0]
	var result []map[string]string

	for _, record := range records[1:] {
		row := make(map[string]string)
		for i, value := range record {
			if i < len(headers) {
				row[headers[i]] = value
			}
		}
		result = append(result, row)
	}

	return result, nil
}
