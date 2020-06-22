package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/index"
)

func New(cfg *config.MySQL) (index.Indexer, error) {
	dataSource := getMySQLSource(cfg)
	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}
	return &indexer{db}, nil
}

const schema = `
	CREATE TABLE IF NOT EXISTS indexes(
	id INT
		AUTO_INCREMENT
		PRIMARY KEY
		COMMENT 'Unique identifier for a module line',

	module VARCHAR(255)
		NOT NULL
		COMMENT 'Name of the module',

	version VARCHAR(255)
		NOT NULL
		COMMENT 'Name of the module',

	timestamp TIMESTAMP(6)
		COMMENT 'Date and time when the module was first created',

	INDEX (timestamp),
	UNIQUE INDEX module_version (module, version)
	) CHARACTER SET utf8;
`

type indexer struct {
	db *sql.DB
}

func (i *indexer) Index(ctx context.Context, mod, ver string) error {
	const op errors.Op = "mysql.Index"
	_, err := i.db.ExecContext(
		ctx,
		`INSERT INTO indexes (module, version, timestamp) VALUES (?, ?, ?)`,
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
	const op errors.Op = "mysql.Lines"
	if since.IsZero() {
		since = time.Unix(0, 0)
	}
	sinceStr := since.Format(time.RFC3339Nano)
	rows, err := i.db.QueryContext(ctx, `SELECT module, version, timestamp FROM indexes WHERE timestamp >= ? LIMIT ?`, sinceStr, limit)
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

func getMySQLSource(cfg *config.MySQL) string {
	c := mysql.NewConfig()
	c.Net = cfg.Protocol
	c.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	c.User = cfg.User
	c.Passwd = cfg.Password
	c.DBName = cfg.Database
	c.Params = cfg.Params
	return c.FormatDSN()
}
