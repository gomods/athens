package testutil

import (
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// GetServicePort returns the host and port for a service from docker-compose
func GetServicePort(t *testing.T, service string, containerPort int) (host string, hostPort int) {
	t.Helper()
	cPort := strconv.Itoa(containerPort)
	out, err := exec.Command("docker-compose", "-p", "athensdev", "port", service, cPort).Output()
	require.NoError(t, err)
	addr := strings.TrimSpace(string(out))
	parts := strings.Split(addr, ":")
	require.Lenf(t, parts, 2, "invalid address %q", addr)
	hPort, err := strconv.Atoi(parts[1])
	require.NoError(t, err)
	return parts[0], hPort
}
