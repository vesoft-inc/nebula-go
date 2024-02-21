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

type SchemaManager struct {
	pool *SessionPool
}

func NewSchemaManager(pool *SessionPool) *SchemaManager {
	return &SchemaManager{pool: pool}
}

// ApplyTag applies the given tag to the graph.
// 1. If the tag does not exist, it will be created.
// 2. If the tag exists, it will be checked if the fields are the same.
// 2.1 If not, the new fields will be added.
// 2.2 If the field type is different, it will return an error.
// 2.3 If a field exists in the graph but not in the given tag,
// it will be removed.
//
// Notice:
// We won't change the field type because it has
// unexpected behavior for the data.
func (mgr *SchemaManager) ApplyTag(tag LabelSchema) (*ResultSet, error) {
	// 1. Check if the tag exists
	fields, err := mgr.pool.DescTag(tag.Name)
	if err != nil {
		// 2. If the tag does not exist, create it
		if strings.Contains(err.Error(), ErrorTagNotFound) {
			return mgr.pool.CreateTag(tag)
		}
		return nil, err
	}

	// 3. If the tag exists, check if the fields are the same
	if err != nil {
		return nil, err
	}

	// 4. Add new fields
	// 4.1 Prepare the new fields
	addFieldQLs := []string{}
	for _, expected := range tag.Fields {
		found := false
		for _, actual := range fields {
			if expected.Field == actual.Field {
				found = true
				// 4.2 Check if the field type is different
				if expected.Type != actual.Type {
					return nil, fmt.Errorf("field type is different. "+
						"Expected: %s, Actual: %s", expected.Type, actual.Type)
				}
				break
			}
		}
		if !found {
			// 4.3 Add the not exists field QL
			q := expected.BuildAddTagFieldQL(tag.Name)
			addFieldQLs = append(addFieldQLs, q)
		}
	}
	// 4.4 Execute the add field QLs if needed
	if len(addFieldQLs) > 0 {
		queries := strings.Join(addFieldQLs, "")
		_, err = mgr.pool.ExecuteAndCheck(queries)
		if err != nil {
			return nil, err
		}
	}

	// 5. Remove the not expected field
	for _, actual := range fields {
		redundant := true
		for _, expected := range tag.Fields {
			if expected.Field == actual.Field {
				redundant = false
				break
			}
		}
		if redundant {
			// 5.1 Remove the not expected field
			q := actual.BuildDropTagFieldQL(tag.Name)
			_, err := mgr.pool.ExecuteAndCheck(q)
			if err != nil {
				return nil, err
			}
		}
	}

	return nil, nil
}
