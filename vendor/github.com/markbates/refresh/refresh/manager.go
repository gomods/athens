package refresh

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/events"
	"github.com/gobuffalo/genny/movinglater/gotools/gomods"
)

type Manager struct {
	*Configuration
	Logger     *Logger
	Restart    chan bool
	gil        *sync.Once
	ID         string
	context    context.Context
	cancelFunc context.CancelFunc
}

func New(c *Configuration) *Manager {
	return NewWithContext(c, context.Background())
}

func NewWithContext(c *Configuration, ctx context.Context) *Manager {
	ctx, cancelFunc := context.WithCancel(ctx)
	m := &Manager{
		Configuration: c,
		Logger:        NewLogger(c),
		Restart:       make(chan bool),
		gil:           &sync.Once{},
		ID:            ID(),
		context:       ctx,
		cancelFunc:    cancelFunc,
	}
	return m
}

func (r *Manager) Start() error {
	w := NewWatcher(r)
	w.Start()
	go r.build(fsnotify.Event{Name: ":start:"})
	if !r.Debug {
		go func() {
			for {
				select {
				case event := <-w.Events:
					events.EmitPayload(EvtRaw, events.Payload{"event": event})
					if event.Op != fsnotify.Chmod {
						go r.build(event)
					}
					w.Remove(event.Name)
					w.Add(event.Name)
				case <-r.context.Done():
					break
				}
			}
		}()
	}
	go func() {
		for {
			select {
			case err := <-w.Errors:
				r.Logger.Error(err)
			case <-r.context.Done():
				break
			}
		}
	}()
	r.runner()
	return nil
}

func (r *Manager) build(event fsnotify.Event) {
	r.gil.Do(func() {
		defer func() {
			r.gil = &sync.Once{}
		}()
		r.buildTransaction(func() error {
			// time.Sleep(r.BuildDelay * time.Millisecond)

			payload := events.Payload{
				"event": event,
			}

			now := time.Now()
			r.Logger.Print("Rebuild on: %s", event.Name)

			args := []string{"build", "-v"}
			if !gomods.On() {
				args = append(args, "-i")
			}
			args = append(args, r.BuildFlags...)
			args = append(args, "-o", r.FullBuildPath(), r.BuildTargetPath)
			cmd := exec.Command(envy.Get("GO_BIN", "go"), args...)
			payload["cmd"] = cmd.Args

			events.EmitPayload(EvtBuildStarted, payload)

			err := r.runAndListen(cmd)
			if err != nil {
				events.EmitError(EvtErrBuild, err, payload)
				if strings.Contains(err.Error(), "no buildable Go source files") {
					r.cancelFunc()
					log.Fatal(err)
				}
				return err
			}

			tt := time.Since(now)
			payload["pid"] = cmd.Process.Pid
			payload["build_time"] = tt
			events.EmitPayload(EvtBuildFinished, payload)
			r.Logger.Success("Building Completed (PID: %d) (Time: %s)", cmd.Process.Pid, tt)
			r.Restart <- true
			return nil
		})
	})
}

func (r *Manager) buildTransaction(fn func() error) {
	lpath := ErrorLogPath()
	err := fn()
	if err != nil {
		f, _ := os.Create(lpath)
		fmt.Fprint(f, err)
		r.Logger.Error("Error!")
		r.Logger.Error(err)
	} else {
		os.Remove(lpath)
	}
}
