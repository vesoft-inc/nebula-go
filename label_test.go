/*
 *
 * Copyright (c) 2024 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */
package nebula_go

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildCreateTagQL(t *testing.T) {
	tag := LabelSchema{
		Name: "account",
		Fields: []LabelFieldSchema{
			{
				Field:    "name",
				Nullable: false,
			},
			{
				Field:    "email",
				Nullable: true,
			},
			{
				Field:    "phone",
				Nullable: true,
			},
		},
	}
	assert.Equal(t, "CREATE TAG IF NOT EXISTS account (name string NOT NULL, email string NULL, phone string NULL);", tag.BuildCreateTagQL())
	assert.Equal(t, "DROP TAG IF EXISTS account;", tag.BuildDropTagQL())

	tag.TTLDuration = 100
	tag.TTLCol = "created_at"
	assert.PanicsWithError(t, "TTL column created_at does not exist in the fields", func() {
		tag.BuildCreateTagQL()
	})

	tag.Fields = append(tag.Fields, LabelFieldSchema{
		Field:    "created_at",
		Type:     "timestamp",
		Nullable: true,
	})

	assert.Equal(t, `CREATE TAG IF NOT EXISTS account (name string NOT NULL, email string NULL, phone string NULL, created_at timestamp NULL) TTL_DURATION = 100, TTL_COL = "created_at";`, tag.BuildCreateTagQL())
}

func TestBuildCreateEdgeQL(t *testing.T) {
	edge := LabelSchema{
		Name: "account_email",
		Fields: []LabelFieldSchema{
			{
				Field:    "email",
				Nullable: false,
			},
		},
	}
	assert.Equal(t, "CREATE EDGE IF NOT EXISTS account_email (email string NOT NULL);", edge.BuildCreateEdgeQL())
	assert.Equal(t, "DROP EDGE IF EXISTS account_email;", edge.BuildDropEdgeQL())

	edge.TTLDuration = 100
	edge.TTLCol = "created_at"

	assert.PanicsWithError(t, "TTL column created_at does not exist in the fields", func() {
		edge.BuildCreateEdgeQL()
	})

	edge.Fields = append(edge.Fields, LabelFieldSchema{
		Field:    "created_at",
		Type:     "timestamp",
		Nullable: true,
	})

	assert.Equal(t, `CREATE EDGE IF NOT EXISTS account_email (email string NOT NULL, created_at timestamp NULL) TTL_DURATION = 100, TTL_COL = "created_at";`, edge.BuildCreateEdgeQL())
}

func TestBuildAddFieldQL(t *testing.T) {
	field := LabelFieldSchema{
		Field:    "name",
		Type:     "string",
		Nullable: false,
	}
	// tag
	assert.Equal(t, "ALTER TAG account ADD (name string NOT NULL);", field.BuildAddTagFieldQL("account"))
	field.Nullable = true
	assert.Equal(t, "ALTER TAG account ADD (name string);", field.BuildAddTagFieldQL("account"))
	// edge
	assert.Equal(t, "ALTER EDGE account ADD (name string);", field.BuildAddEdgeFieldQL("account"))
	field.Nullable = false
	assert.Equal(t, "ALTER EDGE account ADD (name string NOT NULL);", field.BuildAddEdgeFieldQL("account"))
}

func TestBuildDropFieldQL(t *testing.T) {
	field := Label{
		Field: "name",
	}
	// tag
	assert.Equal(t, "ALTER TAG account DROP (name);", field.BuildDropTagFieldQL("account"))
	// edge
	assert.Equal(t, "ALTER EDGE account DROP (name);", field.BuildDropEdgeFieldQL("account"))
}
