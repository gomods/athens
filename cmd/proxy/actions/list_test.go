package actions

import (
	"io"
	"strings"
)

type testCase struct {
	version string
	mod     []byte
	zip     io.Reader
	info    []byte
}

func (a *ActionSuite) getReader() io.Reader {
	r := strings.NewReader("Go is a general-purpose language designed with systems programming in mind.")
	return r
}

func (a *ActionSuite) byteSlice() []byte {
	return []byte("1234")
}

func (a *ActionSuite) TestList() {
	r := a.Require()
	store := a.store

	const moduleName = "modtest"
	versions := []testCase{
		{version: "v1.0.0", mod: a.byteSlice(), zip: a.getReader(), info: a.byteSlice()},
	}

	req := a.JSON("/github.com/gomods/athens/@v/list")
	/*
		v1.0.0
		v1.0.1
		v0.0.2
	*/
	res := req.Get()
	r.Equal(200, res.Code)
}
