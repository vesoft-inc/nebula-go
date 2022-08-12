package nebula_go

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func GetManagerConfig(minPoolSize int, maxPoolSize int) ManagerConfig {
	return ManagerConfig{
		username: "root",
		password: "nebula",
		addresses: []HostAddress{
			{
				Host: "127.0.0.1",
				Port: 3699,
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
	err := InitData()
	if err != nil {
		t.Fatalf(err.Error())
	}
	manager, err := NewSessionManager(GetManagerConfig(1, 2), DefaultLogger{})
	if err != nil {
		t.Fatalf("create session manager error, %s", err.Error())
	}
	defer manager.Close()
	session, err := manager.GetSession()
	if err != nil {
		t.Fatalf("fail to get session from session manager, %s", err.Error())
	}
	// schema : https://docs.nebula-graph.com.cn/3.1.0/2.quick-start/4.nebula-graph-crud/
	result, err := session.Execute("GO FROM \"player101\" OVER follow YIELD id($$);")
	if err != nil || !result.IsSucceed() {
		t.Fatalf("execute statment fails")
	}
	assert.True(t, len(result.GetRows()) != 0)
}

func BenchmarkSessionManager(b *testing.B) {
	skipBenchmark(b)
	err := InitData()
	if err != nil {
		b.Fatalf(err.Error())
	}
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

func BenchmarkSessionPool(b *testing.B) {
	skipBenchmark(b)
	err := InitData()
	if err != nil {
		b.Fatalf(err.Error())
	}
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

func InitData() error {
	config := GetManagerConfig(10, 20)
	pool, err := NewConnectionPool(config.addresses, config.poolConfig, DefaultLogger{})
	if err != nil {
		return fmt.Errorf("test session manager, init data fail, %s", err.Error())
	}
	defer pool.Close()

	session, err := pool.GetSession("root", "nebula")
	defer session.Release()
	if err != nil {
		return fmt.Errorf("test session manager, init data, get session fail, %s", err.Error())
	}

	schema := "CREATE SPACE IF NOT EXISTS basketballplayer(partition_num=1, replica_factor=1, vid_type=fixed_string(30));" +
		"USE basketballplayer;" +
		"CREATE TAG IF NOT EXISTS player(name string, age int);" +
		"CREATE EDGE IF NOT EXISTS follow(degree int);"

	_, err = Execute(session, schema)
	if err != nil {
		return fmt.Errorf("test session manager, init data schema fail, %s", err.Error())
	}

	dataStatement := "INSERT VERTEX player(name, age) VALUES \"player101\":(\"Tony Parker\", 36), \"player100\":(\"Tim Duncan\", 42);"
	_, err = Execute(session, dataStatement)
	if err != nil {
		return fmt.Errorf("test session manager, init data fail, %s", err.Error())
	}

	time.Sleep(2 * time.Second)

	dataStatement = "INSERT EDGE follow(degree) VALUES \"player101\" -> \"player100\":(95);"
	_, err = Execute(session, dataStatement)
	if err != nil {
		return fmt.Errorf("test session manager, init data fail, %s", err.Error())
	}
	return nil
}

func skipBenchmark(b *testing.B) {
	if os.Getenv("session_manager_benchmark") != "true" {
		b.Skip("skip session manager benchmark testing")
	}
}

func Execute(session *Session, query string) (resp *ResultSet, err error) {
	for i := 3; i > 0; i-- {
		resp, err = session.Execute(query)
		if err == nil && resp.IsSucceed() {
			return
		}
		time.Sleep(2 * time.Second)
	}
	return
}
