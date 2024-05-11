//go:build integration
// +build integration

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

func TestSchemaManagerApplyTag(t *testing.T) {
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

	schemaManager := NewSchemaManager(sessionPool).WithVerbose(true)

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
				Type:     "string",
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

func TestSchemaManagerApplyEdge(t *testing.T) {
	spaceName := "test_space_apply_edge"
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

	schemaManager := NewSchemaManager(sessionPool).WithVerbose(true)

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

	edgeSchema := LabelSchema{
		Name: "account_email",
		Fields: []LabelFieldSchema{
			{
				Field:    "email",
				Type:     "string",
				Nullable: false,
			},
		},
	}
	_, err = schemaManager.ApplyEdge(edgeSchema)
	if err != nil {
		t.Fatal(err)
	}
	edges, err := sessionPool.ShowEdges()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(edges))
	assert.Equal(t, "account_email", edges[0].Name)
	labels, err := sessionPool.DescEdge("account_email")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, "email", labels[0].Field)
	assert.Equal(t, "string", labels[0].Type)

	edgeSchema = LabelSchema{
		Name: "account_email",
		Fields: []LabelFieldSchema{
			{
				Field:    "email",
				Type:     "string",
				Nullable: false,
			},
			{
				Field:    "created_at",
				Type:     "timestamp",
				Nullable: true,
			},
		},
	}
	_, err = schemaManager.ApplyEdge(edgeSchema)
	if err != nil {
		t.Fatal(err)
	}
	edges, err = sessionPool.ShowEdges()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(edges))
	assert.Equal(t, "account_email", edges[0].Name)
	labels, err = sessionPool.DescEdge("account_email")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, len(labels))
	assert.Equal(t, "email", labels[0].Field)
	assert.Equal(t, "string", labels[0].Type)
	assert.Equal(t, "created_at", labels[1].Field)
	assert.Equal(t, "timestamp", labels[1].Type)

	edgeSchema = LabelSchema{
		Name: "account_email",
		Fields: []LabelFieldSchema{
			{
				Field:    "email",
				Type:     "string",
				Nullable: false,
			},
		},
	}
	_, err = schemaManager.ApplyEdge(edgeSchema)
	if err != nil {
		t.Fatal(err)
	}
	edges, err = sessionPool.ShowEdges()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(edges))
	assert.Equal(t, "account_email", edges[0].Name)
	labels, err = sessionPool.DescEdge("account_email")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, "email", labels[0].Field)
	assert.Equal(t, "string", labels[0].Type)
}

func TestSchemaManagerApplyTagWithTTL(t *testing.T) {
	spaceName := "test_space_apply_tag_with_ttl"
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

	schemaManager := NewSchemaManager(sessionPool).WithVerbose(true)

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
		Name: "user",
		Fields: []LabelFieldSchema{
			{
				Field:    "name",
				Type:     "string",
				Nullable: false,
			},
			{
				Field:    "created_at",
				Type:     "int64",
				Nullable: false,
			},
		},
		TTLDuration: 100,
		TTLCol:      "created_at",
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
	labels, err := sessionPool.DescTag("user")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, len(labels))
	ttlCol, ttlDuration, err := sessionPool.GetTagTTL("user")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "created_at", ttlCol)
	assert.Equal(t, uint(100), ttlDuration)

	// update ttl
	tagSchema.TTLDuration = 300

	_, err = schemaManager.ApplyTag(tagSchema)
	if err != nil {
		t.Fatal(err)
	}
	labels, err = sessionPool.DescTag("user")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, len(labels))
	ttlCol, ttlDuration, err = sessionPool.GetTagTTL("user")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "created_at", ttlCol)
	assert.Equal(t, uint(300), ttlDuration)

	// remove ttl
	tagSchema.TTLDuration = 0
	tagSchema.TTLCol = ""

	_, err = schemaManager.ApplyTag(tagSchema)
	if err != nil {
		t.Fatal(err)
	}
	labels, err = sessionPool.DescTag("user")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, len(labels))
	ttlCol, ttlDuration, err = sessionPool.GetTagTTL("user")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "", ttlCol)
	assert.Equal(t, uint(0), ttlDuration)
}

func TestSchemaManagerApplyEdgeWithTTL(t *testing.T) {
	spaceName := "test_space_apply_edge_with_ttl"
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

	schemaManager := NewSchemaManager(sessionPool).WithVerbose(true)

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

	edgeSchema := LabelSchema{
		Name: "friend",
		Fields: []LabelFieldSchema{
			{
				Field:    "created_at",
				Type:     "int64",
				Nullable: false,
			},
		},
		TTLDuration: 100,
		TTLCol:      "created_at",
	}
	_, err = schemaManager.ApplyEdge(edgeSchema)
	if err != nil {
		t.Fatal(err)
	}
	edges, err := sessionPool.ShowEdges()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(edges))
	labels, err := sessionPool.DescEdge("friend")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(labels))
	ttlCol, ttlDuration, err := sessionPool.GetEdgeTTL("friend")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "created_at", ttlCol)
	assert.Equal(t, uint(100), ttlDuration)

	// update ttl
	edgeSchema.TTLDuration = 300

	_, err = schemaManager.ApplyEdge(edgeSchema)
	if err != nil {
		t.Fatal(err)
	}
	labels, err = sessionPool.DescEdge("friend")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(labels))
	ttlCol, ttlDuration, err = sessionPool.GetEdgeTTL("friend")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "created_at", ttlCol)
	assert.Equal(t, uint(300), ttlDuration)

	// remove ttl
	edgeSchema.TTLDuration = 0
	edgeSchema.TTLCol = ""

	_, err = schemaManager.ApplyEdge(edgeSchema)
	if err != nil {
		t.Fatal(err)
	}
	labels, err = sessionPool.DescEdge("friend")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(labels))
	ttlCol, ttlDuration, err = sessionPool.GetEdgeTTL("friend")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "", ttlCol)
	assert.Equal(t, uint(0), ttlDuration)
}
