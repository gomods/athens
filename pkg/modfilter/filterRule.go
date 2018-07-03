package modfilter

// Rule defines behavior of module communication
type Rule int

const (
	// Default filter rule does not alter default behavior
	Default Rule = iota
	// Include filter rule includes package and its children from communication
	Include
	// Exclude filter rule excludes package and its children from communication
	Exclude
)
