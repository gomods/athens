package plugin

import (
	"crypto/tls"
	"errors"
	"flag"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/gomods/athens/pkg/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	stpb "github.com/gomods/athens/pkg/storage/plugin/pb/v1/storage"
)

var socket, configFile string

func init() {
	flag.StringVar(&socket, "sock", "/tmp/storage.sock", "plugin unix socket name")
	flag.StringVar(&configFile, "config", "/config.storage.toml", "plugin configuration from proxy")
}

type Storage struct {
	srv      *grpc.Server
	lis      net.Listener
	conf     config
	back     storage.Backend
	confFile string
}

// NewLocal for plugin storage on the same machine/container as the proxy server.
func NewLocal(back storage.Backend, opts ...Option) (Storage, error) {
	if !flag.Parsed() {
		flag.Parse()
	}
	socket, err := filepath.Abs(socket)
	if err != nil {
		return Storage{}, err
	}
	srv := grpc.NewServer()
	stpb.RegisterStorageBackendServiceServer(srv, backend{b: back})

	_ = os.Remove(socket) // don't fail if not found, net.Listen will check
	lis, err := net.Listen("unix", socket)
	if err != nil {
		return Storage{}, err
	}
	return Storage{srv: srv, lis: lis, confFile: configFile}, nil
}

// NewRemote for plugin storage served across network connections.
func NewRemote(back storage.Backend, host, port string, tlsConf *tls.Config) (Storage, error) {
	lis, err := net.Listen("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return Storage{}, err
	}
	opt := []grpc.ServerOption{}
	if tlsConf != nil {
		opt = append(opt, grpc.Creds(credentials.NewTLS(tlsConf)))
	}
	return Storage{
		lis:  lis,
		back: back,
		conf: config{opts: opt},
	}, nil
}

// Close the server
func (p *Storage) Close() error {
	log.Println("closing plugin")
	p.srv.Stop()
	return p.lis.Close()
}

// Serve the storage plugin.  This will allow the athens proxy to connect and serve from storage.
func (p *Storage) Serve() error {
	srv := grpc.NewServer(p.conf.opts...)
	stpb.RegisterStorageBackendServiceServer(srv, backend{b: p.back})
	return p.srv.Serve(p.lis)
}

// ConfigFile gets the configuration file name for local plugins.
// This allows passing the configuration file from the athens proxy config.
func (p *Storage) ConfigFile() (string, error) {
	if p.confFile == "" {
		return "", errors.New("no configuration file received")
	}
	return p.confFile, nil
}

type Option func(*config)

type config struct {
	srv  *grpc.Server
	opts []grpc.ServerOption
}

func WithServerOptions(opts ...grpc.ServerOption) Option {
	return func(cfg *config) {
		cfg.opts = append(cfg.opts, opts...)
	}
}
