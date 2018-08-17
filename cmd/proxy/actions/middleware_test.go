package actions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gobuffalo/envy"

	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/download"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/module"
	"github.com/markbates/willie"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func middlewareFilterApp() *buffalo.App {
	h := func(c buffalo.Context) error {
		return c.Render(200, nil)
	}

	a := buffalo.New(buffalo.Options{})
	mf := newTestFilter()
	a.Use(newFilterMiddleware(mf))
	initializeTracing(a)

	a.GET(download.PathList, h)
	a.GET(download.PathVersionInfo, h)
	return a
}

func newTestFilter() *module.Filter {
	f := module.NewFilter()
	f.AddRule("github.com/gomods/athens/", module.Include)
	f.AddRule("github.com/athens-artifacts/no-tags", module.Exclude)
	f.AddRule("github.com/athens-artifacts", module.Direct)
	return f
}

func Test_FilterMiddleware(t *testing.T) {
	r := require.New(t)

	w := willie.New(middlewareFilterApp())

	// Public, expects to be redirected to olympus
	res := w.Request("/github.com/gomods/athens/@v/list").Get()
	r.Equal(303, res.Code)
	r.Equal(GetOlympusEndpoint()+"/github.com/gomods/athens/@v/list", res.HeaderMap.Get("Location"))

	// Excluded, expects a 403
	res = w.Request("/github.com/athens-artifacts/no-tags/@v/list").Get()
	r.Equal(403, res.Code)

	// Private, the proxy is working and returns a 200
	res = w.Request("/github.com/athens-artifacts/happy-path/@v/list").Get()
	r.Equal(200, res.Code)
}

func hookFilterApp() *buffalo.App {
	h := func(c buffalo.Context) error {
		return c.Render(200, nil)
	}

	a := buffalo.New(buffalo.Options{})
	a.Use(LogEntryMiddleware(newValidationMiddleware, log.New("none", "debug")))
	initializeTracing(a)

	a.GET(download.PathList, h)
	a.GET(download.PathVersionInfo, h)
	return a
}

type hookMock struct {
	invoked bool
	params  validationParams
	resCode int
}

func (m *hookMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.invoked = true
	w.WriteHeader(m.resCode)
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&m.params)
}

type HookTestsSuite struct {
	suite.Suite
	mock   hookMock
	server *httptest.Server
	w      *willie.Willie
}

func (suite *HookTestsSuite) SetupSuite() {
	fmt.Println("setup")
	suite.server = httptest.NewServer(&suite.mock)
	envy.Set("ATHENS_PROXY_VALIDATOR", suite.server.URL)
	suite.w = willie.New(hookFilterApp())
}

func (suite *HookTestsSuite) SetupTest() {
	suite.mock.invoked = false
	suite.mock.resCode = 0
}

func (suite *HookTestsSuite) TearDownSuite() {
	envy.Set("ATHENS_PROXY_VALIDATOR", "")
	suite.server.Close()
}

func TestHookTestSuite(t *testing.T) {
	suite.Run(t, new(HookTestsSuite))
}

func (suite *HookTestsSuite) TestHookOnList() {
	r := suite.Require()
	// list path, hook should not be hit
	suite.w.Request("/github.com/gomods/athens/@v/list").Get()
	r.False(suite.mock.invoked)
}

func (suite *HookTestsSuite) TestHookPass() {
	r := suite.Require()
	// hit and pass
	suite.mock.resCode = http.StatusOK
	res := suite.w.Request("/github.com/athens-artifacts/happy-path/@v/v1.0.0.info").Get()
	r.True(suite.mock.invoked)
	r.Equal(http.StatusOK, res.Code)
	r.Equal("github.com/athens-artifacts/happy-path", suite.mock.params.Module)
	r.Equal("v1.0.0", suite.mock.params.Version)
}

func (suite *HookTestsSuite) TestHookBlocks() {
	r := suite.Require()

	// hit but hook blocks
	suite.mock.resCode = http.StatusForbidden
	res := suite.w.Request("/github.com/athens-artifacts/happy-path/@v/v1.0.0.info").Get()
	r.True(suite.mock.invoked)
	r.Equal(http.StatusForbidden, res.Code)
}

func (suite *HookTestsSuite) TestHookUnexpectedError() {
	r := suite.Require()

	// hit but unexpected error
	suite.mock.resCode = http.StatusGone
	res := suite.w.Request("/github.com/athens-artifacts/happy-path/@v/v1.0.0.info").Get()
	r.True(suite.mock.invoked)
	r.Equal(http.StatusInternalServerError, res.Code)
}
