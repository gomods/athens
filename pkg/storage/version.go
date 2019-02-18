package storage

// Version represents a version of a module and contains .mod file, a .info file and zip file of a specific version
type Version struct {
	Mod    []byte
	Zip    Zip
	Info   []byte
	Semver string
}
