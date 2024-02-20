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

func TestSessionPoolApplyTag(t *testing.T) {
	spaceName := "test_space_apply_tag"
	err := prepareSpace(spaceName)
	if err != nil {
		t.Fatal(err)
	}
	defer dropSpace(spaceName)

	hostAddress := HostAddress{Host: address, Port: port}
	config, err := NewSessionPoolConf(
		"root",
		"nebula",
		[]HostAddress{hostAddress},
		spaceName)
	if err != nil {
		t.Errorf("failed to create session pool config, %s", err.Error())
	}

	// allow only one session in the pool so it is easier to test
	config.maxSize = 1

	// create session pool
	sessionPool, err := NewSessionPool(*config, DefaultLogger{})
	if err != nil {
		t.Fatal(err)
	}
	defer sessionPool.Close()

	schemaManager := NewSchemaManager(sessionPool)

	spaces, err := sessionPool.ShowSpaces()
	if err != nil {
		t.Fatal(err)
	}
	assert.LessOrEqual(t, 1, len(spaces))
	var spaceNames []string
	for _, space := range spaces {
		spaceNames = append(spaceNames, space.Name)
	}
	assert.Contains(t, spaceNames, spaceName)

	tagSchema := LabelSchema{
		Name: "account",
		Fields: []LabelFieldSchema{
			{
				Field:    "name",
				Nullable: false,
			},
		},
	}
	_, err = schemaManager.ApplyTag(tagSchema)
	if err != nil {
		t.Fatal(err)
	}
	tags, err := sessionPool.ShowTags()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(tags))
	assert.Equal(t, "account", tags[0].Name)
	labels, err := sessionPool.DescTag("account")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, "name", labels[0].Field)
	assert.Equal(t, "string", labels[0].Type)

	tagSchema = LabelSchema{
		Name: "account",
		Fields: []LabelFieldSchema{
			{
				Field:    "name",
				Type:     "string",
				Nullable: false,
			},
			{
				Field:    "email",
				Type:     "string",
				Nullable: true,
			},
			{
				Field:    "phone",
				Type:     "int64",
				Nullable: true,
			},
		},
	}
	_, err = schemaManager.ApplyTag(tagSchema)
	if err != nil {
		t.Fatal(err)
	}
	tags, err = sessionPool.ShowTags()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(tags))
	assert.Equal(t, "account", tags[0].Name)
	labels, err = sessionPool.DescTag("account")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 3, len(labels))
	assert.Equal(t, "name", labels[0].Field)
	assert.Equal(t, "string", labels[0].Type)
	assert.Equal(t, "email", labels[1].Field)
	assert.Equal(t, "string", labels[1].Type)
	assert.Equal(t, "phone", labels[2].Field)
	assert.Equal(t, "int64", labels[2].Type)
	tagSchema = LabelSchema{
		Name: "account",
		Fields: []LabelFieldSchema{
			{
				Field:    "name",
				Type:     "string",
				Nullable: false,
			},
			{
				Field:    "phone",
				Type:     "int64",
				Nullable: true,
			},
		},
	}
	_, err = schemaManager.ApplyTag(tagSchema)
	if err != nil {
		t.Fatal(err)
	}
	tags, err = sessionPool.ShowTags()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(tags))
	assert.Equal(t, "account", tags[0].Name)
	labels, err = sessionPool.DescTag("account")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, len(labels))
	assert.Equal(t, "name", labels[0].Field)
	assert.Equal(t, "string", labels[0].Type)
	assert.Equal(t, "phone", labels[1].Field)
	assert.Equal(t, "int64", labels[1].Type)
}
