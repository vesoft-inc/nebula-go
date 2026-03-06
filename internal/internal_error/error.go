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
package internal_error

import (
	"fmt"
	"time"

	"github.com/vesoft-inc/nebula-go/v5/pkg/errors"
)

func ErrAddressNotValid(address string, msg string) error {
	var (
		format string
		args   []any
	)

	if msg == "" {
		format = "address %s is not valid"
		args = []any{address}
	} else {
		format = "address %s is not valid, %s"
		args = []any{address, msg}
	}
	return errors.NewNebulaError(errors.ERROR_ADDRESS_NOT_VALID, format, args...)
}

func ErrConnCannotOpen(address string, msg string) error {
	return errors.NewNebulaError(errors.ERROR_CANNOT_OPEN, "cannot open connection to %s, %s", address, msg)
}

func ErrConnUnavailable(address string, msg string) error {
	if msg != "" {
		return errors.NewNebulaError(errors.ERROR_CONN_UNAVAILABLE, "connection to %s is unavailable, %s", address, msg)
	}
	return errors.NewNebulaError(errors.ERROR_CONN_UNAVAILABLE, "connection to %s is unavailable", address)
}

func ErrConnBroken(address string) error {
	return errors.NewNebulaError(errors.ERROR_CONN_UNAVAILABLE, "connection to %s is broken", address)
}

func ErrConnConnectTimeout(address string, msg string) error {
	if msg != "" {
		return errors.NewNebulaError(errors.ERROR_CONN_CONNECT_TIMEOUT, "connection to %s timeout, %s", address, msg)
	}
	return errors.NewNebulaError(errors.ERROR_CONN_CONNECT_TIMEOUT, "connection to %s timeout", address)

}

func ErrConnRequestTimeout(address string, msg string, duration time.Duration) error {
	var tpl string
	if duration == 0 {
		tpl = fmt.Sprintf("request to %s timeout", address)
	} else {
		d := duration / time.Millisecond
		if duration%time.Millisecond != 0 {
			d++
		}
		tpl = fmt.Sprintf("request to %s timeout after %dms", address, d)
	}
	if msg != "" {
		return errors.NewNebulaError(errors.ERROR_CONN_REQUEST_TIMEOUT, "%s, %s", tpl, msg)
	}
	return errors.NewNebulaError(errors.ERROR_CONN_REQUEST_TIMEOUT, tpl)
}

func ErrConnIsClosed(address string) error {
	return errors.NewNebulaError(errors.ERROR_CONN_IS_CLOSED, "connection to %s is closed", address)
}

func ErrWaitPoolTimeout() error {
	return errors.NewNebulaError(errors.ERROR_WAIT_POOL_TIMEOUT, "get from pool timeout")

}

func ErrIllegal(msg string) error {
	return errors.NewNebulaError(errors.ERROR_ILLEGAL, "Illegal, %s", msg)
}

func ErrType(msg string) error {
	return errors.NewNebulaError(errors.ERROR_TYPE, "Type error, %s", msg)
}

// client internal error
// user should not see this error
func ErrInternal(msg string) error {
	return errors.NewNebulaError(errors.ERROR_CLIENT_INTERNAL, "Internal error, %s", msg)
}

func ErrServerResponse(code string, msg string) error {
	return errors.NewNebulaError(errors.ErrorCode(code), "%s", msg)
}

func ErrorFromBytes(c []byte) errors.ErrorCode {
	return errors.ErrorCode(c)
}

func ErrTLS(msg string) error {
	return errors.NewNebulaError(errors.ERROR_TLS_ERROR, "TLS error, %s", msg)
}

func ErrDecodeFailed(msg string) error {
	return errors.NewNebulaError(errors.ERROR_DECODE_FAILED, "decode failed, %s", msg)
}
