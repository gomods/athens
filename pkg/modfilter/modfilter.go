package modfilter

import (
	"bufio"
	"os"
	"strings"
)

var (
	pathSeparator         = "/"
	configurationFileName = "filter.conf"
)

// ModFilter is a filter of modules
type ModFilter struct {
	root ruleNode
}

// NewModFilter creates new filter based on rules defined in a configuration file
// Configuration consists of two operations + for include and - for exclude
// e.g.
//    - github.com/a
//    + github.com/a/b
// will communicate all modules except github.com/a and its children, but github.com/a/b will be communicated
// example 2:
//   -
//   + github.com/a
// will exclude all items from communication except github.com/a
func NewModFilter() *ModFilter {
	rn := newRule(Default)
	modFilter := ModFilter{}
	modFilter.root = rn

	modFilter.initFromConfig()

	return &modFilter
}

// AddRule adds rule for specified path
func (f *ModFilter) AddRule(path string, rule Rule) {
	f.ensurePath(path)

	segments := getPathSegments(path)

	if len(segments) == 0 {
		f.root.rule = rule
		return
	}

	// look for latest node in a path
	latest := f.root
	for _, p := range segments[:len(segments)-1] {
		latest = latest.next[p]
	}

	// replace with updated node
	last := segments[len(segments)-1]
	rn := latest.next[last]
	rn.rule = rule
	latest.next[last] = rn
}

// ShouldProcess evaluates path and determines if module should be communicated or not
func (f *ModFilter) ShouldProcess(path string) bool {
	segs := getPathSegments(path)
	rule := f.shouldProcess(segs...)

	// process everything unless it's excluded
	return rule != Exclude
}

func (f *ModFilter) ensurePath(path string) {
	latest := f.root.next
	pathSegments := getPathSegments(path)

	for _, p := range pathSegments {
		if _, ok := latest[p]; !ok {
			latest[p] = newRule(Default)
		}
		latest = latest[p].next
	}
}

func (f *ModFilter) shouldProcess(path ...string) Rule {
	if len(path) == 0 {
		return f.root.rule
	}

	rules := make([]Rule, 0, len(path))
	rn := f.root
	for _, p := range path {
		if _, ok := rn.next[p]; !ok {
			break
		}
		rn = rn.next[p]
		rules = append(rules, rn.rule)
	}

	if len(rules) == 0 {
		return f.root.rule
	}

	for i := len(rules) - 1; i >= 0; i-- {
		if rules[i] != Default {
			return rules[i]
		}
	}

	return f.root.rule
}

func (f *ModFilter) initFromConfig() {
	lines, err := getConfigLines()

	if err != nil || len(lines) == 0 {
		return
	}

	for _, line := range lines {
		split := strings.Split(line, " ")
		if len(split) > 2 {
			continue
		}
		ruleSign := strings.TrimSpace(split[0])
		rule := Default
		switch ruleSign {
		case "+":
			rule = Include
		case "-":
			rule = Exclude
		default:
			continue
		}

		// is root config
		if len(split) == 1 {
			f.AddRule("", rule)
		}

		path := strings.TrimSpace(split[1])
		f.AddRule(path, rule)
	}
}

func getPathSegments(path string) []string {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, pathSeparator)

	if path == "" {
		return []string{}
	}

	return strings.Split(path, pathSeparator)
}

func newRule(r Rule) ruleNode {
	rn := ruleNode{}
	rn.next = make(map[string]ruleNode)
	rn.rule = r

	return rn
}

func getConfigLines() ([]string, error) {
	configName := configurationFileName

	f, err := os.Open(configName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		lines = append(lines, line)
	}

	return lines, nil
}
