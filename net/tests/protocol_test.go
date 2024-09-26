package tests

import (
	"bytes"
	"fmt"
	"net"
	"testing"
)

func TestPassthrough(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	cliResult := make(chan error)
	go func() {
		conn, e := net.Dial("tcp", l.Addr().String())
		if e != nil {
			cliResult <- e
			return
		}
		defer conn.Close()

		if _, e2 := conn.Write([]byte("ping")); e2 != nil {
			cliResult <- e2
			return
		}

		recv := make([]byte, 4)
		if _, err = conn.Read(recv); err != nil {
			cliResult <- err
			return
		}

		if !bytes.Equal(recv, []byte("pong")) {
			cliResult <- fmt.Errorf("bad: %v", recv)
			return
		}
		close(cliResult)
	}()

	conn, err := l.Accept()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer conn.Close()

	recv := make([]byte, 4)
	_, err = conn.Read(recv)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bytes.Equal(recv, []byte("ping")) {
		t.Fatalf("bad: %v", recv)
	}

	if _, err := conn.Write([]byte("pong")); err != nil {
		t.Fatalf("err: %v", err)
	}
	err = <-cliResult
	if err != nil {
		t.Fatalf("client error: %v", err)
	}
}
