package plugdeps

import (
	"io"
	"sort"
	"sync"

	toml "github.com/pelletier/go-toml"
	"github.com/pkg/errors"
)

// Plugins manages the config/buffalo-plugins.toml file
// as well as the plugins available from the file.
type Plugins struct {
	plugins map[string]Plugin
	moot    *sync.RWMutex
}

// Encode the list of plugins, in TOML format, to the reader
func (plugs *Plugins) Encode(w io.Writer) error {
	tp := tomlPlugins{
		Plugins: plugs.List(),
	}

	if err := toml.NewEncoder(w).Encode(tp); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Decode the list of plugins, in TOML format, from the reader
func (plugs *Plugins) Decode(r io.Reader) error {
	tp := &tomlPlugins{
		Plugins: []Plugin{},
	}
	if err := toml.NewDecoder(r).Decode(tp); err != nil {
		return errors.WithStack(err)
	}
	for _, p := range tp.Plugins {
		plugs.Add(p)
	}
	return nil
}

// List of dependent plugins listed in order of Plugin.String()
func (plugs *Plugins) List() []Plugin {
	m := map[string]Plugin{}
	plugs.moot.RLock()
	for _, p := range plugs.plugins {
		m[p.String()] = p
	}
	plugs.moot.RUnlock()
	var pp []Plugin
	for _, v := range m {
		pp = append(pp, v)
	}
	sort.Slice(pp, func(a, b int) bool {
		return pp[a].Binary < pp[b].Binary
	})
	return pp
}

// Add plugin(s) to the list of dependencies
func (plugs *Plugins) Add(pp ...Plugin) {
	plugs.moot.Lock()
	for _, p := range pp {
		plugs.plugins[p.String()] = p
	}
	plugs.moot.Unlock()
}

// Remove plugin(s) from the list of dependencies
func (plugs *Plugins) Remove(pp ...Plugin) {
	plugs.moot.Lock()
	for _, p := range pp {
		delete(plugs.plugins, p.String())
	}
	plugs.moot.Unlock()
}

// New returns a configured *Plugins value
func New() *Plugins {
	return &Plugins{
		plugins: map[string]Plugin{},
		moot:    &sync.RWMutex{},
	}
}

type tomlPlugins struct {
	Plugins []Plugin `toml:"plugin"`
}
