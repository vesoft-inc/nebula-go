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
	"github.com/vesoft-inc/nebula-go/v3/nebula"
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

func TestSession_CreateSpace_ShowSpaces(t *testing.T) {
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

	newSpaceName := "new_created_space"
	conf := SpaceConf{
		Name:      newSpaceName,
		Partition: 1,
		Replica:   1,
		VidType:   "FIXED_STRING(12)",
	}

	_, err = sess.CreateSpace(conf)
	if err != nil {
		t.Fatal(err)
	}

	conf.IgnoreIfExists = true
	// Create again should work
	_, err = sess.CreateSpace(conf)
	if err != nil {
		t.Fatal(err)
	}

	spaceNames, err := sess.ShowSpaces()
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

func TestExecuteWithParameterAllTypes(t *testing.T) {
	params := map[string]interface{}{
		"p_bool":    true,
		"p_int":     42,
		"p_int64":   int64(9223372036854775807),
		"p_float32": float32(3.14),
		"p_float64": 3.14159265359,
		"p_string":  "nebula",
		"p_nil":     nil,
		"p_list":    []interface{}{1, 2, 3},
		"p_map":     map[string]interface{}{"key": "value"},
		"p_date":    nebula.Date{Year: 2023, Month: 5, Day: 15},
		"p_time":    nebula.Time{Hour: 14, Minute: 30, Sec: 45, Microsec: 123456},
		"p_datetime": nebula.DateTime{
			Year: 2023, Month: 5, Day: 15,
			Hour: 14, Minute: 30, Sec: 45, Microsec: 123456,
		},
	}

	query := `
		RETURN 
			$p_bool AS bool_val,
			$p_int AS int_val,
			$p_int64 AS int64_val,
			$p_float32 AS float32_val,
			$p_float64 AS float64_val,
			$p_string AS string_val,
			$p_nil AS nil_val,
			$p_list AS list_val,
			$p_map AS map_val,
			date($p_date) AS date_val,
			time($p_time) AS time_val,
			datetime($p_datetime) AS datetime_val
	`

	resultSet, err := session.ExecuteWithParameter(query, params)
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	if !resultSet.IsSucceed() {
		t.Fatalf("Query execution failed: %s", resultSet.GetErrorMsg())
	}

	record, err := resultSet.GetRowValuesByIndex(0)
	if err != nil {
		t.Fatalf("Failed to get row values: %v", err)
	}

	// Check bool
	boolVal, err := record.GetBool("bool_val")
	assert.NoError(t, err)
	assert.Equal(t, true, boolVal)

	// Check int
	intVal, err := record.GetInt("int_val")
	assert.NoError(t, err)
	assert.Equal(t, 42, intVal)

	// Check int64
	int64Val, err := record.GetInt64("int64_val")
	assert.NoError(t, err)
	assert.Equal(t, int64(9223372036854775807), int64Val)

	// Check float32
	float32Val, err := record.GetFloat32("float32_val")
	assert.NoError(t, err)
	assert.InDelta(t, float32(3.14), float32Val, 0.0001)

	// Check float64
	float64Val, err := record.GetFloat64("float64_val")
	assert.NoError(t, err)
	assert.InDelta(t, 3.14159265359, float64Val, 0.0000000001)

	// Check string
	stringVal, err := record.GetString("string_val")
	assert.NoError(t, err)
	assert.Equal(t, "nebula", stringVal)

	// Check nil
	nilVal, err := record.GetValue("nil_val")
	assert.NoError(t, err)
	assert.Nil(t, nilVal)

	// Check list
	listVal, err := record.GetList("list_val")
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{int64(1), int64(2), int64(3)}, listVal)

	// Check map
	mapVal, err := record.GetMap("map_val")
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"key": "value"}, mapVal)

	// Check date
	dateVal, err := record.GetDate("date_val")
	assert.NoError(t, err)
	assert.Equal(t, nebula.Date{Year: 2023, Month: 5, Day: 15}, dateVal)

	// Check time
	timeVal, err := record.GetTime("time_val")
	assert.NoError(t, err)
	assert.Equal(t, nebula.Time{Hour: 14, Minute: 30, Sec: 45, Microsec: 123456}, timeVal)

	// Check datetime
	datetimeVal, err := record.GetDateTime("datetime_val")
	assert.NoError(t, err)
	assert.Equal(t, nebula.DateTime{
		Year: 2023, Month: 5, Day: 15,
		Hour: 14, Minute: 30, Sec: 45, Microsec: 123456,
	}, datetimeVal)
}
