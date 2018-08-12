package actions

import (
	"testing"

	"github.com/gobuffalo/suite"
)

type ActionSuite struct {
	*suite.Action
}

func Test_ActionSuite(t *testing.T) {
	app, err := App()
	if err != nil {
		t.Fatal(err)
	}
	as := &ActionSuite{suite.NewAction(app)}
	suite.Run(t, as)
}

/*
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
}*/
