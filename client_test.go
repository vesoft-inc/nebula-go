package ngdb

import (
	"fmt"
	"testing"
)

const (
	address  = "127.0.0.1"
	port     = 3699
	username = "user"
	password = "password"
)

func TestClient(t *testing.T) {
	client, err := NewClient(fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		t.Errorf("Fail to create client, address: %s, port: %d", address, port)
	}

	if err = client.Connect(username, password); err != nil {
		t.Errorf("Fail to connect server, username: %s, password: %s", username, password)
	}

	defer client.Disconnect()

	if resp, err := client.Execute("SHOW HOSTS;"); err != nil {
		t.Errorf("ErrorCode: %v, ErrorMsg: %s", resp.GetErrorCode(), resp.GetErrorMsg())
	}
}
