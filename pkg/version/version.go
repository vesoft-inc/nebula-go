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
package version

import (
	"runtime/debug"
	"strings"
)

const pkgPath = "github.com/vesoft-inc/nebula-ng-golang"

var ClientVersion string

// getVersion from buildInfo. e.g. go.mod in application:
// require github.com/vesoft-inc/nebula-ng-golang v5.0.0

// then the version should be v5.0.0
// if the module is replaced, the version should be (devel)
func getVersion() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, dep := range buildInfo.Deps {
		if strings.HasPrefix(strings.TrimSpace(dep.Path), pkgPath) {
			if dep.Replace != nil {
				return dep.Replace.Version
			} else {
				return dep.Version
			}
		}
	}
	return ""
}

func init() {
	ClientVersion = getVersion()
}
