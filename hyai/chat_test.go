package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestStreamChat(t *testing.T) {
	gogogo()
}

type RequestBody struct {
	Stream  bool     `json:"stream"`
	History []string `json:"history"`
	Query   string   `json:"query"`
}

func gogogo() {
	// 创建请求体
	body := RequestBody{
		Stream:  true,
		History: []string{},
		Query:   "分享一个笑话",
	}
	jsonBody, _ := json.Marshal(body)

	// 发送POST请求
	resp, err := http.Post("http://localhost:30033/hyai/query", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "POST request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// 流式接收响应 方法一
	buf := make([]byte, 256)
	for {
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "reading response body: %v\n", err)
			os.Exit(1)
		}
		if n == 0 {
			break
		}
		fmt.Print(string(buf[:n]), "\n")
	}
	// 流式接收响应方法二、
	//scanner := bufio.NewScanner(resp.Body)
	//for scanner.Scan() {
	//	fmt.Println(scanner.Text(), "--")
	//}
	//
	//if err := scanner.Err(); err != nil {
	//	fmt.Fprintf(os.Stderr, "reading response body: %v\n", err)
	//	os.Exit(1)
	//}
}
