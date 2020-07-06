package mysql

import (
	"os"
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/index/compliance"
	"github.com/gomods/athens/pkg/internal/testutil"
)

func TestMySQL(t *testing.T) {
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

func getTestConfig(t *testing.T) *config.MySQL {
	t.Helper()
	if os.Getenv("SKIP_INDEX_MYSQL") != "" {
		t.SkipNow()
		return nil
	}
	c, err := config.Load("")
	if err != nil {
		t.Fatal(err)
	}
	cfg := c.Index.MySQL
	if os.Getenv("STATIC_PORTS") == "" {
		cfg.Host, cfg.Port = testutil.GetServicePort(t, "mysql", 3306)
	}
	return cfg
}
