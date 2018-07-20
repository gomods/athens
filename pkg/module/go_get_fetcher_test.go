package module

func (s *ModuleSuite) TestNewGoGetFetcher() {
	const (
		repoURI = "github.com/arschles/assert"
		version = "v1.0.0"
	)
	r := s.Require()
	fs := s.fs
	fetcher, err := NewGoGetFetcher(fs, repoURI, version)
	r.NoError(err)
	goGetFetcher, ok := fetcher.(*goGetFetcher)
	r.True(ok)
	r.Equal(repoURI, goGetFetcher.repoURI)
	r.Equal(version, goGetFetcher.version)
}
