//go:build integration
// +build integration

/* Copyright (c) 2021 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 */

package nebula_go

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

		// test Ping()
		err = s.Ping()
		if err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				break
			default:
				f(sess)
			}
		}
	}(ctx)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				break
			default:
				f(sess)
			}
		}
	}(ctx)
	time.Sleep(300 * time.Millisecond)

}

func TestSession_Recover(t *testing.T) {
	query := "show hosts"
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
	assert.Equal(t, 1, pool.getActiveConnCount()+pool.getIdleConnCount())
	go func() {
		for {
			_, _ = sess.Execute(query)
		}
	}()
	stopContainer(t, "nebula-docker-compose_graphd_1")
	stopContainer(t, "nebula-docker-compose_graphd1_1")
	stopContainer(t, "nebula-docker-compose_graphd2_1")
	defer func() {
		startContainer(t, "nebula-docker-compose_graphd1_1")
		startContainer(t, "nebula-docker-compose_graphd2_1")
	}()
	<-time.After(3 * time.Second)
	startContainer(t, "nebula-docker-compose_graphd_1")
	<-time.After(3 * time.Second)
	_, err = sess.Execute(query)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, pool.getActiveConnCount()+pool.getIdleConnCount())
}
