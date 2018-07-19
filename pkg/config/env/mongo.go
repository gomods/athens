package env

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gobuffalo/envy"
	"github.com/gomods/athens/pkg/storage/mongo"
)

const (
	defaultMongoTimeoutSec = "10"
)

// MongoURI returns Athens Mongo Storage URI defined by ATHENS_MONGO_STORAGE_URL
func MongoURI() (string, error) {
	env, err := envy.MustGet("ATHENS_MONGO_STORAGE_URL")
	if err != nil {
		return "", fmt.Errorf("missing mongo URL: %s", err)
	}

	return env, nil
}

// MongoHost returns Athens Mongo host defined by MONGO_HOST
func MongoHost() (string, error) {
	env, err := envy.MustGet("MONGO_HOST")
	if err != nil {
		return "", fmt.Errorf("missing mongo host: %s", err)
	}

	return env, nil
}

// MongoPort returns Athens Mongo port defined by MONGO_PORT
func MongoPort() (string, error) {
	env, err := envy.MustGet("MONGO_PORT")
	if err != nil {
		return "", fmt.Errorf("missing mongo port: %s", err)
	}

	return env, nil
}

// MongoUser returns Athens Mongo Storage user defined by MONGO_USER
func MongoUser() (string, error) {
	env, err := envy.MustGet("MONGO_USER")
	if err != nil {
		return "", fmt.Errorf("missing mongo user: %s", err)
	}

	return env, nil
}

// MongoPassword returns Athens Mongo Storage user password defined by MONGO_PASSWORD
func MongoPassword() (string, error) {
	env, err := envy.MustGet("MONGO_PASSWORD")
	if err != nil {
		return "", fmt.Errorf("missing mongo user password: %s", err)
	}

	return env, nil
}

// MongoConnectionTimeoutWithDefault returns Athens Mongo Storage connection timeout defined by MONGO_CONN_TIMEOUT_SEC.
// Values are in seconds.
func MongoConnectionTimeoutWithDefault(value string) string {
	return envy.Get("MONGO_CONN_TIMEOUT_SEC", value)
}

// MongoSSLWithDefault returns Athens Mongo Storage SSL flag defined by MONGO_SSL.
// Defines whether or not SSL should be used.
func MongoSSLWithDefault(value string) string {
	return envy.Get("MONGO_SSL", value)
}

// MongoConnDetails fetches appropriate environment variables from other
// Mongo* functions and returns a mongo.ConnDetails according to their
// values. Otherwise, returns nil and a non-nil error
func MongoConnDetails() (*mongo.ConnDetails, error) {
	host, err := MongoHost()
	if err != nil {
		return nil, err
	}
	portStr, err := MongoPort()
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	user, err := MongoUser()
	if err != nil {
		return nil, err
	}

	password, err := MongoPassword()
	if err != nil {
		return nil, err
	}

	timeoutSecStr := MongoConnectionTimeoutWithDefault(defaultMongoTimeoutSec)
	timeoutSec, err := strconv.Atoi(timeoutSecStr)
	if err != nil {
		return nil, err
	}

	sslStr := MongoSSLWithDefault("true")
	ssl, err := strconv.ParseBool(sslStr)
	if err != nil {
		return nil, err
	}

	return &mongo.ConnDetails{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Timeout:  time.Duration(timeoutSec) * time.Second,
		SSL:      ssl,
	}, nil
}
