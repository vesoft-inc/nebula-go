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
				Field: "name",
				Null:  false,
			},
			{
				Field: "email",
				Null:  true,
			},
			{
				Field: "phone",
				Null:  true,
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
				Field: "email",
				Null:  false,
			},
		},
	}
	assert.Equal(t, "CREATE EDGE IF NOT EXISTS account_email (email string NOT NULL);", edge.BuildCreateEdgeQL())
	assert.Equal(t, "DROP EDGE IF EXISTS account_email;", edge.BuildDropEdgeQL())
}
