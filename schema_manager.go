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
	"log"
	"strings"
)

type SchemaManager struct {
	pool    *SessionPool
	verbose bool
}

func NewSchemaManager(pool *SessionPool) *SchemaManager {
	return &SchemaManager{pool: pool}
}

func (mgr *SchemaManager) WithVerbose(verbose bool) *SchemaManager {
	mgr.verbose = verbose
	return mgr
}

// ApplyTag applies the given tag to the graph.
// 1. If the tag does not exist, it will be created.
// 2. If the tag exists, it will be checked if the fields are the same.
// 2.1 If not, the new fields will be added.
// 2.2 If the field type is different, it will return an error.
// 2.3 If a field exists in the graph but not in the given tag,
// it will be removed.
// 3. If the tag exists and the fields are the same,
// it will be checked if the TTL is set as expected.
//
// Notice:
// We won't change the field type because it has
// unexpected behavior for the data.
func (mgr *SchemaManager) ApplyTag(tag LabelSchema) (*ResultSet, error) {
	// 1. Make sure the tag exists
	fields, err := mgr.pool.DescTag(tag.Name)
	if err != nil {
		// If the tag does not exist, create it
		if strings.Contains(err.Error(), ErrorTagNotFound) {
			if mgr.verbose {
				log.Printf("ApplyTag: create the not existing tag. name=%s\n", tag.Name)
			}
			return mgr.pool.CreateTag(tag)
		}
		return nil, err
	}

	// 2. The tag exists, add new fields if needed
	// 2.1 Prepare the new fields
	addFieldQLs := []string{}
	for _, expected := range tag.Fields {
		found := false
		for _, actual := range fields {
			if expected.Field == actual.Field {
				found = true
				// 2.2 Check if the field type is different
				if expected.Type != actual.Type {
					return nil, fmt.Errorf("field type is different. "+
						"Expected: %s, Actual: %s", expected.Type, actual.Type)
				}
				break
			}
		}
		if !found {
			// 2.3 Add the not exists field QL
			q := expected.BuildAddTagFieldQL(tag.Name)
			addFieldQLs = append(addFieldQLs, q)
		}
	}
	// 2.4 Execute the add field QLs if needed
	if len(addFieldQLs) > 0 {
		queries := strings.Join(addFieldQLs, " ")
		if mgr.verbose {
			log.Printf("ApplyTag: add the not existing fields. name=%s queries=%s\n", tag.Name, queries)
		}
		_, err = mgr.pool.ExecuteAndCheck(queries)
		if err != nil {
			return nil, err
		}
	}

	// 3. Remove the not expected field if needed
	// 3.1 Prepare the not expected fields
	dropFieldQLs := []string{}
	for _, actual := range fields {
		redundant := true
		for _, expected := range tag.Fields {
			if expected.Field == actual.Field {
				redundant = false
				break
			}
		}
		if redundant {
			// 3.2 Remove the not expected field
			q := actual.BuildDropTagFieldQL(tag.Name)
			dropFieldQLs = append(dropFieldQLs, q)
		}
	}
	// 3.3 Execute the drop field QLs if needed
	if len(dropFieldQLs) > 0 {
		queries := strings.Join(dropFieldQLs, " ")
		if mgr.verbose {
			log.Printf("ApplyTag: remove the not expected fields. name=%s queries=%s\n", tag.Name, queries)
		}
		_, err := mgr.pool.ExecuteAndCheck(queries)
		if err != nil {
			return nil, err
		}
	}

	// 4. Check if the TTL is set as expected.
	ttlCol, ttlDuration, err := mgr.pool.GetTagTTL(tag.Name)
	if err != nil {
		return nil, err
	}

	if ttlCol != tag.TTLCol || ttlDuration != tag.TTLDuration {
		if mgr.verbose {
			log.Printf(
				"ApplyTag: alter the tag TTL. name=%s, col from %s to %s, duration from %d to %d\n",
				tag.Name,
				ttlCol,
				tag.TTLCol,
				ttlDuration,
				tag.TTLDuration,
			)
		}

		_, err = mgr.pool.AddTagTTL(tag.Name, tag.TTLCol, tag.TTLDuration)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

// ApplyEdge applies the given edge to the graph.
// 1. If the edge does not exist, it will be created.
// 2. If the edge exists, it will be checked if the fields are the same.
// 2.1 If not, the new fields will be added.
// 2.2 If the field type is different, it will return an error.
// 2.3 If a field exists in the graph but not in the given edge,
// it will be removed.
// 3. If the edge exists and the fields are the same,
// it will be checked if the TTL is set as expected.
//
// Notice:
// We won't change the field type because it has
// unexpected behavior for the data.
func (mgr *SchemaManager) ApplyEdge(edge LabelSchema) (*ResultSet, error) {
	// 1. Make sure the edge exists
	fields, err := mgr.pool.DescEdge(edge.Name)
	if err != nil {
		// If the edge does not exist, create it
		if strings.Contains(err.Error(), ErrorEdgeNotFound) {
			if mgr.verbose {
				log.Printf("ApplyEdge: create the not existing edge. name=%s\n", edge.Name)
			}
			return mgr.pool.CreateEdge(edge)
		}
		return nil, err
	}

	// 2. The edge exists now, add new fields if needed
	// 2.1 Prepare the new fields
	addFieldQLs := []string{}
	for _, expected := range edge.Fields {
		found := false
		for _, actual := range fields {
			if expected.Field == actual.Field {
				found = true
				// 2.2 Check if the field type is different
				if expected.Type != actual.Type {
					return nil, fmt.Errorf("field type is different. "+
						"Expected: %s, Actual: %s", expected.Type, actual.Type)
				}
				break
			}
		}
		if !found {
			// 2.3 Add the not exists field QL
			q := expected.BuildAddEdgeFieldQL(edge.Name)
			addFieldQLs = append(addFieldQLs, q)
		}
	}
	// 2.4 Execute the add field QLs if needed
	if len(addFieldQLs) > 0 {
		queries := strings.Join(addFieldQLs, " ")
		if mgr.verbose {
			log.Printf("ApplyEdge: add the not existing fields. name=%s queries=%s\n", edge.Name, queries)
		}
		_, err := mgr.pool.ExecuteAndCheck(queries)
		if err != nil {
			return nil, err
		}
	}

	// 3. Remove the not expected field if needed
	// 3.1 Prepare the not expected fields
	dropFieldQLs := []string{}
	for _, actual := range fields {
		redundant := true
		for _, expected := range edge.Fields {
			if expected.Field == actual.Field {
				redundant = false
				break
			}
		}
		if redundant {
			// 3.2 Remove the not expected field
			q := actual.BuildDropEdgeFieldQL(edge.Name)
			dropFieldQLs = append(dropFieldQLs, q)
		}
	}
	// 3.3 Execute the drop field QLs if needed
	if len(dropFieldQLs) > 0 {
		queries := strings.Join(dropFieldQLs, "")
		if mgr.verbose {
			log.Printf("ApplyEdge: remove the not expected fields. name=%s queries=%s\n", edge.Name, queries)
		}
		_, err := mgr.pool.ExecuteAndCheck(queries)
		if err != nil {
			return nil, err
		}
	}

	// 4. Check if the TTL is set as expected.
	ttlCol, ttlDuration, err := mgr.pool.GetEdgeTTL(edge.Name)
	if err != nil {
		return nil, err
	}

	if ttlCol != edge.TTLCol || ttlDuration != edge.TTLDuration {
		if mgr.verbose {
			log.Printf(
				"ApplyEdge: alter the edge TTL. name=%s, col from %s to %s, duration from %d to %d\n",
				edge.Name,
				ttlCol,
				edge.TTLCol,
				ttlDuration,
				edge.TTLDuration,
			)
		}

		_, err = mgr.pool.AddEdgeTTL(edge.Name, edge.TTLCol, edge.TTLDuration)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}
