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
package main

import (
	"fmt"

	nebula "github.com/vesoft-inc/nebula-go/v5"
)

const (
	address  = "127.0.0.1"
	port     = 16720
	username = "root"
	password = "nebula"
)

// Initialize logger
var log = nebula.DefaultLogger

func main() {
	fmt.Println("TLS example starts ...")
	addresses := fmt.Sprintf("%s:%d", address, port)
	client, err := nebula.NewNebulaClient(
		addresses,
		username,
		password,
		nebula.WithClientTLS("ca.crt", "client.crt", "client.key", false),
	)
	if err != nil {
		panic(err.Error())
	}
	defer client.Close()
	if err := client.Ping(); err != nil {
		panic(err.Error())
	}
}
