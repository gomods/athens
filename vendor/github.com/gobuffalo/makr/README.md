# makr

[![GoDoc](https://godoc.org/github.com/gobuffalo/makr?status.svg)](https://godoc.org/github.com/gobuffalo/makr)

Makr is a file generation system for Go.

## Usage

### Execute a command
```go
// Execute a npm install command.
g := makr.New()
g.Add(makr.NewCommand(exec.Command("npm", "install")))
err = g.Run(".", makr.Data{})
if err != nil {
    // Error!
}
```

### Execute Go commands
```go
// Get golang.org/x/tools/cmd/goimports package, update it if it's already in GOPATH.
g := makr.New()
g.Add(makr.NewCommand(makr.GoGet("golang.org/x/tools/cmd/goimports", "-u")))
err = g.Run(".", makr.Data{})
if err != nil {
    // Error!
}
```

**Available Go commands**:
* GoGet
* GoInstall
* GoFmt

### Create a file from a golang template string
```go
s := "my file contents"
g := makr.New()
g.Add(makr.NewFile("file.txt", s))
err = g.Run(".", makr.Data{})
if err != nil {
    // Error!
}
```

## Chained usage
```go
s = `
{
  "name": "buffalo",
  "version": "1.0.0",
  "main": "index.js",
  "license": "MIT",
  "dependencies": {
    "babel-cli": "~6.24.1",
    "babel-core": "~6.25.0",
    "babel-loader": "~7.0.0",
    "babel-preset-env": "~1.5.2",
    "bootstrap-sass": "~3.3.7",
    "clean-webpack-plugin": "~0.1.17",
    "copy-webpack-plugin": "~4.0.1",
    "css-loader": "~0.28.4",
    "expose-loader": "~0.7.3",
    "extract-text-webpack-plugin": "2.1.2",
    "file-loader": "~0.11.2",
    "font-awesome": "~4.7.0",
    "gopherjs-loader": "^0.0.1",
    "highlightjs": "^9.10.0",
    "jquery": "~3.2.1",
    "jquery-ujs": "~1.2.2",
    "node-sass": "~4.7.2",
    "npm-install-webpack-plugin": "4.0.4",
    "path": "~0.12.7",
    "sass-loader": "~6.0.5",
    "style-loader": "~0.18.2",
    "uglifyjs-webpack-plugin": "~0.4.6",
    "url-loader": "~0.5.9",
    "webpack": "~2.3.0",
    "webpack-manifest-plugin": "~1.2.1"
  }
}
`
// Create a package.json file, then execute npm install
g := makr.New()
g.Add(makr.NewFile("package.json", s))
g.Add(makr.NewCommand(exec.Command("npm", "install")))

err = g.Run(".", makr.Data{})
if err != nil {
    // Error!
}
```
