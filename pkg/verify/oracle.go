package verify

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/mod/module"
)

// ErrNotInSumDB indicates the module version is absent from the checksum
// database (e.g. a private module). Callers must treat this as "cannot
// verify" and keep the module.
var ErrNotInSumDB = errors.New("module version not found in checksum database")

// Oracle returns the canonical h1: zip hash for a module version.
type Oracle interface {
	CanonicalHash(ctx context.Context, mod, ver string) (string, error)
}

// sumdbOracle looks up the canonical hash from a Go checksum database
// (e.g. https://sum.golang.org) via its /lookup endpoint.
type sumdbOracle struct {
	baseURL string
	client  *http.Client
}

// NewSumDBOracle returns an Oracle backed by the checksum database at baseURL.
func NewSumDBOracle(baseURL string, client *http.Client) Oracle {
	return &sumdbOracle{baseURL: strings.TrimRight(baseURL, "/"), client: client}
}

func (o *sumdbOracle) CanonicalHash(ctx context.Context, mod, ver string) (string, error) {
	escMod, err := module.EscapePath(mod)
	if err != nil {
		return "", fmt.Errorf("verify: escape path %q: %w", mod, err)
	}
	escVer, err := module.EscapeVersion(ver)
	if err != nil {
		return "", fmt.Errorf("verify: escape version %q: %w", ver, err)
	}
	url := o.baseURL + "/lookup/" + escMod + "@" + escVer

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("verify: build request: %w", err)
	}
	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("verify: lookup %s: %w", url, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// fall through to parse
	case http.StatusNotFound, http.StatusGone:
		return "", fmt.Errorf("verify: %s@%s: %w", mod, ver, ErrNotInSumDB)
	default:
		return "", fmt.Errorf("verify: lookup %s: unexpected status %d", url, resp.StatusCode)
	}

	// Response lines: "<mod> <ver> h1:...=" and "<mod> <ver>/go.mod h1:...=".
	// We want the zip line (second field == the plain version).
	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) == 3 && f[0] == mod && f[1] == ver {
			return f[2], nil
		}
	}
	if err := sc.Err(); err != nil {
		return "", fmt.Errorf("verify: read lookup body: %w", err)
	}
	return "", fmt.Errorf("verify: %s@%s: %w", mod, ver, ErrNotInSumDB)
}
