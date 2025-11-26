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
// go run scan_example.go                                                                                                                                    fix_golang [bd83b66] deleted modified added untracked
// graph name: sf1, node type: Person
// graph name: sf1, node type: Person
// graph name: sf1, node type: Person
// graph name: sf1, node type: Person
// graph name: sf1, node type: Person
// birthday: 1987-9-1, first name: Ben
// birthday: 1987-4-1, first name: Alexander
// birthday: 1986-5-23, first name: Li
// birthday: 1985-7-9, first name: Lin
// birthday: 1988-11-4, first name: Cornelis

package main

import (
	"fmt"

	nebula "github.com/vesoft-inc/nebula-go/v5"
)

const (
	address  = "127.0.0.1"
	port     = 10015
	username = "root"
	password = "NebulaGraph01"
)

func exitWithError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

	// Initialize logger
	var log = nebula.DefaultLogger
	addresses := fmt.Sprintf("%s:%d", address, port)
	client, err := nebula.NewNebulaClient(addresses, username, password)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create client: %v", err))
		return
	}
	defer client.Close()
	log.Info("Client created successfully")

	// Execute a query and then print the node info
	query := "use sf1 match(v:Person) return v limit 5"
	resp, err := client.Execute(query)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to execute query: %v", err))
		return
	}
	for resp.HasNext() {
		var v nebula.NullNode
		if err := resp.Scan(&v); err != nil {
			log.Error(fmt.Sprintf("Failed to scan row: %v", err))
			return
		}
		if v.Valid {
			fmt.Printf("graph name: %s, ", v.Data.GetGraph())
			fmt.Printf("node type: %s ", v.Data.GetType())
			fmt.Printf("\n")
		}
	}
	// Execute a query and then print the properties
	type Person struct {
		id        nebula.NullInt64
		birthday  nebula.NullDate
		firstName nebula.NullString
	}
	query = "use sf1 match(v:Person) return v.id, v.birthday, v.firstName limit 5"
	resp, err = client.Execute(query)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to execute query: %v", err))
		return
	}
	for resp.HasNext() {
		var p Person
		if err := resp.Scan(&p.id, &p.birthday, &p.firstName); err != nil {
			log.Error(fmt.Sprintf("Failed to scan row: %v", err))
			return
		}
		if p.id.Valid && p.birthday.Valid && p.firstName.Valid {
			fmt.Printf("id: %d, ", p.id.Data)
			fmt.Printf("birthday: %d-%d-%d, ", p.birthday.Data.GetYear(),
				p.birthday.Data.GetMonth(), p.birthday.Data.GetDay())
			fmt.Printf("first name: %s\n", p.firstName.Data)
		}
	}
}
