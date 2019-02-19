package module

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/gomods/athens/pkg/errors"
)

var (
	pathSeparator    = "/"
	versionSeparator = "."
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
func (f *Filter) AddRule(path string, qualifiers []string, rule FilterRule) {
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
	rn.qualifiers = qualifiers
	latest.next[last] = rn
}

// Rule returns the filter rule to be applied to the given path
func (f *Filter) Rule(path, version string) FilterRule {
	segs := getPathSegments(path)
	rule := f.getAssociatedRule(version, segs...)
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

func (f *Filter) getAssociatedRule(version string, path ...string) FilterRule {
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
		// default to true if no version filter, false otherwise
		match := len(rn.qualifiers) == 0
		for _, q := range rn.qualifiers {
			if matches(version, q) {
				match = true
				break
			}
		}
		if match || version == "" {
			rules = append(rules, rn.rule)
		}
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
		if len(split) > 3 {
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
			f.AddRule("", nil, rule)
			continue
		}
		var vers []string
		if len(split) == 3 {
			vers = strings.Split(split[2], ",")
			for i := range vers {
				vers[i] = strings.TrimRight(vers[i], "*")
				if vers[i][len(vers[i])-1] != '.' && strings.Count(vers[i], ".") < 2 {
					vers[i] += "."
				}
			}
		}

		path := strings.TrimSpace(split[1])
		f.AddRule(path, vers, rule)
	}
	return f, nil
}

// matches checks if the given version matches the given qualifier.
// Qualifiers can be:
// - plain versions
// - v1.2.3 enables v1.2.3
// - ~1.2.3: enables 1.2.x  which are at least 1.2.3
// - ^1.2.3: enables 1.x.x which are at least 1.2.3
// - <1.2.3: enables everything lower than 1.2.3 includes 1.2.2 and 0.58.9 as well
func matches(version, qualifier string) bool {
	prefix := qualifier[0]

	// v1.2.3 means we accept every version starting with v.1.2.3
	if prefix == 'v' {
		return strings.HasPrefix(version, qualifier)
	}

	v, err := getVersionSegments(version[1:])
	if err != nil {
		return false
	}

	q, err := getVersionSegments(qualifier[1:])
	if err != nil {
		return false
	}

	if len(v) != len(q) {
		return false
	}
	// no semver
	if len(v) != 3 || len(q) != 3 {
		return false
	}

	switch prefix {
	case '~':
		if v[0] == q[0] && v[1] == q[1] && v[2] >= q[2] {
			return true
		}
		return false
	case '^':
		if v[0] == q[0] && v[1] >= q[1] {
			return true
		}
		if v[0] == q[0] && v[1] == q[1] && v[2] >= q[2] {
			return true
		}
		return false
	case '<':
		if v[0] < q[0] {
			return true
		}
		if v[0] == q[0] && v[1] < q[1] {
			return true
		}
		if v[0] == q[0] && v[1] == q[1] && v[2] <= q[2] {
			return true
		}
		return false
	}
	return false
}

func getPathSegments(path string) []string {
	return getSegments(path, pathSeparator)
}

func getVersionSegments(path string) ([]int, error) {
	vv := getSegments(path, versionSeparator)
	res := make([]int, len(vv))
	for i, v := range vv {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		res[i] = n
	}
	return res, nil
}

func getSegments(path, separator string) []string {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, separator)
	if path == "" {
		return []string{}
	}
	return strings.Split(path, separator)
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
