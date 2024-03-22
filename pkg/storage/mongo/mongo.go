package mongo

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
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
	client, err := ms.newClient()
	if err != nil {
		return nil, errors.E(op, err)
	}

	ms.client = client
	ms.db = conf.DefaultDBName
	ms.coll = conf.DefaultCollectionName

	_ = ms.initDatabase()

	return ms, nil
}

func (s *ModuleStore) initDatabase() *mongo.Collection {
	if s.db == "" {
		s.db = "athens"
	}

	if s.coll == "" {
		s.coll = "modules"
	}

	c := s.client.Database(s.db).Collection(s.coll)
	indexView := c.Indexes()
	keys := bson.D{
		{Key: "base_url", Value: 1},
		{Key: "module", Value: 1},
		{Key: "version", Value: 1},
	}
	indexOptions := options.Index().SetSparse(true).SetUnique(true)
	_, _ = indexView.CreateOne(context.Background(), mongo.IndexModel{Keys: keys, Options: indexOptions}, options.CreateIndexes())

	return c
}

func (s *ModuleStore) newClient() (*mongo.Client, error) {
	const op errors.Op = "mongo.newClient"

	tlsConfig := &tls.Config{}
	clientOptions := options.Client()
	clientOptions = clientOptions.ApplyURI(s.url)

	err := clientOptions.Validate()
	if err != nil {
		return nil, errors.E(op, err)
	}

	if s.certPath != "" {
		// Sets only when the env var is setup in config.dev.toml
		tlsConfig.InsecureSkipVerify = s.insecure
		var roots *x509.CertPool
		// See if there is a system cert pool
		roots, err := x509.SystemCertPool()
		if err != nil {
			// If there is no system cert pool, create a new one
			roots = x509.NewCertPool()
		}

		cert, err := os.ReadFile(s.certPath)
		if err != nil {
			return nil, errors.E(op, err)
		}

		if ok := roots.AppendCertsFromPEM(cert); !ok {
			return nil, fmt.Errorf("failed to parse certificate from: %s", s.certPath)
		}

		tlsConfig.ClientCAs = roots
		clientOptions = clientOptions.SetTLSConfig(tlsConfig)
	}
	clientOptions = clientOptions.SetConnectTimeout(s.timeout)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return client, nil
}

func (s *ModuleStore) gridFileName(mod, ver string) string {
	return strings.ReplaceAll(mod, "/", "_") + "_" + ver + ".zip"
}
