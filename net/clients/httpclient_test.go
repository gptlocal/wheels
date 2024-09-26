package clients

import (
	"net"
	"testing"
)

func TestHTTPClient(t *testing.T) {
	conn, err := net.Dial("tcp", "www.baidu.com:80")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("GET / HTTP/1.1\r\nHost: www.baidu.com\r\n\r\n")); err != nil {
		t.Fatalf("err: %v", err)
	}

	recv := make([]byte, 4096)
	if _, err := conn.Read(recv); err != nil {
		t.Fatalf("err: %v", err)
	}

	t.Logf("recv:\n%s", recv)
}
