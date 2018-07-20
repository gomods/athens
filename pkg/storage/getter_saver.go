package storage

// GetterSaver is a Getter and a Saver in one
type GetterSaver interface {
	Getter
	Saver
}
