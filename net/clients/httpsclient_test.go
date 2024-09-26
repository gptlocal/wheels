package clients

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"testing"
)

func TestHTTPSClient(t *testing.T) {
	targetHost := "www.baidu.com"
	// 建立基本的TCP连接
	conn, err := net.Dial("tcp", targetHost+":443")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer conn.Close()

	// 设置TLS配置
	tlsConfig := &tls.Config{
		ServerName: "www.baidu.com",
	}

	// 将TCP连接升级到TLS连接
	tlsConn := tls.Client(conn, tlsConfig)
	err = tlsConn.Handshake()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer tlsConn.Close()

	// 构建HTTP请求
	request := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", "/", targetHost)

	// 发送请求
	_, err = tlsConn.Write([]byte(request))
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// 读取响应
	reader := bufio.NewReader(tlsConn)
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	fmt.Println("Status:", statusLine)

	// 读取并忽略响应头
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		fmt.Printf("Header: %s", line)
		if line == "\r\n" {
			break
		}
	}

	// 读取响应主体
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	fmt.Println("Body:", string(body))
}
