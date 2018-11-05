package module

// FilterRule defines behavior of module communication
type FilterRule int

const (
	// Default filter rule does not alter default/parent behavior
	Default FilterRule = iota
	// Include treats modules the usual way
	// Used for reverting Exclude of parent path
	Include
	// Exclude filter rule excludes package and its children from communication
	Exclude
	// Direct filter rule forces the package to be fetched directly from upstream proxy
	Direct
)
