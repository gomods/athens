package actions

import (
	"errors"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/log"
)

const homepage = `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8"></meta>
	<title>Athens</title>
	<style>
		body {
			font-family: Arial, sans-serif;
			margin: 20px;
		}

		pre {
				background-color: #f4f4f4;
				padding: 5px;
				border-radius: 5px;
				width: fit-content;
  				padding: 10px;
		}


		code {
			background-color: #f4f4f4;
			padding: 5px;
			border-radius: 5px;
		}

	</style>
</head>
<body>
	
	<h1>Welcome to Athens</h1>

	<h2>Configuring your client</h2>
	<pre>GOPROXY={{ .Host }},direct</pre>
	{{ if .NoSumPatterns }}
	<h3>Excluding checksum database</h3>
	<p>Use the following GONOSUM environment variable to exclude checksum database:</p>
	<pre>GONOSUM={{ .NoSumPatterns }}</pre>
	{{ end }}

	<h2>How to use the Athens API</h2>
	<p>Use the <a href="/catalog">catalog</a> endpoint to get a list of all modules in the proxy</p>

	<h3>List of versions</h3>
	<p>This endpoint returns a list of versions that Athens knows about for <code>acidburn/htp</code>:</p>
	<pre>GET {{ .Host }}/github.com/acidburn/htp/@v/list</pre>

	<h3>Version info</h3>
	<p>This endpoint returns information about a specific version of a module:</p>
	<pre>GET {{ .Host }}/github.com/acidburn/htp/@v/v1.0.0.info</pre>
	<p>This returns JSON with information about v1.0.0. It looks like this:
	<pre>{
	"Name": "v1.0.0",
	"Short": "v1.0.0",
	"Version": "v1.0.0",
	"Time": "1972-07-18T12:34:56Z"
}</pre>

	<h3>go.mod file</h3>
	<p>This endpoint returns the go.mod file for a specific version of a module:</p>
	<pre>GET {{ .Host }}/github.com/acidburn/htp/@v/v1.0.0.mod</pre>
	<p>This returns the go.mod file for version v1.0.0. If {{ .Host }}/github.com/acidburn/htp version v1.0.0 has no dependencies, the response body would look like this:</p>
	<pre>module github.com/acidburn/htp</pre>

	<h3>Module sources</h3>
	<pre>GET {{ .Host }}/github.com/acidburn/htp/@v/v1.0.0.zip</pre>
	<p>This is what it sounds like — it sends back a zip file with the source code for the module in version v1.0.0.</p>

	<h3>Latest</h3>
	<pre>GET {{ .Host }}/github.com/acidburn/htp/@latest</pre>
	<p>This endpoint returns the latest version of the module. If the version does not exist it should retrieve the hash of latest commit.</p>

</body>
</html>
`

func proxyHomeHandler(c *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lggr := log.EntryFromContext(r.Context())

		templateData := make(map[string]string)

		templateContents := homepage

		// load the template from the file system if it exists, otherwise revert to default
		rawTemplateFileContents, err := os.ReadFile(c.HomeTemplatePath)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				// this is some other error, log it and revert to default
				lggr.SystemErr(err)
			}
		} else {
			templateContents = string(rawTemplateFileContents)
		}

		// This should be correct in most cases. If it is not, users can supply their own template
		templateData["Host"] = r.Host

		// if the host does not have a scheme, add one based on the request
		if !strings.HasPrefix(templateData["Host"], "http://") && !strings.HasPrefix(templateData["Host"], "https://") {
			if r.TLS != nil {
				templateData["Host"] = "https://" + templateData["Host"]
			} else {
				templateData["Host"] = "http://" + templateData["Host"]
			}
		}

		templateData["NoSumPatterns"] = strings.Join(c.NoSumPatterns, ",")

		tmp, err := template.New("home").Parse(templateContents)
		if err != nil {
			lggr.SystemErr(err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		err = tmp.ExecuteTemplate(w, "home", templateData)
		if err != nil {
			lggr.SystemErr(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
