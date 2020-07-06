package postgres

import (
	"os"
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/index/compliance"
	"github.com/gomods/athens/pkg/internal/testutil"
)

func TestPostgres(t *testing.T) {
	cfg := getTestConfig(t)
	i, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	compliance.RunTests(t, i, i.(*indexer).clear)
}

func (i *indexer) clear() error {
	_, err := i.db.Exec(`DELETE FROM indexes`)
	return err
}

func getTestConfig(t *testing.T) *config.Postgres {
	t.Helper()
	c, err := config.Load("")
	if err != nil {
		t.Fatal(err)
	}
	cfg := c.Index.Postgres
	if cfg.Password == "" {
		cfg.Password = "postgres"
	}
	if os.Getenv("STATIC_PORTS") == "" {
		cfg.Host, cfg.Port = testutil.GetServicePort(t, "postgres", 5432)
	}
	return cfg
}
