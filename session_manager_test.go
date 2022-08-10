package nebula_go

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func GetManagerConfig(minPoolSize int, maxPoolSize int) ManagerConfig {
	return ManagerConfig{
		username: "root",
		password: "",
		addresses: []HostAddress{
			{
				Host: "127.0.0.1",
				Port: 9669,
			},
		},
		// schema : https://docs.nebula-graph.com.cn/3.1.0/2.quick-start/4.nebula-graph-crud/
		spaceName:  "basketballplayer",
		poolConfig: GetPoolConfig(minPoolSize, maxPoolSize),
	}
}

func GetPoolConfig(minPoolSize int, maxPoolSize int) PoolConfig {
	return PoolConfig{
		MinConnPoolSize: minPoolSize,
		MaxConnPoolSize: maxPoolSize,
		IdleTime:        0,
	}
}

func TestSessionManager(t *testing.T) {
	manager, err := NewSessionManager(GetManagerConfig(1, 2), DefaultLogger{})
	if err != nil {
		t.Fail()
		return
	}
	defer manager.Close()
	session, err := manager.GetSession()
	if err != nil {
		t.Fail()
		return
	}
	// schema : https://docs.nebula-graph.com.cn/3.1.0/2.quick-start/4.nebula-graph-crud/
	result, err := session.Execute("GO FROM \"player101\" OVER follow YIELD id($$);")
	if err != nil || !result.IsSucceed() {
		t.Fail()
		return
	}
	assert.True(t, len(result.GetRows()) != 0)
}

// go test -bench=SessionManager -benchtime=10000x -run=^a
func BenchmarkSessionManager(b *testing.B) {
	manager, err := NewSessionManager(GetManagerConfig(10, 20), DefaultLogger{})
	if err != nil {
		b.Fail()
		return
	}
	defer manager.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			session, err := manager.GetSession()
			if err != nil {
				fmt.Sprintf("get session err: %v", err)
				panic(err)
			}

			result, err := session.Execute("GO FROM \"player101\" OVER follow YIELD id($$)")
			if err != nil || !result.IsSucceed() {
				fmt.Sprintf("execute statment err: %v", err)
				session.Release()
				panic(err)
			}
			session.Release()
		}
	})
}

// go test -bench=SessionPool -benchtime=10000x -run=^a
func BenchmarkSessionPool(b *testing.B) {
	config := GetManagerConfig(10, 20)
	pool, err := NewConnectionPool(config.addresses, config.poolConfig, DefaultLogger{})
	if err != nil {
		b.Fail()
		return
	}
	defer pool.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			session, err := pool.GetSession(config.username, config.password)
			if err != nil {
				fmt.Sprintf("get session err: %v", err)
				panic(err)
			}

			result, err := session.Execute("USE basketballplayer;GO FROM \"player101\" OVER follow YIELD id($$);")
			if err != nil || !result.IsSucceed() {
				fmt.Sprintf("execute statment err: %v", err)
				session.Release()
				panic(err)
			}
			session.Release()
		}
	})
}
