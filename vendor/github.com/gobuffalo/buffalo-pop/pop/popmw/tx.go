package popmw

import (
	"time"

	"github.com/pkg/errors"

	"github.com/gobuffalo/buffalo"
	pp "github.com/gobuffalo/buffalo-pop/pop"
	"github.com/gobuffalo/events"
	"github.com/gobuffalo/pop"
)

// PopTransaction is a piece of Buffalo middleware that wraps each
// request in a transaction. The transaction will automatically get
// committed if there's no errors and the response status code is a
// 2xx or 3xx, otherwise it'll be rolled back. It will also add a
// field to the log, "db", that shows the total duration spent during
// the request making database calls.
func Transaction(db *pop.Connection) buffalo.MiddlewareFunc {
	events.NamedListen("popmw.Transaction", func(e events.Event) {
		if e.Kind != "buffalo:app:start" {
			return
		}
		i, err := e.Payload.Pluck("app")
		if err != nil {
			return
		}
		if app, ok := i.(*buffalo.App); ok {
			pop.SetLogger(pp.Logger(app))
		}
	})
	return func(h buffalo.Handler) buffalo.Handler {
		return func(c buffalo.Context) error {
			// wrap all requests in a transaction and set the length
			// of time doing things in the db to the log.
			err := db.Transaction(func(tx *pop.Connection) error {
				start := tx.Elapsed
				defer func() {
					finished := tx.Elapsed
					elapsed := time.Duration(finished - start)
					c.LogField("db", elapsed)
				}()
				c.Set("tx", tx)
				if err := h(c); err != nil {
					return err
				}
				if res, ok := c.Response().(*buffalo.Response); ok {
					if res.Status < 200 || res.Status >= 400 {
						return errNonSuccess
					}
				}
				return nil
			})
			if err != nil && errors.Cause(err) != errNonSuccess {
				return err
			}
			return nil
		}
	}
}

var errNonSuccess = errors.New("non success status code")
