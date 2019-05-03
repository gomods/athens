package storage

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Module represents a vgo module saved in a storage backend.
type Module struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Module  string             `bson:"module"`
	Version string             `bson:"version"`
	Mod     []byte             `bson:"mod"`
	Zip     []byte             `bson:"zip"`
	Info    []byte             `bson:"info"`
}
