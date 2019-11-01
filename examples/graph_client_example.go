/* Copyright (c) 2019 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License,
 * attached with Common Clause Condition 1.0, found in the LICENSES directory.
 */

package main

import (
	nebula "github.com/vesoft-inc/nebula-go"
	"github.com/vesoft-inc/nebula-go/gen-go/nebula/graph"
	"log"
)

func main() {
	client, err := nebula.NewGraphClient("127.0.0.1:3699")
	if err != nil {
		log.Fatal(err)
	}

	if err = client.Connect("user", "password"); err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect()

	if resp, err := client.Execute("SHOW HOSTS;"); err != nil {
		log.Fatal(err)
	} else {
		if resp.GetErrorCode() != graph.ErrorCode_SUCCEEDED {
			log.Printf("ErrorCode: %v, ErrorMsg: %s", resp.GetErrorCode(), resp.GetErrorMsg())
		}
	}
}
