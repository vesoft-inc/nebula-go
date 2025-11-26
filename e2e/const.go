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
package e2e

import "fmt"

const (
	nebulaHost           = "127.0.0.1"
	nebulaPort           = 9527
	nebulaSSLPort        = 9528
	nebulaMetaPort       = 9537
	nebulaUser           = "root"
	nebulaPassword       = "NebulaGraph01"
	dockerComposeFile    = "docker-compose/docker-compose.yaml"
	dockerComposeSSLFile = "docker-compose/docker-compose-ssl.yaml"
)

var nebulaAddress = fmt.Sprintf("%s:%d", nebulaHost, nebulaPort)
