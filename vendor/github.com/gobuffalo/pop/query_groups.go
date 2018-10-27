package pop

import "github.com/gobuffalo/pop/logging"

// GroupBy will append a GROUP BY clause to the query
func (q *Query) GroupBy(field string, fields ...string) *Query {
	if q.RawSQL.Fragment != "" {
		log(logging.Warn, "Query is setup to use raw SQL")
		return q
	}
	q.groupClauses = append(q.groupClauses, GroupClause{field})
	if len(fields) > 0 {
		for i := range fields {
			q.groupClauses = append(q.groupClauses, GroupClause{fields[i]})
		}
	}
	return q
}
