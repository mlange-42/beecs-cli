package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type CsvWriter struct {
	builder     strings.Builder
	files       []*os.File
	initialized bool
}

func NewCsvWriter(files []string) (CsvWriter, error) {
	f := []*os.File{}

	for _, path := range files {
		err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
		if err != nil {
			return CsvWriter{}, err
		}

		ff, err := os.Create(path)
		if err != nil {
			return CsvWriter{}, err
		}
		f = append(f, ff)
	}

	return CsvWriter{
		files: f,
	}, nil
}

func (w *CsvWriter) Write(tables *Tables) error {
	if !w.initialized {
		for i := range tables.Headers {
			_, err := fmt.Fprintln(w.files[i], strings.Join(tables.Headers[i], ","))
			if err != nil {
				return err
			}
		}
		w.initialized = true
	}

	for i := range tables.Data {
		table := tables.Data[i]
		w.builder.Reset()
		for _, row := range table {
			for i, v := range row {
				fmt.Fprint(&w.builder, v)
				if i < len(row)-1 {
					fmt.Fprint(&w.builder, ",")
				}
			}
			fmt.Fprint(&w.builder, "\n")
		}
		_, err := fmt.Fprint(w.files[i], w.builder.String())
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *CsvWriter) Close() error {
	for _, f := range w.files {
		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
}
