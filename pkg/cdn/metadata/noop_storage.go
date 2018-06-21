package metadata

type noOpStorage struct {
	s Storage
}

// NoOpStorage wraps storage Storage with Connect functionality
func NoOpStorage(s Storage) Storage {
	return noOpStorage{s: s}
}

func (n noOpStorage) Get(module string) (string, error) {
	return n.s.Get(module)
}

func (n noOpStorage) Save(module, redirectURL string) error {
	return n.s.Save(module, redirectURL)
}
