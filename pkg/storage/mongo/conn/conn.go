package conn

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/globalsign/mgo"
)

// Details represents connection details to a mongo database
type Details struct {
	Host     string
	Port     int
	User     string
	Password string
	Timeout  time.Duration
	SSL      bool
}

// NewSession takes connection details and a database name & dials
// to the mongo database
func NewSession(deets *Details, db string) (*mgo.Session, error) {
	dialInfo := &mgo.DialInfo{
		Addrs:    []string{fmt.Sprintf("%s:%d", deets.Host, deets.Port)},
		Timeout:  deets.Timeout,
		Database: db,
		Username: deets.User,
		Password: deets.Password,
	}
	if deets.SSL {
		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", addr.String(), &tls.Config{})
		}
	}
	return mgo.DialWithInfo(dialInfo)
}
