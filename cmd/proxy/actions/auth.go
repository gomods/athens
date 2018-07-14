package actions

import (
	"fmt"
	"log"
	"os"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
)

func init() {
	app, err := App()
	if err != nil {
		log.Fatal(err)
	}
	gothic.Store = app.SessionStore

	goth.UseProviders(
		github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"), fmt.Sprintf("%s%s", app.Host, "/auth/github/callback")),
	)
}
