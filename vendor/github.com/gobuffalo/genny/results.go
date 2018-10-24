package genny

import (
	"net/http"
	"os/exec"
)

type Results struct {
	Files    []File
	Commands []*exec.Cmd
	Requests []RequestResult
}

type RequestResult struct {
	Request  *http.Request
	Response *http.Response
	Client   *http.Client
	Error    error
}
