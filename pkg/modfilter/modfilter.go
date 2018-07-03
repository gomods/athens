package modfilter

import (
	"strings"
)

var (
	pathSeparator = "/"
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

func getPathSegments(path string) []string {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, pathSeparator)

	return strings.Split(path, pathSeparator)
}

func newRule(r Rule) ruleNode {
	rn := ruleNode{}
	rn.next = make(map[string]ruleNode)
	rn.rule = r

	return rn
}
