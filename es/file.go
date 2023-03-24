package es

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"io"

	"github.com/tidwall/gjson"
)

type Writer interface {
	io.Closer
	Write(ctx context.Context, b json.RawMessage) error
}

type CSVWriter struct {
	f      *csv.Writer
	fields []string
}

func (c *CSVWriter) Close() error {
	c.f.Flush()
	return nil
}

func (c *CSVWriter) Write(ctx context.Context, b json.RawMessage) error {
	results := gjson.GetManyBytes(b, c.fields...)
	fields := make([]string, len(results))
	for i, result := range results {
		fields[i] = result.String()
	}
	return c.f.Write(fields)

}

func NewCSVWriter(w io.Writer, fields []string) Writer {
	return &CSVWriter{
		f:      csv.NewWriter(w),
		fields: fields,
	}
}
