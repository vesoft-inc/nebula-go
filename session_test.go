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
	errCh := make(chan error, 1)

	f := func(s *Session) {
		time.Sleep(10 * time.Microsecond)
		reps, err := s.Execute("yield 1")
		if err != nil {
			errCh <- err
		}
		if !reps.IsSucceed() {
			t.Fatal(reps.resp.ErrorMsg)
		}

		// test Ping()
		err = s.Ping()
		if err != nil {
			errCh <- err
		}
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 300*time.Millisecond)
	defer cancel()
	go func(ctx context.Context) {
		sess, err := pool.GetSession("root", "nebula")
		if err != nil {
			errCh <- err
		}
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
		sess, err := pool.GetSession("root", "nebula")
		if err != nil {
			errCh <- err
		}
		for {
			select {
			case <-ctx.Done():
			default:
				f(sess)
			}
		}
	}(ctx)

	for {
		select {
		case err := <-errCh:
			t.Fatal(err)
		case <-ctx.Done():
			return
		}
	}

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
	stopContainer(t, "nebula-docker-compose_graphd0_1")
	stopContainer(t, "nebula-docker-compose_graphd1_1")
	stopContainer(t, "nebula-docker-compose_graphd2_1")
	defer func() {
		startContainer(t, "nebula-docker-compose_graphd1_1")
		startContainer(t, "nebula-docker-compose_graphd2_1")
	}()
	<-time.After(3 * time.Second)
	startContainer(t, "nebula-docker-compose_graphd0_1")
	<-time.After(3 * time.Second)
	_, err = sess.Execute(query)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, pool.getActiveConnCount()+pool.getIdleConnCount())
}

func TestSession_ShowSpaces(t *testing.T) {
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

	spaceNames, err := sess.ShowSpaces()
	if err != nil {
		t.Fatal(err)
	}
	assert.LessOrEqual(t, 1, len(spaceNames))
}
