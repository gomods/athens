package env

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gobuffalo/envy"
	"github.com/gomods/athens/pkg/storage/mongo"
)

// ForMongo reads env vars for the mongo connection and returns a ConnDetails
// representation of them. If env vars are not formatted or missing, returns
// an error
func ForMongo() (*mongo.ConnDetails, error) {
	host, err := envy.MustGet("MONGO_HOST")
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(envy.Get("MONGO_PORT", "27017"))
	if err != nil {
		return nil, fmt.Errorf("invalid MONGO_PORT (%s)", err)
	}
	user, err := envy.MustGet("MONGO_USER")
	if err != nil {
		return nil, err
	}
	pass, err := envy.MustGet("MONGO_PASSWORD")
	if err != nil {
		return nil, err
	}
	timeoutSec, err := strconv.Atoi(envy.Get("MONGO_CONN_TIMEOUT_SEC", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid MONGO_CONN_TIMEOUT_SEC (%s)", err)
	}
	sslBool, err := strconv.ParseBool(envy.Get("MONGO_SSL", "true"))
	if err != nil {
		return nil, fmt.Errorf("invalid MONGO_SSL (%s)", err)
	}
	details := &mongo.ConnDetails{
		Host:     host,
		Port:     port,
		User:     user,
		Password: pass,
		Timeout:  time.Duration(timeoutSec) * time.Second,
		SSL:      sslBool,
	}
	return details, nil
}
