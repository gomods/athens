package pop

// AvailableDialects lists the available database dialects
var AvailableDialects = []string{}

// DialectSupported checks support for the given database dialect
func DialectSupported(d string) bool {
	for _, ad := range AvailableDialects {
		if ad == d {
			return true
		}
	}
	return false
}
