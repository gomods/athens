package events

import (
	"context"
)

//go:generate stringer -type=Type

// Type describe various event types
type Type int

// HeaderKey is the HTTP Header that Athens will send
// along every event. This helps you know which JSON shape
// to use when parsing a request body
const HeaderKey = "Athens-Event"

// Event types
const (
	Ping Type = iota + 1
	Stashed
)

// BaseEvent is the common data that all
// event payloads are composed of.
type BaseEvent struct {
	Event   string
	Version string
}

// PingEvent describes the payload for a Ping event
type PingEvent struct {
	BaseEvent
}

// StashedEvent describes the payload for the Stashed event
type StashedEvent struct {
	BaseEvent
	Module, Version string
}

// Hook describes a service that can be used to send events to
type Hook interface {
	// Ping pings the underlying server to ensure that
	// the event hook url is ready to receive requests
	Ping(ctx context.Context) error

	// Stashed is called whenever a new module is succesfully persisted
	// to the storage Backend
	Stashed(ctx context.Context, mod, ver string) error
}
