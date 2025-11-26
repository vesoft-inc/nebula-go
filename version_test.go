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

import "testing"

func TestVersion(t *testing.T) {
	testcases := []struct {
		version string
		expect  nebulaVersion
	}{
		{"", nebulaVersionV5_0},
		{"v5.0.0", nebulaVersionV5_0},
		{"v5.0.1", nebulaVersionV5_0},
		{"v5.1.0", nebulaVersionV5_1},
		{"5.1.1", nebulaVersionV5_1},
	}
	for _, tc := range testcases {
		if v := parseNebulaVersion(tc.version); v != tc.expect {
			t.Errorf("getNebulaVersion(%s) = %s, expect %s", tc.version, v, tc.expect)
		}
	}
}
