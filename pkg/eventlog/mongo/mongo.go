package mongo

import (
	"errors"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/gomods/athens/pkg/eventlog"
)

var (
	// ErrNoIDReturned is error returned if mongo does not generate unique id for entry
	ErrNoIDReturned = errors.New("Mongo returned empty _id")
)

// Log is event log fetched from olympus server
type Log struct {
	s   *mgo.Session
	d   string // database
	c   string // collection
	url string
}

// NewLog creates log reader from remote log
func NewLog(url string) (*Log, error) {
	return NewLogWithCollection(url, "eventlog")
}

// NewLogWithCollection creates log reader from remote log
func NewLogWithCollection(url, collection string) (*Log, error) {
	m := &Log{
		url: url,
		c:   collection,
		d:   "athens",
	}
	return m, m.Connect()
}

// Connect establishes a session to the mongo cluster.
func (m *Log) Connect() error {
	s, err := mgo.Dial(m.url)
	if err != nil {
		panic(err)
	}
	m.s = s

	return nil
}

// Read reads all events in event log.
func (m *Log) Read() ([]eventlog.Event, error) {
	var events []eventlog.Event

	c := m.s.DB(m.d).C(m.c)
	err := c.Find(nil).All(&events)

	return events, err
}

// ReadFrom reads all events from the log starting at event with specified id (excluded).
// If id is not found behaves like Read().
func (m *Log) ReadFrom(id string) ([]eventlog.Event, error) {
	var events []eventlog.Event

	c := m.s.DB(m.d).C(m.c)
	oid := bson.ObjectId(id)
	err := c.Find(bson.M{"_id": bson.M{"$gt": oid}}).All(&events)

	return events, err
}

// Write appends Event to event log and returns its ID.
func (m *Log) Write(event eventlog.Event) (string, error) {
	event.ID = "" // clear id so mongo generates correct one
	c := m.s.DB(m.d).C(m.c)
	info, err := c.Upsert(nil, event)

	if err != nil {
		return "", err
	}

	if info.UpsertedId == nil {
		return "", ErrNoIDReturned
	}

	objectID := info.UpsertedId.(bson.ObjectId)
	return objectID.Hex(), nil
}

// Clear is a method for clearing entire state of event log
func (m *Log) Clear(id string) error {
	oid := bson.ObjectId(id)
	c := m.s.DB(m.d).C(m.c)
	_, err := c.RemoveAll(bson.M{"_id": bson.M{"$lte": oid}})
	return err
}
