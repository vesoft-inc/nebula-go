package nebula_go

import (
	"strings"
)

type Label struct {
	Field   string `nebula:"Field"`
	Type    string `nebula:"Type"`
	Null    string `nebula:"Null"`
	Default string `nebula:"Default"`
	Comment string `nebula:"Comment"`
}

type LabelFieldSchema struct {
	Field string
	Type  string
	Null  bool
}

type LabelSchema struct {
	Name   string
	Fields []LabelFieldSchema
}

func (tag LabelSchema) BuildCreateTagQL() string {
	q := "CREATE TAG IF NOT EXISTS " + tag.Name + " ("

	fs := []string{}
	for _, field := range tag.Fields {
		t := field.Type
		if t == "" {
			t = "string"
		}
		n := "NULL"
		if !field.Null {
			n = "NOT NULL"
		}
		fs = append(fs, field.Field+" "+t+" "+n)
	}

	q += strings.Join(fs, ", ") + ");"

	return q
}

func (tag LabelSchema) BuildDropTagQL() string {
	q := "DROP TAG IF EXISTS " + tag.Name + ";"
	return q
}

func (edge LabelSchema) BuildCreateEdgeQL() string {
	q := "CREATE EDGE IF NOT EXISTS " + edge.Name + " ("

	fs := []string{}
	for _, field := range edge.Fields {
		t := field.Type
		if t == "" {
			t = "string"
		}
		n := "NULL"
		if !field.Null {
			n = "NOT NULL"
		}
		fs = append(fs, field.Field+" "+t+" "+n)
	}

	if len(fs) > 0 {
		q += strings.Join(fs, ", ")
	}

	return q + ");"
}

func (edge LabelSchema) BuildDropEdgeQL() string {
	q := "DROP EDGE IF EXISTS " + edge.Name + ";"
	return q
}
