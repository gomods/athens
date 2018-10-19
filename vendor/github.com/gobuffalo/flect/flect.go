package flect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func init() {
	pwd, _ := os.Getwd()
	cfg := filepath.Join(pwd, "inflections.json")
	if p := os.Getenv("INFLECT_PATH"); p != "" {
		cfg = p
	}
	if _, err := os.Stat(cfg); err == nil {
		b, err := ioutil.ReadFile(cfg)
		if err != nil {
			fmt.Printf("could not read inflection file %s (%s)\n", cfg, err)
			return
		}
		if err = LoadReader(bytes.NewReader(b)); err != nil {
			fmt.Println(err)
		}
	}
}

//LoadReader loads rules from io.Reader param
func LoadReader(r io.Reader) error {
	m := map[string]string{}
	err := json.NewDecoder(r).Decode(&m)
	if err != nil {
		return fmt.Errorf("could not decode inflection JSON from reader: %s", err)
	}
	pluralMoot.Lock()
	defer pluralMoot.Unlock()
	singularMoot.Lock()
	defer singularMoot.Unlock()

	for s, p := range m {
		singleToPlural[s] = p
		pluralToSingle[p] = s
	}
	return nil
}

var spaces = []rune{'_', ' ', ':', '-', '/'}

func isSpace(c rune) bool {
	for _, r := range spaces {
		if r == c {
			return true
		}
	}
	return unicode.IsSpace(c)
}

func xappend(a []string, ss ...string) []string {
	for _, s := range ss {
		s = strings.TrimSpace(s)
		for _, x := range spaces {
			s = strings.Trim(s, string(x))
		}
		if _, ok := baseAcronyms[strings.ToUpper(s)]; ok {
			s = strings.ToUpper(s)
		}
		if s != "" {
			a = append(a, s)
		}
	}
	return a
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
