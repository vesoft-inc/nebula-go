package ngdb

import (
	"fmt"
	"testing"
)

const (
	address  = "127.0.0.1"
	port     = 6699
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
	} else {
		const expectedResult = `=============================================================================================
| Ip         | Port  | Status | Leader count | Leader distribution | Partition distribution |
=============================================================================================
| 172.28.1.2 | 44500 | online | 0            |                     |                        |
---------------------------------------------------------------------------------------------
`
		result := client.PrintResult(resp)
		if result != expectedResult {
			t.Errorf("\nexpected\n%s\nreal\n%s", expectedResult, result)
		}
	}
}
