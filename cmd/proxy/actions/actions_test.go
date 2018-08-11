package actions

import (
	"testing"

	"github.com/gobuffalo/suite"
	"github.com/gomods/athens/pkg/module"
)

type ActionSuite struct {
	*suite.Action
}

func newTestFilter() *module.Filter {
	f := module.NewFilter()
	f.AddRule("github.com/gomods/athens/", module.Include)
	f.AddRule("github.com/athens-artifacts/no-tags", module.Exclude)
	f.AddRule("github.com/athens-artifacts", module.Private)
	return f
}

func Test_ActionSuite(t *testing.T) {
	f := newTestFilter()
	app, err := App(f)
	if err != nil {
		t.Fatal(err)
	}
	as := &ActionSuite{suite.NewAction(app)}
	suite.Run(t, as)
}

func (a *ActionSuite) Test_Filter() {
	r := a.Require()

	// Public, expects to be redirected to olympus
	req := a.JSON("/github.com/gomods/athens/@v/list")
	res := req.Get()
	r.Equal(303, res.Code)
	r.Equal(GetOlympusEndpoint()+"/github.com/gomods/athens/@v/list", res.HeaderMap.Get("Location"))

	// Excluded, expects a 403
	req = a.JSON("/github.com/athens-artifacts/no-tags/@v/list")
	res = req.Get()
	r.Equal(403, res.Code)

	// Private, the proxy is working and returns a 200
	req = a.JSON("/github.com/athens-artifacts/happy-path/@v/list")
	res = req.Get()
	r.Equal(200, res.Code)
}
