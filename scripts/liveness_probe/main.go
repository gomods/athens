// Package main is a simple script for our CI/CD workflow
// that ensures our sidecar proxy is running before proceeding
package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

var goproxy = os.Getenv("GOPROXY")

func main() {
	timeout := time.After(time.Minute)
	for {
		select {
		case <-timeout:
			fmt.Println("liveness probe timed out")
			os.Exit(1)
		default:
		}
		isLive, err := probe()
		if err != nil {
			shouldPrintErr := true
			// connection-refused errors are expected, don't print them
			var opErr *net.OpError
			if errors.As(err, &opErr) && opErr.Op == "read" {
				shouldPrintErr = false
			}
			if shouldPrintErr {
				fmt.Println(err)
			}
		}
		if isLive {
			fmt.Println("proxy is live")
			return
		}
		time.Sleep(time.Second)
	}
}

func probe() (bool, error) {
	req, err := http.NewRequest(http.MethodGet, goproxy, nil)
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	return resp.StatusCode == http.StatusOK, nil
}
