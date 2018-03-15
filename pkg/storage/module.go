package storage

type Module struct {
	BaseURL string
	Module  string
	Version string
	Mod     []byte
	Zip     []byte
}
