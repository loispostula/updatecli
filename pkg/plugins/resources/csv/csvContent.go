package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/updatecli/updatecli/pkg/plugins/utils"
	"github.com/updatecli/updatecli/pkg/plugins/utils/dasel"
)

// csvContent is *** of the dasel FileContent
type csvContent struct {
	dasel.FileContent
	rows    []map[string]interface{}
	headers []string
	comma   rune
	comment rune
}

func (c *csvContent) Read(rootDir string) error {

	securePath, err := utils.SanitizeFilePathWithWorkingDirectory(c.FilePath, rootDir)
	if err != nil {
		return err
	}
	c.FilePath = securePath

	// Test at runtime if a file exist
	if !c.ContentRetriever.FileExists(c.FilePath) {
		return fmt.Errorf("the CSV file %q does not exist", c.FilePath)
	}

	textContent, err := c.ContentRetriever.ReadAll(c.FilePath)
	if err != nil {
		return err
	}

	r := csv.NewReader(strings.NewReader(textContent))

	r.Comma = c.comma
	r.Comment = c.comment

	res := make([]map[string]interface{}, 0)
	records, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("could not read csv file: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("no csv record found")
	}

	var headers []string
	for i, row := range records {
		if i == 0 {
			headers = row
			continue
		}
		rowRes := make(map[string]interface{})
		allEmpty := true
		for index, val := range row {
			if val != "" {
				allEmpty = false
			}
			rowRes[headers[index]] = val
		}
		if !allEmpty {
			res = append(res, rowRes)
		}
	}

	c.rows = res
	c.headers = headers
	c.DaselV3Data = c.rows

	return nil
}

// derefValue unwraps pointer/interface indirections around a value. The dasel v3
// engine stores modified values as *interface{}, so this resolves them to the
// underlying value before serialization.
func derefValue(v interface{}) interface{} {
	rv := reflect.ValueOf(v)
	for rv.IsValid() && (rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface) {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if !rv.IsValid() {
		return nil
	}
	return rv.Interface()
}

func (c *csvContent) Write() error {
	newFile, err := os.Create(c.FilePath)
	if err != nil {
		return fmt.Errorf("could not write to file : %w", err)
	}

	defer newFile.Close()

	writer := csv.NewWriter(newFile)

	writer.Comma = c.comma

	// Iterate through the rows and write the output.
	for i, r := range c.rows {
		if i == 0 {
			if err := writer.Write(c.headers); err != nil {
				return fmt.Errorf("could not write headers: %w", err)
			}
		}

		values := make([]string, 0)
		for _, header := range c.headers {
			val, ok := r[header]
			if !ok {
				val = ""
			}
			values = append(values, fmt.Sprint(derefValue(val)))
		}

		if err := writer.Write(values); err != nil {
			return fmt.Errorf("could not write headers: %w", err)
		}

		writer.Flush()
	}

	return nil
}
