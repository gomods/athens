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

// New returns a new Indexer with a MySQL implementation.
// It attempts to connect to the DB and create the index table
// if it doesn ot already exist.
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

	path VARCHAR(255)
		NOT NULL
		COMMENT 'Import path of the module',

	version VARCHAR(255)
		NOT NULL
		COMMENT 'Module version',

	timestamp TIMESTAMP(6)
		COMMENT 'Date and time when the module was first created',

	INDEX (timestamp),
	UNIQUE INDEX idx_module_version (path, version)
	) CHARACTER SET utf8;
`

type indexer struct {
	db *sql.DB
}

func (i *indexer) Index(ctx context.Context, mod, ver string) error {
	const op errors.Op = "mysql.Index"
	_, err := i.db.ExecContext(
		ctx,
		`INSERT INTO indexes (path, version, timestamp) VALUES (?, ?, ?)`,
		mod,
		ver,
		time.Now().Format(time.RFC3339Nano),
	)
	if err != nil {
		return errors.E(op, err, getKind(err))
	}
	return nil
}

func (i *indexer) Lines(ctx context.Context, since time.Time, limit int) ([]*index.Line, error) {
	const op errors.Op = "mysql.Lines"
	if since.IsZero() {
		since = time.Unix(0, 0)
	}
	sinceStr := since.Format(time.RFC3339Nano)
	rows, err := i.db.QueryContext(ctx, `SELECT path, version, timestamp FROM indexes WHERE timestamp >= ? LIMIT ?`, sinceStr, limit)
	if err != nil {
		return nil, errors.E(op, err)
	}
	defer rows.Close()
	lines := []*index.Line{}
	for rows.Next() {
		var line index.Line
		err = rows.Scan(&line.Path, &line.Version, &line.Timestamp)
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

func getKind(err error) int {
	mysqlErr, ok := err.(*mysql.MySQLError)
	if !ok {
		return errors.KindUnexpected
	}
	switch mysqlErr.Number {
	case 1062:
		return errors.KindAlreadyExists
	}
	return errors.KindUnexpected
}
