package module

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/gomods/athens/pkg/errors"
)

var (
	pathSeparator = "/"
)

// Filter is a filter of modules
type Filter struct {
	root     ruleNode
	filePath string
}

// NewFilter creates new filter based on rules defined in a configuration file
// WARNING: this is not concurrently safe
// Configuration consists of two operations: + for include and - for exclude
// e.g.
//    - github.com/a
//    + github.com/a/b
// will communicate all modules except github.com/a and its children, but github.com/a/b will be communicated
// example 2:
//   -
//   + github.com/a
// will exclude all items from communication except github.com/a
func NewFilter(filterFilePath string) (*Filter, error) {
	// Do not return an error if the file path is empty
	// Do not attempt to parse it as well.
	if filterFilePath == "" {
		return nil, nil
	}

	return initFromConfig(filterFilePath)

}

// AddRule adds rule for specified path
func (f *Filter) AddRule(path string, rule FilterRule) {
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

// Rule returns the filter rule to be applied to the given path
func (f *Filter) Rule(path string) FilterRule {
	segs := getPathSegments(path)
	rule := f.getAssociatedRule(segs...)
	if rule == Default {
		rule = Include
	}

	return rule
}

func (f *Filter) ensurePath(path string) {
	latest := f.root.next
	pathSegments := getPathSegments(path)

	for _, p := range pathSegments {
		if _, ok := latest[p]; !ok {
			latest[p] = newRule(Default)
		}
		latest = latest[p].next
	}
}

func (f *Filter) getAssociatedRule(path ...string) FilterRule {
	if len(path) == 0 {
		return f.root.rule
	}

	rules := make([]FilterRule, 0, len(path))
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

func initFromConfig(filePath string) (*Filter, error) {
	const op errors.Op = "module.initFromConfig"
	lines, err := getConfigLines(filePath)
	if err != nil {
		return nil, err
	}

	rn := newRule(Default)
	f := &Filter{
		filePath: filePath,
	}
	f.root = rn

	for idx, line := range lines {

		// Ignore newline
		if len(line) == 0 {
			continue
		}
		if len(line) > 0 && line[0] == '#' {
			continue
		}

		split := strings.Split(line, " ")
		if len(split) > 2 {
			return nil, errors.E(op, "Invalid configuration found in filter file at the line "+strconv.Itoa(idx+1))
		}

		ruleSign := strings.TrimSpace(split[0])
		rule := Default
		switch ruleSign {
		case "+":
			rule = Include
		case "-":
			rule = Exclude
		case "D":
			rule = Direct
		default:
			return nil, errors.E(op, "Invalid configuration found in filter file at the line "+strconv.Itoa(idx+1))
		}
		// is root config
		if len(split) == 1 {
			f.AddRule("", rule)
			continue
		}

		path := strings.TrimSpace(split[1])
		f.AddRule(path, rule)
	}
	return f, nil
}

func getPathSegments(path string) []string {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, pathSeparator)
	if path == "" {
		return []string{}
	}
	return strings.Split(path, pathSeparator)
}

func newRule(r FilterRule) ruleNode {
	rn := ruleNode{}
	rn.next = make(map[string]ruleNode)
	rn.rule = r
	return rn
}

func getConfigLines(filterFile string) ([]string, error) {
	const op errors.Op = "module.getConfigLines"

	f, err := os.Open(filterFile)
	if err != nil {
		return nil, errors.E(op, err)
	}

	scanner := bufio.NewScanner(f)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}

	return lines, f.Close()
}
