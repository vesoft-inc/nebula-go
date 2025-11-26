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
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	nebula "github.com/vesoft-inc/nebula-go/v5"
	"github.com/vesoft-inc/nebula-go/v5/pkg/types"
)

const (
	nebulaHost           = "127.0.0.1"
	nebulaPort           = 9527
	nebulaSSLPort        = 9528
	nebulaUser           = "root"
	nebulaPassword       = "NebulaGraph01"
	dockerComposeFile    = "docker-compose/docker-compose.yaml"
	dockerComposeSSLFile = "docker-compose/docker-compose-ssl.yaml"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

var nebulaAddress = fmt.Sprintf("%s:%d", nebulaHost, nebulaPort)

func prepareData(ctx context.Context, c types.Client, scanner *bufio.Scanner) error {
	var stmt string
	var multiLine bool
	var execute bool
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == `"""` {
			if multiLine {
				multiLine = false
				execute = true
			} else {
				multiLine = true
			}
		} else {
			if multiLine {
				stmt += line + "\n"
			} else {
				stmt = line
				execute = true
			}
		}
		if execute {
			execute = false
			var err error
			for {
				if ctx.Err() != nil {
					return fmt.Errorf("context error, %s", ctx.Err().Error())
				}
				if _, err = c.ExecuteContext(ctx, stmt); err == nil {
					break
				}
				time.Sleep(300 * time.Millisecond)
			}
		}
	}
	return nil
}

func importData(ctx context.Context) error {
	file, err := os.Open("../movie.ngql")
	if err != nil {
		return fmt.Errorf("open file failed, %s", err.Error())
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	c, err := nebula.NewNebulaClient(nebulaAddress, nebulaUser, nebulaPassword)
	if err != nil {
		return fmt.Errorf("NewNebulaClient failed, %s", err.Error())
	}
	defer c.Close()
	if err := prepareData(ctx, c, scanner); err != nil {
		return fmt.Errorf("prepareData failed, %s", err.Error())
	}
	_, err = c.ExecuteContext(ctx, "create schema \"/test_schema\"")
	if err != nil {
		return fmt.Errorf("create schema failed, %s", err.Error())
	}
	_, err = c.ExecuteContext(ctx, "session set schema \"/test_schema\"")
	if err != nil {
		return fmt.Errorf("session set schema failed, %s", err.Error())
	}
	_, err = c.ExecuteContext(ctx, `
	CREATE GRAPH TYPE test_graph_type AS {
		Node Actor (:Person {id INT PRIMARY KEY, name STRING, birthDate Date})
	}`)
	if err != nil {
		return fmt.Errorf("execute failed, %s", err.Error())
	}
	_, err = c.ExecuteContext(ctx, `CREATE GRAPH test_graph TYPED test_graph_type`)
	if err != nil {
		return fmt.Errorf("execute failed, %s", err.Error())
	}
	return nil
}

func main() {
	maxWaitTime := 60 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), maxWaitTime)
	defer cancel()
L:
	for {
		select {
		case <-ctx.Done():
			panic("wait nebula server start timeout")
		default:
			c, err := nebula.NewNebulaClient(nebulaAddress, nebulaUser, nebulaPassword)
			if err == nil && c.Ping() == nil {
				c.Close()
				break L
			}
			fmt.Println("wait nebula server start...")
			time.Sleep(2 * time.Second)
		}
	}
	//wait for storage leader change
	time.Sleep(3 * time.Second)
	end, _ := ctx.Deadline()
	fmt.Println("current time:", time.Now().Format(time.RFC3339))
	fmt.Println("start import data, deadline:", end.Format(time.RFC3339))
	if err := importData(ctx); err != nil {
		panic(err)
	}
}
