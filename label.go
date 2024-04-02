/*
 *
 * Copyright (c) 2024 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */
package nebula_go

import (
	"fmt"
	"strings"
)

type LabelName struct {
	Name string `nebula:"Name"`
}

type SpaceName struct {
	Name string `nebula:"Name"`
}

type Label struct {
	Field   string `nebula:"Field"`
	Type    string `nebula:"Type"`
	Null    string `nebula:"Null"`
	Default string `nebula:"Default"`
	Comment string `nebula:"Comment"`
}

type LabelSchema struct {
	Name        string
	Fields      []LabelFieldSchema
	TTLDuration uint
	TTLCol      string
}

type LabelFieldSchema struct {
	Field    string
	Type     string
	Nullable bool
}

func (tag LabelSchema) BuildCreateTagQL() string {
	q := "CREATE TAG IF NOT EXISTS " + tag.Name + " ("

	fields := []string{}
	for _, field := range tag.Fields {
		t := field.Type
		if t == "" {
			t = "string"
		}
		n := "NULL"
		if !field.Nullable {
			n = "NOT NULL"
		}
		fields = append(fields, field.Field+" "+t+" "+n)
	}

	ttl := tag.buildTTL_QL()

	if ttl != "" {
		q += strings.Join(fields, ", ") + ") " + ttl + ";"
	} else {
		q += strings.Join(fields, ", ") + ");"
	}

	return q
}

func (tag LabelSchema) BuildDropTagQL() string {
	q := "DROP TAG IF EXISTS " + tag.Name + ";"
	return q
}

func (edge LabelSchema) BuildCreateEdgeQL() string {
	q := "CREATE EDGE IF NOT EXISTS " + edge.Name + " ("

	fields := []string{}
	for _, field := range edge.Fields {
		t := field.Type
		if t == "" {
			t = "string"
		}
		n := "NULL"
		if !field.Nullable {
			n = "NOT NULL"
		}
		fields = append(fields, field.Field+" "+t+" "+n)
	}

	ttl := edge.buildTTL_QL()

	if ttl != "" {
		q += strings.Join(fields, ", ") + ") " + ttl + ";"
	} else {
		q += strings.Join(fields, ", ") + ");"
	}

	return q
}

func (edge LabelSchema) BuildDropEdgeQL() string {
	q := "DROP EDGE IF EXISTS " + edge.Name + ";"
	return q
}

func (label LabelSchema) buildTTL_QL() string {
	ttl := ""
	if label.TTLCol != "" {
		if !label.isTTLColValid() {
			panic(fmt.Errorf("TTL column %s does not exist in the fields", label.TTLCol))
		}
		ttl = fmt.Sprintf(`TTL_DURATION = %d, TTL_COL = "%s"`, label.TTLDuration, label.TTLCol)
	}

	return ttl
}

func (label LabelSchema) isTTLColValid() bool {
	if label.TTLCol == "" {
		// no ttl column is valid
		return true
	}

	for _, field := range label.Fields {
		if field.Field == label.TTLCol {
			return true
		}
	}

	return false
}

func (field LabelFieldSchema) BuildAddTagFieldQL(labelName string) string {
	q := "ALTER TAG " + labelName + " ADD (" + field.Field + " " + field.Type
	if !field.Nullable {
		q += " NOT NULL"
	}
	return q + ");"
}

func (field LabelFieldSchema) BuildAddEdgeFieldQL(labelName string) string {
	q := "ALTER EDGE " + labelName + " ADD (" + field.Field + " " + field.Type
	if !field.Nullable {
		q += " NOT NULL"
	}
	return q + ");"
}

func (field Label) BuildDropTagFieldQL(labelName string) string {
	return "ALTER TAG " + labelName + " DROP (" + field.Field + ");"
}

func (field Label) BuildDropEdgeFieldQL(labelName string) string {
	return "ALTER EDGE " + labelName + " DROP (" + field.Field + ");"
}
