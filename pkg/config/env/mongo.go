package env

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gobuffalo/envy"
	"github.com/gomods/athens/pkg/storage/mongo/conn"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
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

// MongoConnDetails returns mongo connection details defined by all the Mongo
// environment variables
func MongoConnDetails() (*conn.Details, error) {
	var errs error
	host, err := MongoHost()
	if err != nil {
		multierror.Append(errs, err)
	}
	portStr, err := MongoPort()
	if err != nil {
		multierror.Append(errs, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		multierror.Append(errs, err)
	}
	user, err := MongoUser()
	if err != nil {
		multierror.Append(errs, err)
	}
	pass, err := MongoPassword()
	if err != nil {
		multierror.Append(errs, err)
	}
	timeoutStr := MongoConnectionTimeoutWithDefault("1")
	timeoutSec, err := strconv.Atoi(timeoutStr)
	if err != nil {
		multierror.Append(errs, errors.WithMessage(err, "invalid value for timeout"))
	}
	timeout := time.Duration(timeoutSec) * time.Second
	sslStr := MongoSSLWithDefault("true")
	ssl, err := strconv.ParseBool(sslStr)
	if err != nil {
		multierror.Append(errs, errors.WithMessage(err, "invalid value for mongo SSL"))
	}

	if errs != nil {
		return nil, errs
	}
	return &conn.Details{
		Host:     host,
		Port:     port,
		User:     user,
		Password: pass,
		Timeout:  timeout,
		SSL:      ssl,
	}, nil
}
