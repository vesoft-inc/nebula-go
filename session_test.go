//go:build integration
// +build integration

/* Copyright (c) 2021 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 */

package nebula_go

import (
	"testing"
	"time"
)

func TestSession_Execute(t *testing.T) {
	config := GetDefaultConf()
	host := HostAddress{address, port}
	pool, err := NewConnectionPool([]HostAddress{host}, config, DefaultLogger{})
	if err != nil {
		t.Fatal(err)
	}

	sess, err := pool.GetSession("root", "nebula")
	if err != nil {
		t.Fatal(err)
	}

	f := func(s *Session) {
		time.Sleep(10 * time.Microsecond)
		reps, err := s.Execute("yield 1")
		if err != nil {
			t.Fatal(err)
		}
		if !reps.IsSucceed() {
			t.Fatal(reps.resp.ErrorMsg)
		}
	}
	go func() {
		for {
			f(sess)
		}
	}()
	go func() {
		for {
			f(sess)
		}
	}()
	time.Sleep(300 * time.Millisecond)
}
