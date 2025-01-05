---
title: Home template configuration
description: How to customize the home template
weight: 8
---

As of v0.14.0 Athens ships with a default, minimal HTML home page that advises users on how to connect to the proxy. It factors in whether `GoNoSumPatterns` is configured, and attempts
to build configuration for `GO_PROXY`. It relies on the users request Host header (on HTTP 1.1) or the Authority header (on HTTP 2) as well as whether the request was over TLS to advise 
on configuring `GO_PROXY`. Lastly, the homepage provides a quick guide on how users can leverage the Athens API.

Of course, not all instructions will be this simple. Some installations may be reachable at different addresses in CI than for desktop users. In this case, and others where the default
home page does not make sense it is possible to override the template.

Do so by configuring `HomeTemplatePath` via the config or `ATHENS_HOME_TEMPLATE_PATH` to a location on disk with a Go HTML template or placing a template file at `/var/lib/athens/home.html`.

Athens automatically injects the following variables in templates:

| Setting         | Source |
| :------         | :----- |
| `Host`          | Built from the request Host (HTTP1) or Authority (HTTP2) header and presence of TLS. Includes ports. |
| `NoSumPatterns` | Comes directly from the configuration. |

Using these values is done by wrapping them in bracers with a dot prepended. Example: `{{ .Host }}`

For more advanced formatting read more about [Go HTML templates](https://pkg.go.dev/html/template).

```html
<!DOCTYPE html>
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
```