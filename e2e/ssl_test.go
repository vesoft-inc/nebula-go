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

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	nebula_ng "github.com/vesoft-inc/nebula-go/v5"
	"github.com/vesoft-inc/nebula-go/v5/pkg/errors"
)

func TestTLS(t *testing.T) {
	addr := fmt.Sprintf("%s:%d", nebulaHost, nebulaSSLPort)
	c, err := nebula_ng.NewNebulaClient(addr, nebulaUser, nebulaPassword,
		nebula_ng.WithClientTLS(
			"./docker-compose-ssl/certs/ca.crt",
			"./docker-compose-ssl/certs/client.crt",
			"./docker-compose-ssl/certs/client.key",
			false,
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_, err = c.Execute("return 1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestPoolTLS(t *testing.T) {
	addr := fmt.Sprintf("%s:%d", nebulaHost, nebulaSSLPort)
	p, err := nebula_ng.NewNebulaPool(addr, nebulaUser, nebulaPassword,
		nebula_ng.WithPoolTLS(
			"./docker-compose-ssl/certs/ca.crt",
			"./docker-compose-ssl/certs/client.crt",
			"./docker-compose-ssl/certs/client.key",
			false,
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	c, err := p.GetClient()
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Execute("return 1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestTLSWrongCerts(t *testing.T) {
	addr := fmt.Sprintf("%s:%d", nebulaHost, nebulaSSLPort)
	testcases := []struct {
		name               string
		ca                 string
		cert               string
		key                string
		insecureSkipVerify bool
		expectedConnect    bool
	}{
		{
			"wrong certs",
			"./docker-compose-ssl/certs.wrong/ca.crt",
			"./docker-compose-ssl/certs.wrong/client.crt",
			"./docker-compose-ssl/certs.wrong/client.key",
			false,
			false,
		},
		{
			"wrong ca",
			"./docker-compose-ssl/certs.wrong/ca.crt",
			"./docker-compose-ssl/certs/client.crt",
			"./docker-compose-ssl/certs/client.key",
			true,
			true,
		},
		{
			"no client cert",
			"",
			"",
			"",
			true,
			false,
		},
	}
	for _, tc := range testcases {
		c, err := nebula_ng.NewNebulaClient(addr, nebulaUser, nebulaPassword,
			nebula_ng.WithClientTLS(
				tc.ca,
				tc.cert,
				tc.key,
				tc.insecureSkipVerify,
			),
		)
		if !tc.expectedConnect {
			assert.NotNil(t, err, fmt.Sprintf("test case %s should fail", tc.name))
			ngErr, ok := err.(*errors.NebulaError)
			if !ok {
				t.Fatalf("not a NebulaError, err is %v", err)
			}
			assert.Equal(t, ngErr.Code(), errors.ERROR_CONN_UNAVAILABLE)
		} else {
			assert.Nil(t, err, fmt.Sprintf("test case %s should success", tc.name))
			defer c.Close()
			_, err = c.Execute("return 1")
			assert.Nil(t, err)
		}
	}
}

func TestTLSMismatch(t *testing.T) {
	addr := fmt.Sprintf("%s:%d", nebulaHost, nebulaSSLPort)
	_, err := nebula_ng.NewNebulaClient(addr, nebulaUser, nebulaPassword)
	assert.NotNil(t, err)
	ngErr, ok := err.(*errors.NebulaError)
	if !ok {
		t.Fatalf("not a NebulaError, err is %v", err)
	}
	assert.Equal(t, ngErr.Code(), errors.ERROR_CONN_UNAVAILABLE)
	assert.Equal(t, ngErr.Error(), "[99002]: connection to 127.0.0.1:9528 is unavailable, rpc error: code = Unavailable desc = connection error: desc = \"error reading server preface: EOF\"")
}
