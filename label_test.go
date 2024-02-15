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
}

func TestBuildAddFieldQL(t *testing.T) {
	field := LabelFieldSchema{
		Field:    "name",
		Type:     "string",
		Nullable: false,
	}
	assert.Equal(t, "ALTER TAG account ADD (name string NOT NULL);", field.BuildAddFieldQL("account"))
	field.Nullable = true
	assert.Equal(t, "ALTER TAG account ADD (name string);", field.BuildAddFieldQL("account"))
}

func TestBuildDropFieldQL(t *testing.T) {
	field := Label{
		Field: "name",
	}
	assert.Equal(t, "ALTER TAG account DROP (name);", field.BuildDropFieldQL("account"))
}
