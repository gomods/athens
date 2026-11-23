package mongo

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// moduleDocument is the shape of a module document stored in the mongo
// collection. ID is the mongo-specific primary key assigned on insert.
type moduleDocument struct {
	ID      bson.ObjectID `bson:"_id,omitempty"`
	Module  string        `bson:"module"`
	Version string        `bson:"version"`
	Mod     []byte        `bson:"mod"`
	Info    []byte        `bson:"info"`
}
