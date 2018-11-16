package sessions

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var _ sessions.Store = Null{}

// Null implements the sessions.Store interface, github.com/gorilla/sessions,
// but does nothing with it. Perfect for APIs that don't need sessions.
type Null struct{}

func (n Null) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.NewSession(n, name), nil
}

func (n Null) New(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.NewSession(n, name), nil
}

func (n Null) Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
	return nil
}
