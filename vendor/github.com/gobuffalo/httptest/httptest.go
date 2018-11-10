/*
Package httptest was formerly known as [https://github.com/markbates/willie](https://github.com/markbates/willie). It is used to test HTTP applications easily.
*/
package httptest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

// map the std httptest package for ease
const DefaultRemoteAddr = httptest.DefaultRemoteAddr

var NewServer = httptest.NewServer
var NewRequest = httptest.NewRequest
var NewRecorder = httptest.NewRecorder
var NewTLSServer = httptest.NewTLSServer
var NewUnstartedServer = httptest.NewUnstartedServer

type ResponseRecorder = httptest.ResponseRecorder
type Server = httptest.Server

type encodable interface {
	Encode() string
}

type Handler struct {
	http.Handler
	Cookies    string
	Headers    map[string]string
	HmaxSecret string
}

func (w *Handler) HTML(u string, args ...interface{}) *Request {
	hs := map[string]string{}
	for key, val := range w.Headers {
		hs[key] = val
	}
	return &Request{
		URL:     fmt.Sprintf(u, args...),
		handler: w,
		Headers: hs,
	}
}

func (w *Handler) JSON(u string, args ...interface{}) *JSON {
	hs := map[string]string{}
	for key, val := range w.Headers {
		hs[key] = val
	}
	hs["Accept"] = "application/json"
	return &JSON{
		URL:     fmt.Sprintf(u, args...),
		handler: w,
		Headers: hs,
	}
}

func (w *Handler) XML(u string, args ...interface{}) *XML {
	hs := map[string]string{}
	for key, val := range w.Headers {
		hs[key] = val
	}
	hs["Accept"] = "application/xml"
	return &XML{
		URL:     fmt.Sprintf(u, args...),
		handler: w,
		Headers: hs,
	}
}

func New(h http.Handler) *Handler {
	return &Handler{
		Handler: h,
		Headers: map[string]string{},
	}
}
