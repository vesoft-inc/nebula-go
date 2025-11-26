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

import "strings"

type nebulaVersion string

const (
	nebulaVersionV5_0 = nebulaVersion("5.0")
	nebulaVersionV5_1 = nebulaVersion("5.1")
	// current means the latest version
	nebulaVersionCurrent = nebulaVersion("current")
)

var nebulaVersions = map[nebulaVersion]struct{}{
	nebulaVersionCurrent: {},
	nebulaVersionV5_0:    {},
	nebulaVersionV5_1:    {},
}

func parseNebulaVersion(version string) nebulaVersion {
	if version == "" {
		return nebulaVersionV5_0
	}
	version = strings.TrimLeft(version, "vV")
	for k := range nebulaVersions {
		if strings.HasPrefix(version, string(k)) {
			return k
		}
	}
	return nebulaVersionCurrent
}
