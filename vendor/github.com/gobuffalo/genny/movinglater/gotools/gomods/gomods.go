package gomods

import (
	"strings"

	"github.com/gobuffalo/envy"
	"github.com/pkg/errors"
)

const ENV = "GO111MODULE"

var ErrModsOff = errors.New("go mods are turned off")
var modsOn = (strings.TrimSpace(envy.Get(ENV, "off")) == "on")

func On() bool {
	return modsOn
}

func Disable(fn func() error) error {
	gm := envy.Get("GO111MODULE", "off")
	defer envy.MustSet("GO111MODULE", gm)
	if err := envy.MustSet("GO111MODULE", "off"); err != nil {
		return errors.WithStack(err)
	}

	// this ensures the defer gets called after fn()
	// doing return fn() would have it called before
	if err := fn(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
