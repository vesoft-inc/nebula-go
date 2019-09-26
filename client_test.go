package ngdb

import (
	"fmt"
	"testing"

	"github.com/vesoft-inc/nebula-go/graph"
)

const (
	address  = "127.0.0.1"
	port     = 6699
	username = "user"
	password = "password"
)

func TestClient(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping client test in short mode")
	}

	client, err := NewClient(fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		t.Errorf("Fail to create client, address: %s, port: %d, %s", address, port, err.Error())
	}

	if err = client.Connect(username, password); err != nil {
		t.Errorf("Fail to connect server, username: %s, password: %s, %s", username, password, err.Error())
	}

	defer client.Disconnect()

	if resp, err := client.Execute("SHOW HOSTS;"); err != nil {
		t.Errorf(err.Error())
	} else {
		result := client.PrintResult(resp)
		t.Logf("show hosts:\n%s\nErrorCode: %v, ErrorMsg: %s", result, resp.GetErrorCode(), resp.GetErrorMsg())
	}

	if resp, err := client.Execute("CREATE SPACE client_test(partition_num=1024, replica_factor=1);"); err != nil {
		t.Error(err.Error())
	} else {
		t.Logf("create space response: %s", client.PrintResult(resp))
		if resp.GetErrorCode() != graph.ErrorCode_SUCCEEDED {
			t.Logf("create space: ErrorCode: %v, ErrorMsg: %s", resp.GetErrorCode(), resp.GetErrorMsg())
		}
	}
}
