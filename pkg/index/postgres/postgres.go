package postgres

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	// register the driver with database/sql
	_ "github.com/lib/pq"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/index"
)

func New(cfg *config.Postgres) (index.Indexer, error) {
	dataSource := getPostgresSource(cfg)
	db, err := sql.Open("postgres", dataSource)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	for _, statement := range schema {
		_, err = db.Exec(statement)
		if err != nil {
			return nil, err
		}
	}
	return &indexer{db}, nil
}

var schema = [...]string{
	`
		CREATE TABLE IF NOT EXISTS indexes(
			id SERIAL PRIMARY KEY,
			module VARCHAR(255) NOT NULL,
			version VARCHAR(255) NOT NULL,
			timestamp timestamp NOT NULL
		)
	`,
	`
		CREATE INDEX IF NOT EXISTS idx_timestamp ON indexes (timestamp)
	`,
	`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_module_version ON indexes (module, version)
	`,
}

type indexer struct {
	db *sql.DB
}

func (i *indexer) Index(ctx context.Context, mod, ver string) error {
	const op errors.Op = "postgres.Index"
	_, err := i.db.ExecContext(
		ctx,
		`INSERT INTO indexes (module, version, timestamp) VALUES ($1, $2, $3)`,
		mod,
		ver,
		time.Now().Format(time.RFC3339Nano),
	)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (i *indexer) Lines(ctx context.Context, since time.Time, limit int) ([]*index.Line, error) {
	const op errors.Op = "postgres.Lines"
	if since.IsZero() {
		since = time.Unix(0, 0)
	}
	sinceStr := since.Format(time.RFC3339Nano)
	rows, err := i.db.QueryContext(ctx, `SELECT module, version, timestamp FROM indexes WHERE timestamp >= $1 LIMIT $2`, sinceStr, limit)
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer rows.Close()
	lines := []*index.Line{}
	for rows.Next() {
		var line index.Line
		err = rows.Scan(&line.Module, &line.Version, &line.Timestamp)
		if err != nil {
			return nil, errors.E(op, err)
		}
		lines = append(lines, &line)
	}
	return lines, nil
}

func getPostgresSource(cfg *config.Postgres) string {
	args := []string{}
	args = append(args, "host="+cfg.Host)
	args = append(args, "port=", strconv.Itoa(cfg.Port))
	args = append(args, "user=", cfg.User)
	args = append(args, "dbname=", cfg.Database)
	args = append(args, "password="+cfg.Password)
	for k, v := range cfg.Params {
		args = append(args, k+"="+v)
	}
	return strings.Join(args, " ")
}
