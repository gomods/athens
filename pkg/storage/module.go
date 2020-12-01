package storage

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Module represents a vgo module saved in a storage backend.
type Module struct {
	// TODO(marwan-at-work): ID is a mongo-specific field, it should not be
	// in the generic storage.Module struct.
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Module  string             `bson:"module"`
	Version string             `bson:"version"`
	Mod     []byte             `bson:"mod"`
	Info    []byte             `bson:"info"`
}
