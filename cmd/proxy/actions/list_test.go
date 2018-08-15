package actions

func (a *ActionSuite) TestList() {
	r := a.Require()

	req := a.JSON("/github.com/gomods/athens/@v/list")
	res := req.Get()
	r.Equal(200, res.Code)
}
