// Copyright 2025 vesoft inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// 
//     http://www.apache.org/licenses/LICENSE-2.0
// 
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package decode

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
)

type dummpBatch struct {
	name string
	rows uint32
}

var _ batcher = &dummpBatch{}

func (b *dummpBatch) numRecords() uint32 {
	return b.rows
}
func (b *dummpBatch) getRowByIndex(index uint32, values []types.Value) error {
	if index >= b.rows {
		return io.EOF
	}
	return nil
}

func TestTable(t *testing.T) {
	testcases := []struct {
		name         string
		batches      int
		rowsPerBatch int
		nextTimes    int
		hasErr       bool
	}{
		{"1st", 1, 1, 1, false},
		{"2nd", 1, 1, 2, true},
		{"3rd", 1, 2, 2, false},
		{"4th", 1, 2, 3, true},
		{"5th", 2, 1, 1, false},
		{"6th", 2, 1, 2, false},
		{"7th", 2, 1, 3, true},
		{"8th", 2, 2, 1, false},
		{"9th", 2, 2, 4, false},
		{"10th", 2, 2, 5, true},
		{"11th", 100, 100, 9999, false},
		{"12th", 100, 100, 10000, false},
		{"13th", 100, 100, 10001, true},
	}
	for _, tc := range testcases {
		batches := make([]batcher, tc.batches)
		for i := 0; i < tc.batches; i++ {
			batches[i] = &dummpBatch{name: fmt.Sprintf("%d", i), rows: uint32(tc.rowsPerBatch)}
		}
		tbl := &ResultTable{
			batches:    batches,
			numBatches: tc.batches,
		}
		for i := 0; i < tc.nextTimes-1; i++ {
			_, err := tbl.Next()
			if err != nil {
				t.Fatal(err)
			}
		}
		// the last batch name should be tc.batches-1
		lastBatchName := fmt.Sprintf("%d", tc.batches-1)
		b, ok := tbl.batches[tc.batches-1].(*dummpBatch)
		if !ok {
			t.Fatal("type assertion failed")
		}
		assert.Equal(t, lastBatchName, b.name)
		_, err := tbl.Next()
		if tc.hasErr {
			assert.Error(t, err, tc.name)
		} else {
			assert.NoError(t, err, tc.name)
		}
	}
}
