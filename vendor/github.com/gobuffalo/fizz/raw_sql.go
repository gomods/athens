package fizz

import (
	"strings"
)

func (f fizzer) RawSQL(sql string) {
	if !strings.HasSuffix(sql, ";") {
		sql += ";"
	}
	f.add(sql, nil)
}

// Deprecated: use RawSQL instead.
func (f fizzer) RawSql(sql string) {
	f.RawSQL(sql)
}
