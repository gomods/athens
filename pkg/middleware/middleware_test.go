package middleware

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	ht "github.com/gobuffalo/httptest"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/module"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Avoid import cycle.
const (
	pathList        = "/{module:.+}/@v/list"
	pathVersionInfo = "/{module:.+}/@v/{version}.info"
)

func testConfigFile(t *testing.T) (testConfigFile string) {
	testConfigFile = filepath.Join("..", "..", "config.dev.toml")
	if err := os.Chmod(testConfigFile, 0o700); err != nil {
		t.Fatalf("%s\n", err)
	}
	return testConfigFile
}

func middlewareFilterApp(filterFile, registryEndpoint string) (*mux.Router, error) {
	h := func(w http.ResponseWriter, r *http.Request) {}
	r := mux.NewRouter()
	mf, err := newTestFilter(filterFile)
	if err != nil {
		return nil, err
	}
	r.Use(NewFilterMiddleware(mf, registryEndpoint))

	r.HandleFunc(pathList, h)
	r.HandleFunc(pathVersionInfo, h)
	return r, nil
}

func newTestFilter(filterFile string) (*module.Filter, error) {
	f, err := module.NewFilter(filterFile)
	if err != nil {
		return nil, err
	}
	f.AddRule("github.com/gomods/athens/", nil, module.Direct)
	f.AddRule("github.com/athens-artifacts/no-tags", nil, module.Exclude)
	f.AddRule("github.com/athens-artifacts", nil, module.Include)
	return f, nil
}

func Test_FilterMiddleware(t *testing.T) {
	r := require.New(t)

	filter, err := os.CreateTemp(os.TempDir(), "filter-")
	if err != nil {
		t.FailNow()
	}
	defer os.Remove(filter.Name())

	conf, err := config.GetConf(testConfigFile(t))
	if err != nil {
		t.Fatalf("Unable to parse config file: %s", err.Error())
	}

	// Test with a filter file not existing
	app, err := middlewareFilterApp("nofsfile", conf.GlobalEndpoint)
	r.Nil(app, "app should be nil when a file not exisiting")
	r.Error(err, "Expected error when a file not existing on the filesystem is given")

	app, err = middlewareFilterApp(filter.Name(), conf.GlobalEndpoint)
	r.NoError(err, "app should be successfully created in the test")
	w := ht.New(app)

	path := "/github.com/gomods/athens/@v/list"
	res := w.JSON(path).Get()
	r.Equal(http.StatusSeeOther, res.Code)
	r.Equal(conf.GlobalEndpoint+"/github.com/gomods/athens/@v/list", res.HeaderMap.Get("Location"))

	// Excluded, expects a 403
	res = w.JSON("/github.com/athens-artifacts/no-tags/@v/list").Get()
	r.Equal(http.StatusForbidden, res.Code)

	// Private, the proxy is working and returns a 200
	res = w.JSON("/github.com/athens-artifacts/happy-path/@v/list").Get()
	r.Equal(http.StatusOK, res.Code)
}

func hookFilterApp(hook string) *mux.Router {
	h := func(w http.ResponseWriter, r *http.Request) {}
	r := mux.NewRouter()
	r.Use(NewValidationMiddleware(http.DefaultClient, hook))

	r.HandleFunc(pathList, h)
	r.HandleFunc(pathVersionInfo, h)
	return r
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
	server *ht.Server
	w      *ht.Handler
}

func (suite *HookTestsSuite) SetupSuite() {
	suite.server = ht.NewServer(&suite.mock)
	suite.w = ht.New(hookFilterApp(suite.server.URL))
}

func (suite *HookTestsSuite) SetupTest() {
	suite.mock.invoked = false
	suite.mock.resCode = 0
}

func (suite *HookTestsSuite) TearDownSuite() {
	suite.server.Close()
}

func TestHookTestSuite(t *testing.T) {
	suite.Run(t, new(HookTestsSuite))
}

func (suite *HookTestsSuite) TestHookOnList() {
	r := suite.Require()
	// list path, hook should not be hit
	suite.w.JSON("/github.com/gomods/athens/@v/list").Get()
	r.False(suite.mock.invoked)
}

func (suite *HookTestsSuite) TestHookPass() {
	r := suite.Require()
	// hit and pass
	suite.mock.resCode = http.StatusOK
	res := suite.w.JSON("/github.com/athens-artifacts/happy-path/@v/v1.0.0.info").Get()
	r.True(suite.mock.invoked)
	r.Equal(http.StatusOK, res.Code)
	r.Equal("github.com/athens-artifacts/happy-path", suite.mock.params.Module)
	r.Equal("v1.0.0", suite.mock.params.Version)
}

func (suite *HookTestsSuite) TestHookBlocks() {
	r := suite.Require()

	// hit but hook blocks
	suite.mock.resCode = http.StatusForbidden
	res := suite.w.JSON("/github.com/athens-artifacts/happy-path/@v/v1.0.0.info").Get()
	r.True(suite.mock.invoked)
	r.Equal(http.StatusForbidden, res.Code)
}

func (suite *HookTestsSuite) TestHookUnexpectedError() {
	r := suite.Require()

	// hit but unexpected error
	suite.mock.resCode = http.StatusGone
	res := suite.w.JSON("/github.com/athens-artifacts/happy-path/@v/v1.0.0.info").Get()
	r.True(suite.mock.invoked)
	r.Equal(http.StatusInternalServerError, res.Code)
}
