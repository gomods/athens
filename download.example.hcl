downloadURL = "https://proxy.golang.org"

mode = "sync"

download "github.com/gomods/*" {
    mode = "async_redirect"
}

download "golang.org/x/*" {
    mode = "none"
}

download "github.com/pkg/*" {
    mode = "redirect"
    downloadURL = "https://gocenter.io"
}
