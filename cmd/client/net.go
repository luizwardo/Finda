package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/wrdo/FInda/internal/protocol"
)

func queryServer(addr string, from, to string, timeout time.Duration) (protocol.Response, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return protocol.Response{}, fmt.Errorf("connect %s: %w", addr, err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))

	req := protocol.Request{
		Op:   protocol.OpPathfind,
		From: from,
		To:   to,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return protocol.Response{}, err
	}
	if _, err := fmt.Fprintln(conn, string(payload)); err != nil {
		return protocol.Response{}, err
	}

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return protocol.Response{}, err
		}
		return protocol.Response{}, fmt.Errorf("empty response from %s", addr)
	}

	var resp protocol.Response
	if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
		return protocol.Response{}, err
	}
	return resp, nil
}
