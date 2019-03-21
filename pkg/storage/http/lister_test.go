package http

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

const (
	// pythonIndexPage is an example module index page served from "python -m SimpleServer".
	pythonIndexPage = `<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 3.2 Final//EN"><html>
<title>Directory listing for /golang.org/x/net/@v/</title>
<body>
<h2>Directory listing for /golang.org/x/net/@v/</h2>
<hr>
<ul>
<li><a href="list">list</a>
<li><a href="v0.0.0-20180724234803-3673e40ba225.info">v0.0.0-20180724234803-3673e40ba225.info</a>
<li><a href="v0.0.0-20180724234803-3673e40ba225.mod">v0.0.0-20180724234803-3673e40ba225.mod</a>
<li><a href="v0.0.0-20180906233101-161cd47e91fd.info">v0.0.0-20180906233101-161cd47e91fd.info</a>
<li><a href="v0.0.0-20180906233101-161cd47e91fd.mod">v0.0.0-20180906233101-161cd47e91fd.mod</a>
<li><a href="v0.0.0-20181029044818-c44066c5c816.info">v0.0.0-20181029044818-c44066c5c816.info</a>
<li><a href="v0.0.0-20181029044818-c44066c5c816.mod">v0.0.0-20181029044818-c44066c5c816.mod</a>
</ul>
<hr>
</body>
</html>
`

	// artifactoryIndexPage is an example module index page served from Artifactory.
	artifactoryIndexPage = `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 3.2 Final//EN">
<html>
<head><title>Index of athens-modules/golang.org/x/net/@v</title>
</head>
<body>
<h1>Index of athens-modules/golang.org/x/net/@v</h1>
<pre>Name                                     Last modified      Size</pre><hr/>
<pre><a href="../">../</a>
<a href="v0.0.0-20180724234803-3673e40ba225.info">v0.0.0-20180724234803-3673e40ba225.info</a>   17-Jan-2019 19:40  78 bytes
<a href="v0.0.0-20180724234803-3673e40ba225.mod">v0.0.0-20180724234803-3673e40ba225.mod</a>    17-Jan-2019 19:40  24 bytes
<a href="v0.0.0-20180724234803-3673e40ba225.zip">v0.0.0-20180724234803-3673e40ba225.zip</a>    17-Jan-2019 19:40  1.26 MB
<a href="v0.0.0-20180906233101-161cd47e91fd.info">v0.0.0-20180906233101-161cd47e91fd.info</a>   17-Jan-2019 19:41  78 bytes
<a href="v0.0.0-20180906233101-161cd47e91fd.mod">v0.0.0-20180906233101-161cd47e91fd.mod</a>    17-Jan-2019 19:41  24 bytes
<a href="v0.0.0-20180906233101-161cd47e91fd.zip">v0.0.0-20180906233101-161cd47e91fd.zip</a>    17-Jan-2019 19:41  1.27 MB
<a href="v0.0.0-20181029044818-c44066c5c816.info">v0.0.0-20181029044818-c44066c5c816.info</a>   17-Jan-2019 19:41  78 bytes
<a href="v0.0.0-20181029044818-c44066c5c816.mod">v0.0.0-20181029044818-c44066c5c816.mod</a>    17-Jan-2019 19:41  24 bytes
<a href="v0.0.0-20181029044818-c44066c5c816.zip">v0.0.0-20181029044818-c44066c5c816.zip</a>    17-Jan-2019 19:41  1.27 MB
</pre>
<hr/><address style="font-size:small;">Artifactory/6.6.0 Server at artifactory.server.io Port 8081</address></body></html>
`

	// externalLinkIndexPage is a contrived module index page containing a link to a ".mod" file on another domain.
	externalLinkIndexPage = `<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 3.2 Final//EN"><html>
<title>Directory listing for /golang.org/x/net/@v/</title>
<body>
<h2>Directory listing for /golang.org/x/net/@v/</h2>
<hr>
<ul>
<li><a href="list">list</a>
<li><a href="v0.0.0-20180724234803-3673e40ba225.mod">v0.0.0-20180724234803-3673e40ba225.mod</a>
<li><a href="https://raw.githubusercontent.com/gomods/athens/master/go.mod">go.mod</a>
</ul>
<hr>
</body>
</html>
`
)

func TestCollectLinks(t *testing.T) {

	cases := []struct {
		Reader   io.Reader
		Filter   func(string) bool
		Expected []string
	}{
		{
			Reader: strings.NewReader(artifactoryIndexPage),
			Expected: []string{
				"../",
				"v0.0.0-20180724234803-3673e40ba225.info",
				"v0.0.0-20180724234803-3673e40ba225.mod",
				"v0.0.0-20180724234803-3673e40ba225.zip",
				"v0.0.0-20180906233101-161cd47e91fd.info",
				"v0.0.0-20180906233101-161cd47e91fd.mod",
				"v0.0.0-20180906233101-161cd47e91fd.zip",
				"v0.0.0-20181029044818-c44066c5c816.info",
				"v0.0.0-20181029044818-c44066c5c816.mod",
				"v0.0.0-20181029044818-c44066c5c816.zip",
			},
		},
		{
			Reader: strings.NewReader(pythonIndexPage),
			Expected: []string{
				"list",
				"v0.0.0-20180724234803-3673e40ba225.info",
				"v0.0.0-20180724234803-3673e40ba225.mod",
				"v0.0.0-20180906233101-161cd47e91fd.info",
				"v0.0.0-20180906233101-161cd47e91fd.mod",
				"v0.0.0-20181029044818-c44066c5c816.info",
				"v0.0.0-20181029044818-c44066c5c816.mod",
			},
		},
		{
			Reader: strings.NewReader(artifactoryIndexPage),
			Filter: func(s string) bool {
				return strings.HasSuffix(s, ".mod")
			},
			Expected: []string{
				"v0.0.0-20180724234803-3673e40ba225.mod",
				"v0.0.0-20180906233101-161cd47e91fd.mod",
				"v0.0.0-20181029044818-c44066c5c816.mod",
			},
		},
		{
			Reader: strings.NewReader(externalLinkIndexPage),
			Filter: func(s string) bool {
				baseURL := "http://artifactory.server.io/athens-modules"
				return strings.HasPrefix(absolute(baseURL, s), baseURL) && strings.HasSuffix(s, ".mod")
			},
			Expected: []string{
				"v0.0.0-20180724234803-3673e40ba225.mod",
			},
		},
	}

	for _, c := range cases {
		if actual, _ := collectLinks(c.Reader, c.Filter); !reflect.DeepEqual(actual, c.Expected) {
			t.Errorf("unexpected link list: expected %q but got %q", c.Expected, actual)
		}
	}

}
