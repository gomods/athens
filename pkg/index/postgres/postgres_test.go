package postgres

import (
	"os"
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/index/compliance"
)

func TestPostgres(t *testing.T) {
	if os.Getenv("TEST_INDEX_POSTGRES") != "true" {
		t.SkipNow()
	}
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
	cfg, err := config.Load("")
	if err != nil {
		t.Fatal(err)
	}
	return cfg.Index.Postgres
}
