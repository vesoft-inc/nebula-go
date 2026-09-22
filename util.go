// Copyright 2025 vesoft inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package nebula_ng

import (
	"net"
	"slices"
	"strconv"
	"strings"

	"github.com/vesoft-inc/nebula-go/v5/internal/internal_error"
	"github.com/vesoft-inc/nebula-go/v5/pkg/errors"
)

func parseHostPort(address string) (string, int, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", 0, internal_error.ErrAddressNotValid(address, err.Error())
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return "", 0, internal_error.ErrAddressNotValid(address, err.Error())
	}

	return host, p, nil
}

func parseAddresses(addresses string) ([]string, error) {
	var hostAddresses []string
	addrs := strings.SplitSeq(addresses, ",")
	for addr := range addrs {
		if addr == "" {
			continue
		}
		host, port, err := parseHostPort(addr)
		if err != nil {
			return nil, err
		}

		hostAddresses = append(hostAddresses, net.JoinHostPort(host, strconv.Itoa(port)))
	}
	return hostAddresses, nil
}

func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	ne, ok := err.(*errors.NebulaError)
	if !ok {
		return false
	}
	codes := []errors.ErrorCode{
		errors.ERROR_CONN_UNAVAILABLE,
		errors.ERROR_CONN_CONNECT_TIMEOUT,
		errors.ERROR_CONN_REQUEST_TIMEOUT,
		errors.ERROR_CONN_IS_CLOSED,
	}
	return slices.Contains(codes, ne.Code())
}
