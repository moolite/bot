package db

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"time"
)

func ExportTable(table string) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer([]byte{})
	w := csv.NewWriter(buf)
	defer w.Flush()

	rows, err := dbc.Query(`SELECT FROM ` + table)
	if err != nil {
		return buf, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return buf, err
	}

	err = w.Write(cols)
	if err != nil {
		return buf, err
	}
	colLen := len(cols)
	valsptr := make([]interface{}, colLen)

	for rows.Next() {
		row := make([]string, colLen)

		if err := rows.Scan(valsptr...); err != nil {
			return buf, err
		}

		for i, val := range valsptr {
			switch typed := val.(type) {
			case float64:
			case float32:
				row[i] = fmt.Sprintf("%.2f", typed)
			case int32:
			case int64:
				row[i] = fmt.Sprintf("%d", typed)
			case time.Time:
				row[i] = typed.UTC().Format(time.RFC3339)
			case string:
				row[i] = typed
			default:
				row[i] = fmt.Sprintf("%x", typed)
			}
		}

		err = w.Write(row)
		if err != nil {
			return buf, err
		}
	}

	return buf, nil
}

func ExportDB() (ret map[string]*bytes.Buffer, err error) {
	tables := []string{
		groupsTable,
		abraxoidesTable,
		calloutsTable,
		linksTable,
		statisticsTable,
		statisticsKindTable,
	}

	for _, table := range tables {
		data, err := ExportTable(table)
		if err != nil {
			return ret, err
		}

		ret[table] = data
	}

	return ret, nil
}

func ExportTableToFile(filename, table string) error {
	buf, err := ExportTable(table)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, buf.Bytes(), os.ModePerm)
}

func ExportDBToFiles(filepath string) ([]string, error) {
	var filenames []string
	data, err := ExportDB()
	if err != nil {
		return filenames, err
	}

	for key, val := range data {
		filename := path.Join(filepath, key+".csv")
		err := os.WriteFile(filename, val.Bytes(), os.ModePerm)
		if err != nil {
			return filenames, err
		}
		filenames = append(filenames, filename)
	}

	return filenames, nil
}
