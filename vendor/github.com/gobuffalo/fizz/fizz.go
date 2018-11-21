/*
Package fizz is a common DSL for writing SQL migrations
*/
package fizz

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
)

type Options map[string]interface{}

type fizzer struct {
	Bubbler *Bubbler
}

func (f fizzer) add(s string, err error) error {
	if err != nil {
		return errors.WithStack(err)
	}
	f.Bubbler.data = append(f.Bubbler.data, s)
	return nil
}

func (f fizzer) Exec(out io.Writer) func(string) error {
	return func(s string) error {
		args, err := shellquote.Split(s)
		if err != nil {
			return errors.Wrapf(err, "error parsing command: %s", s)
		}
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = out
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return errors.Wrapf(err, "error executing command: %s", s)
		}
		return nil
	}
}

// AFile reads a fizz file, and translates its contents to SQL.
func AFile(f *os.File, t Translator) (string, error) {
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return AString(string(b), t)
}

// AString reads a fizz string, and translates its contents to SQL.
func AString(s string, t Translator) (string, error) {
	b := NewBubbler(t)
	return b.Bubble(s)
}
