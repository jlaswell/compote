package pkg

import (
	"io"
	"os"

	"github.com/jedib0t/go-pretty/table"
)

type Table struct {
	table  table.Writer
	header []interface{}
	rows   [][]interface{}
	output io.Writer
}

func NewTable() *Table {
	t := table.NewWriter()
	t.Style().Options = table.OptionsNoBordersAndSeparators
	return &Table{
		table:  t,
		header: make([]interface{}, 0),
		rows:   make([][]interface{}, 0),
		output: os.Stdout,
	}
}

func (t *Table) SetOutput(w io.Writer) {
	t.output = w
	t.table.SetOutputMirror(w)
}

func (t *Table) SetHeader(h []interface{}) {
	t.header = h
}

func (t *Table) AppendRow(r []interface{}) {
	t.rows = append(t.rows, r)
}

func (t *Table) Render() {
	t.table.AppendHeader(t.header)
	for _, r := range t.rows {
		t.table.AppendRow(r)
	}
	t.table.Render()
}
