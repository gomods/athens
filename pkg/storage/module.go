package storage

import (
	"github.com/globalsign/mgo/bson"
)

// Module represents a vgo module saved in a storage backend.
type Module struct {
	ID      bson.ObjectId `bson:"_id"`
	Module  string        `bson:"module"`
	Version string        `bson:"version"`
	Mod     []byte        `bson:"mod"`
	Zip     []byte        `bson:"zip"`
	Info    []byte        `bson:"info"`
}
