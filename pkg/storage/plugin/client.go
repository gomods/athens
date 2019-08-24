package plugin

import (
	"context"
	"log"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/pkg/errors"

	"google.golang.org/grpc"

	stpb "github.com/gomods/athens/pkg/storage/plugin/pb/v1/storage"
)

type Conn struct {
	c    stpb.StorageBackendServiceClient
	conn *grpc.ClientConn
	canc context.CancelFunc
}

type LocalConn struct {
	Conn
	cmd *exec.Cmd
}

// DialLocal storage that will be served from a plugin
func DialLocal(ctx context.Context, plugin, unixSock, config string) (LocalConn, error) {
	bin, err := exec.LookPath(plugin)
	if err != nil {
		return LocalConn{}, err
	}
	ctx, canc := context.WithCancel(ctx)
	cFunc := func() *exec.Cmd {
		cmd := exec.CommandContext(ctx, bin, []string{"-sock", unixSock, "-config", config}...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd
	}
	cmd := cFunc()
	if err := cmd.Start(); err != nil {
		log.Println(errors.Wrap(err, "plugin failed to initialize"))
		canc()
		return LocalConn{}, err
	}
	conn, err := grpc.DialContext(ctx, unixSock,
		grpc.WithDialer(dialer),
		grpc.WithInsecure(),
		grpc.WithBackoffMaxDelay(1*time.Second),
		grpc.WithBlock(),
	)
	if err != nil {
		canc()
		return LocalConn{}, err
	}
	p := LocalConn{
		Conn: Conn{
			c:    stpb.NewStorageBackendServiceClient(conn),
			conn: conn,
			canc: canc,
		},
		cmd: cmd,
	}
	go p.watchPlugin(ctx, cFunc)
	return p, nil
}

func dialer(addr string, timeout time.Duration) (net.Conn, error) {
	return net.Dial("unix", addr)
}

// Close the connection to the plugin
func (p LocalConn) Close() error {
	p.canc()
	return p.conn.Close()
}

// watchPlugin ensures that plugin is restarted if it crashes.
func (p LocalConn) watchPlugin(ctx context.Context, f func() *exec.Cmd) {
	for {
		if err := p.cmd.Wait(); err != nil {
			log.Println(errors.Wrap(err, "plugin exited"))
		}
		select {
		case <-ctx.Done():
			return
		default:
		}
		p.cmd = f()
		if err := p.cmd.Start(); err != nil {
			log.Println(errors.Wrap(err, "plugin failed to start"))
		}
	}
}
