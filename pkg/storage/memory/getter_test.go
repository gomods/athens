package memory

func (m *MemoryTests) TestGet() {
	mem := m.mem
	r := m.Require()
	const version = "v1.0.0"
	var (
		gomod = []byte(`module "my/module"`)
		zip   = []byte("asdasdfasdfa")
	)
	r.NoError(mem.Save(baseURL, module, version, gomod, zip))
	vsn, err := mem.Get(baseURL, module, version)
	r.NoError(err)
	r.Equal(version, vsn.RevInfo.Version)
	r.Equal(gomod, vsn.Mod)
	r.Equal(zip, vsn.Zip)
}
