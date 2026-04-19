---
title: 主页模板配置
description: 如何自定义主页模板
weight: 8
---

从 v0.14.0 起，Athens 自带一个默认的最小化 HTML 首页，指导用户如何连接代理。它会根据是否配置了 GoNoSumPatterns，尝试为 GO_PROXY 构建配置。它利用请求 Host 头（HTTP 1.1）或 Authority 头（HTTP 2）以及是否基于 TLS 来指导配置 `GO_PROXY` 。最后，还提供了关于如何利用 Athens API 的快速指南。

当然，并非所有场景都如此简单。某些安装环境在 CI 中可达的地址可能与桌面用户不同。在这种情况下，以及其他默认首页没有意义的情况下，可以覆盖模板。

通过配置 `HomeTemplatePath`（通过配置或 `ATHENS_HOME_TEMPLATE_PATH` 环境变量）指向磁盘上包含 Go HTML 模板的位置，或将模板文件放置在 `/var/lib/athens/home.html`。

Athens 会自动向模板注入以下变量：

| 设置 | 来源 |
| :------ | :----- |
| `Host` | 根据请求 Host（HTTP1）或 Authority（HTTP2）头以及 TLS 的存在构建，包含端口。 |
| `NoSumPatterns` | 直接来自配置。 |

使用这些值的方法是将它们包裹在带前置点的括号中。示例：`{{ .Host }}`

有关更高级的格式化，请阅读 [Go HTML 模板](https://pkg.go.dev/html/template) 相关文档。

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
	<p>Use the following GONOSUMDB environment variable to exclude checksum database:</p>
	<pre>GONOSUMDB={{ .NoSumPatterns }}</pre>
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
