package pop

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop/logging"
	"github.com/pkg/errors"

	"gopkg.in/yaml.v2"
)

// ErrConfigFileNotFound is returned when the pop config file can't be found,
// after looking for it.
var ErrConfigFileNotFound = errors.New("unable to find pop config file")

var lookupPaths = []string{"", "./config", "/config", "../", "../config", "../..", "../../config"}

// ConfigName is the name of the YAML databases config file
var ConfigName = "database.yml"

func init() {
	SetLogger(defaultLogger)

	ap := os.Getenv("APP_PATH")
	if ap != "" {
		AddLookupPaths(ap)
	}
	ap = os.Getenv("POP_PATH")
	if ap != "" {
		AddLookupPaths(ap)
	}
	if err := LoadConfigFile(); err != nil {
		// this is debug because there are a lot of cases where
		// this being logged as an error is causes problems
		// buffalo plugins, for one
		// also, it's ok to not always have a config file, like
		// in a new project where one hasn't be generated
		log(logging.Debug, "Unable to load config file: %v", err)
	}
}

// LoadConfigFile loads a POP config file from the configured lookup paths
func LoadConfigFile() error {
	path, err := findConfigPath()
	if err != nil {
		return errors.WithStack(err)
	}
	Connections = map[string]*Connection{}
	log(logging.Debug, "Loading config file from %s", path)
	f, err := os.Open(path)
	if err != nil {
		return errors.WithStack(err)
	}
	return LoadFrom(f)
}

// LookupPaths returns the current configuration lookup paths
func LookupPaths() []string {
	return lookupPaths
}

// AddLookupPaths add paths to the current lookup paths list
func AddLookupPaths(paths ...string) error {
	lookupPaths = append(paths, lookupPaths...)
	return LoadConfigFile()
}

func findConfigPath() (string, error) {
	for _, p := range LookupPaths() {
		path, _ := filepath.Abs(filepath.Join(p, ConfigName))
		if _, err := os.Stat(path); err == nil {
			return path, err
		}
	}
	return "", ErrConfigFileNotFound
}

// LoadFrom reads a configuration from the reader and sets up the connections
func LoadFrom(r io.Reader) error {
	envy.Load()
	tmpl := template.New("test")
	tmpl.Funcs(map[string]interface{}{
		"envOr": func(s1, s2 string) string {
			return envy.Get(s1, s2)
		},
		"env": func(s1 string) string {
			return envy.Get(s1, "")
		},
	})
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return errors.WithStack(err)
	}
	t, err := tmpl.Parse(string(b))
	if err != nil {
		return errors.Wrap(err, "couldn't parse config template")
	}

	var bb bytes.Buffer
	err = t.Execute(&bb, nil)
	if err != nil {
		return errors.Wrap(err, "couldn't execute config template")
	}

	deets := map[string]*ConnectionDetails{}
	err = yaml.Unmarshal(bb.Bytes(), &deets)
	if err != nil {
		return errors.Wrap(err, "couldn't unmarshal config to yaml")
	}
	for n, d := range deets {
		con, err := NewConnection(d)
		if err != nil {
			log(logging.Warn, "unable to load connection %s: %v", n, err)
			continue
		}
		Connections[n] = con
	}
	return nil
}
