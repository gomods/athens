package mongo

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ModuleStore represents a mongo backed storage backend.
type ModuleStore struct {
	client   *mongo.Client
	db       string // database
	coll     string // collection
	url      string
	certPath string
	insecure bool // Only to be used for development instances
	timeout  time.Duration
}

// NewStorage returns a connected Mongo backed storage
// that satisfies the Backend interface.
func NewStorage(conf *config.MongoConfig, timeout time.Duration) (*ModuleStore, error) {
	const op errors.Op = "mongo.NewStorage"
	if conf == nil {
		return nil, errors.E(op, "No Mongo Configuration provided")
	}
	ms := &ModuleStore{url: conf.URL, certPath: conf.CertPath, timeout: timeout, insecure: conf.InsecureConn}
	client, err := ms.newClient(conf)
	ms.client = client

	if err != nil {
		return nil, errors.E(op, err)
	}

	_, err = ms.connect(conf)

	if err != nil {
		return nil, errors.E(op, err)
	}

	return ms, nil
}

func (m *ModuleStore) connect(conf *config.MongoConfig) (*mongo.Collection, error) {
	const op errors.Op = "mongo.connect"

	var err error
	err = m.client.Connect(context.Background())

	if err != nil {
		return nil, errors.E(op, err)
	}

	return m.initDatabase(), nil
}

func (m *ModuleStore) initDatabase() *mongo.Collection {
	// TODO: database and collection as env vars, or params to New()? together with user/mongo
	m.db = "athens"
	m.coll = "modules"

	c := m.client.Database(m.db).Collection(m.coll)
	indexView := c.Indexes()
	keys := make(map[string]int)
	keys["base_url"] = 1
	keys["module"] = 1
	keys["version"] = 1
	indexOptions := options.Index().SetBackground(true).SetSparse(true).SetUnique(true)
	indexView.CreateOne(context.Background(), mongo.IndexModel{Keys: keys, Options: indexOptions}, options.CreateIndexes())

	return c
}

func (m *ModuleStore) newClient(conf *config.MongoConfig) (*mongo.Client, error) {
	tlsConfig := &tls.Config{}
	clientOptions := options.Client()
	// Maybe check for error using Validate()?
	clientOptions = clientOptions.ApplyURI(m.url)

	if m.certPath != "" {
		// Sets only when the env var is setup in config.dev.toml
		tlsConfig.InsecureSkipVerify = m.insecure
		var roots *x509.CertPool
		// See if there is a system cert pool
		roots, err := x509.SystemCertPool()
		if err != nil {
			// If there is no system cert pool, create a new one
			roots = x509.NewCertPool()
		}

		cert, err := ioutil.ReadFile(m.certPath)
		if err != nil {
			return nil, err
		}

		if ok := roots.AppendCertsFromPEM(cert); !ok {
			return nil, fmt.Errorf("failed to parse certificate from: %s", m.certPath)
		}

		tlsConfig.ClientCAs = roots
		clientOptions = clientOptions.SetTLSConfig(tlsConfig)
	}
	clientOptions = clientOptions.SetConnectTimeout(m.timeout)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (m *ModuleStore) gridFileName(mod, ver string) string {
	return strings.Replace(mod, "/", "_", -1) + "_" + ver + ".zip"
}
