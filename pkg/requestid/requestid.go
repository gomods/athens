package requestid

import "context"

// HeaderKey is the header key that athens uses
// to pass request ids into logs and outbound requests.
const HeaderKey = "Athens-Request-ID"

type key struct{}

// SetInContext sets the given requestID into the context.
func SetInContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, key{}, id)
}

// FromContext returns a requestID from the context or an empty
// string if not found.
func FromContext(ctx context.Context) string {
	id, _ := ctx.Value(key{}).(string)
	return id
}
