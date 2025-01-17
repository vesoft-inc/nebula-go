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

func TestSession_Execute_AfterContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	config := GetDefaultConf()
	host := HostAddress{address, port}
	pool, err := NewConnectionPool(ctx, []HostAddress{host}, config, DefaultLogger{})
	if err != nil {
		t.Fatal(err)
	}

	cancel()

	_, err = pool.GetSession(ctx, username, password)

	assert.Contains(t, err.Error(), "context canceled")
}

func TestSession_Execute(t *testing.T) {
	ctx := context.Background()

	config := GetDefaultConf()
	host := HostAddress{address, port}
	pool, err := NewConnectionPool(ctx, []HostAddress{host}, config, DefaultLogger{})
	if err != nil {
		t.Fatal(err)
	}
	errCh := make(chan error, 1)

	f := func(s *Session) {
		time.Sleep(10 * time.Microsecond)
		reps, err := s.Execute(ctx, "yield 1")
		if err != nil {
			errCh <- err
		}
		if !reps.IsSucceed() {
			t.Fatal(reps.resp.ErrorMsg)
		}

		// test Ping()
		err = s.Ping(ctx)
		if err != nil {
			errCh <- err
		}
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 300*time.Millisecond)
	defer cancel()
	go func(ctx context.Context) {
		sess, err := pool.GetSession(ctx, "root", "nebula")
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
		sess, err := pool.GetSession(ctx, "root", "nebula")
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
	ctx := context.Background()

	query := "show hosts"
	config := GetDefaultConf()
	host := HostAddress{address, port}
	pool, err := NewConnectionPool(ctx, []HostAddress{host}, config, DefaultLogger{})
	if err != nil {
		t.Fatal(err)
	}

	sess, err := pool.GetSession(ctx, "root", "nebula")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, pool.getActiveConnCount()+pool.getIdleConnCount())
	go func() {
		for {
			_, _ = sess.Execute(ctx, query)
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
	_, err = sess.Execute(ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, pool.getActiveConnCount()+pool.getIdleConnCount())
}

func TestSession_CreateSpace_ShowSpaces(t *testing.T) {
	ctx := context.Background()

	config := GetDefaultConf()
	host := HostAddress{address, port}
	pool, err := NewConnectionPool(ctx, []HostAddress{host}, config, DefaultLogger{})
	if err != nil {
		t.Fatal(err)
	}

	sess, err := pool.GetSession(ctx, "root", "nebula")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, pool.getActiveConnCount()+pool.getIdleConnCount())

	newSpaceName := "new_created_space"
	conf := SpaceConf{
		Name:      newSpaceName,
		Partition: 1,
		Replica:   1,
		VidType:   "FIXED_STRING(12)",
	}

	_, err = sess.CreateSpace(ctx, conf)
	if err != nil {
		t.Fatal(err)
	}

	conf.IgnoreIfExists = true
	// Create again should work
	_, err = sess.CreateSpace(ctx, conf)
	if err != nil {
		t.Fatal(err)
	}

	spaceNames, err := sess.ShowSpaces(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, space := range spaceNames {
		names = append(names, space.Name)
	}
	assert.LessOrEqual(t, 1, len(names))
	assert.Contains(t, names, newSpaceName)
}
