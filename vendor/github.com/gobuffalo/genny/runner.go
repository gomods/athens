package genny

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/markbates/oncer"
	"github.com/markbates/safe"
	"github.com/pkg/errors"
)

type RunFn func(r *Runner) error

// Runner will run the generators
type Runner struct {
	Logger     Logger                                                    // Logger to use for the run
	Context    context.Context                                           // context to use for the run
	ExecFn     func(*exec.Cmd) error                                     // function to use when executing files
	FileFn     func(File) (File, error)                                  // function to use when writing files
	ChdirFn    func(string, func() error) error                          // function to use when changing directories
	DeleteFn   func(string) error                                        // function used to delete files/folders
	RequestFn  func(*http.Request, *http.Client) (*http.Response, error) // function used to make http requests
	LookPathFn func(string) (string, error)                              // function used to make exec.LookPath lookups
	Root       string                                                    // the root of the write path
	Disk       *Disk
	generators []*Generator
	moot       *sync.RWMutex
	results    Results
	curGen     *Generator
}

func (r *Runner) Results() Results {
	r.moot.Lock()
	defer r.moot.Unlock()
	r.results.Files = r.Disk.Files()
	return r.results
}

func (r *Runner) WithRun(fn RunFn) {
	g := New()
	g.RunFn(fn)
	r.With(g)
}

// With adds a Generator to the Runner
func (r *Runner) With(g *Generator) {
	r.moot.Lock()
	defer r.moot.Unlock()
	r.generators = append(r.generators, g)
}

func (r *Runner) WithGroup(gg *Group) {
	for _, g := range gg.Generators {
		r.With(g)
	}
}

// WithNew takes a Generator and an error.
// Perfect for new-ing up generators
/*
	// foo.New(Options) (*genny.Generator, error)
	if err := run.WithNew(foo.New(opts)); err != nil {
		return err
	}
*/
func (r *Runner) WithNew(g *Generator, err error) error {
	if err != nil {
		return errors.WithStack(err)
	}
	r.With(g)
	return nil
}

// WithFn will evaluate the function and if successful it will add
// the Generator to the Runner, otherwise it will return the error
// Deprecated
func (r *Runner) WithFn(fn func() (*Generator, error)) error {
	oncer.Deprecate(5, "genny.Runner#WithFn", "")
	return safe.RunE(func() error {
		g, err := fn()
		if err != nil {
			return errors.WithStack(err)
		}
		r.With(g)
		return nil
	})
}

// Run all of the generators!
func (r *Runner) Run() error {
	r.moot.Lock()
	defer r.moot.Unlock()
	for _, g := range r.generators {
		r.curGen = g
		if g.Should != nil {
			err := safe.RunE(func() error {
				if !g.Should(r) {
					return io.EOF
				}
				return nil
			})
			if err != nil {
				continue
			}
		}
		err := r.Chdir(r.Root, func() error {
			for _, fn := range g.runners {
				err := safe.RunE(func() error {
					return fn(r)
				})
				if err != nil {
					return errors.WithStack(err)
				}
			}
			return nil
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// Exec can be used inside of Generators to run commands
func (r *Runner) Exec(cmd *exec.Cmd) error {
	r.results.Commands = append(r.results.Commands, cmd)
	r.Logger.Debug("Exec: ", strings.Join(cmd.Args, " "))
	if r.ExecFn == nil {
		return nil
	}
	return safe.RunE(func() error {
		return r.ExecFn(cmd)
	})
}

func (r *Runner) LookPath(s string) (string, error) {
	r.Logger.Debug("LookPath: ", s)
	if r.LookPathFn != nil {
		return r.LookPathFn(s)
	}
	return s, nil
}

// File can be used inside of Generators to write files
func (r *Runner) File(f File) error {
	if r.curGen != nil {
		var err error
		f, err = r.curGen.Transform(f)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	name := f.Name()
	if !filepath.IsAbs(name) {
		name = filepath.Join(r.Root, name)
	}
	r.Logger.Debug("File: ", name)
	if r.FileFn != nil {
		err := safe.RunE(func() error {
			var e error
			if f, e = r.FileFn(f); e != nil {
				return errors.WithStack(e)
			}
			if s, ok := f.(io.Seeker); ok {
				s.Seek(0, 0)
			}
			return nil
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}
	f = NewFile(f.Name(), f)
	if s, ok := f.(io.Seeker); ok {
		s.Seek(0, 0)
	}
	r.Disk.Add(f)
	return nil
}

func (r *Runner) FindFile(name string) (File, error) {
	return r.Disk.Find(name)
}

// Chdir will change to the specified directory
// and revert back to the current directory when
// the runner function has returned.
// If the directory does not exist, it will be
// created for you.
func (r *Runner) Chdir(path string, fn func() error) error {
	if len(path) == 0 {
		return fn()
	}
	r.Logger.Debug("Chdir: ", path)

	if r.ChdirFn != nil {
		return safe.RunE(func() error {
			return r.ChdirFn(path, fn)
		})
	}

	if err := safe.RunE(fn); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (r *Runner) Delete(path string) error {
	r.Logger.Debug("Delete: ", path)

	defer r.Disk.Remove(path)
	if r.DeleteFn != nil {
		return safe.RunE(func() error {
			return r.DeleteFn(path)
		})
	}
	return nil
}

func (r *Runner) Request(req *http.Request) (*http.Response, error) {
	return r.RequestWithClient(req, http.DefaultClient)
}

func (r *Runner) RequestWithClient(req *http.Request, c *http.Client) (*http.Response, error) {
	key := fmt.Sprintf("[%s] %s\n", strings.ToUpper(req.Method), req.URL)
	r.Logger.Debug("Request: ", key)
	store := func(res *http.Response, err error) (*http.Response, error) {
		r.moot.Lock()
		r.results.Requests = append(r.results.Requests, RequestResult{
			Request:  req,
			Response: res,
			Client:   c,
			Error:    err,
		})
		r.moot.Unlock()
		return res, err
	}
	if r.RequestFn == nil {
		return store(nil, nil)
	}
	var res *http.Response
	err := safe.RunE(func() error {
		var e error
		res, e = r.RequestFn(req, c)
		if e != nil {
			return errors.WithStack(e)
		}
		return nil
	})
	return store(res, err)
}
