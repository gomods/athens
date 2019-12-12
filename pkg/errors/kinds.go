package errors

// IsNotFoundErr helper function for KindNotFound
func IsNotFoundErr(err error) bool {
	return Kind(err) == KindNotFound
}

// IsConfigErr returns true if the given err is a configuration error
func IsConfigErr(err error) bool {
	_, ok := err.(*configErr)
	return ok
}
