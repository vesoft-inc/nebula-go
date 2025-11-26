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
package nebula_ng

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHost(t *testing.T) {
	testcases := []struct {
		addresses string
		expected  []string
		err       string
	}{
		{"127.0.0.1:9669,127.0.0.2:9669,127.0.0.3:9669,", []string{
			"127.0.0.1:9669",
			"127.0.0.2:9669",
			"127.0.0.3:9669",
		}, ""},
		{"127.0.0.1:9669,127.0.0.2:9669,127.0.0.3:9669a,", nil, `[99000]: address 127.0.0.3:9669a is not valid, strconv.Atoi: parsing "9669a": invalid syntax`},
		{"127.0.0.1:9669,127.0.0.2:9669,127.0.0.39669a,", nil, "[99000]: address 127.0.0.39669a is not valid, address 127.0.0.39669a: missing port in address"},
		{"harris:9669,", []string{
			"harris:9669",
		}, ""},
		{"[2001:0db8:85a3::8a2e:0370:7334]:9669", []string{
			"[2001:0db8:85a3::8a2e:0370:7334]:9669",
		}, ""},
	}
	for _, tc := range testcases {
		actual, err := parseAddresses(tc.addresses)
		if err != nil {
			assert.Equal(t, tc.err, err.Error())
		} else {
			assert.Equal(t, tc.err, "")
		}
		assert.Equal(t, tc.expected, actual)
	}
}
