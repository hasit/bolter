package main

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

type tableFormatter struct {
	noValues bool
}

func (tf tableFormatter) DumpBuckets(w io.Writer, buckets []bucket) {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Buckets"})
	for _, b := range buckets {
		row := []string{b.Name}
		table.Append(row)
	}
	table.Render()
}

func (tf tableFormatter) DumpBucketItems(w io.Writer, bucket string, items []item) {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Key", "Value"})
	for _, item := range items {
		var row []string
		if tf.noValues {
			row = []string{item.Key, ""}
		} else {
			row = []string{item.Key, item.Value}
		}
		table.Append(row)
	}
	table.Render()
}
