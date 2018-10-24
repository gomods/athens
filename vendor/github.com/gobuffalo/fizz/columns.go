package fizz

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Deprecated: Fizz won't force you to have an ID field now.
var INT_ID_COL = Column{
	Name:    "id",
	Primary: true,
	ColType: "integer",
	Options: Options{},
}

// Deprecated: Fizz won't force you to have an ID field now.
var UUID_ID_COL = Column{
	Name:    "id",
	Primary: true,
	ColType: "uuid",
	Options: Options{},
}

var CREATED_COL = Column{Name: "created_at", ColType: "timestamp", Options: Options{}}
var UPDATED_COL = Column{Name: "updated_at", ColType: "timestamp", Options: Options{}}

type Column struct {
	Name    string
	ColType string
	Primary bool
	Options map[string]interface{}
}

func (c Column) String() string {
	if c.Primary || c.Options != nil {
		var opts map[string]interface{}
		if c.Options == nil {
			opts = make(map[string]interface{}, 0)
		} else {
			opts = c.Options
		}
		if c.Primary {
			opts["primary"] = true
		}

		o := make([]string, 0, len(opts))
		for k, v := range opts {
			vv, _ := json.Marshal(v)
			o = append(o, fmt.Sprintf("%s: %s", k, string(vv)))
		}
		sort.SliceStable(o, func(i, j int) bool { return o[i] < o[j] })
		return fmt.Sprintf(`t.Column("%s", "%s", {%s})`, c.Name, c.ColType, strings.Join(o, ", "))
	}
	return fmt.Sprintf(`t.Column("%s", "%s")`, c.Name, c.ColType)
}

func (f fizzer) ChangeColumn(table, name, ctype string, options Options) {
	t := Table{
		Name: table,
		Columns: []Column{
			{Name: name, ColType: ctype, Options: options},
		},
	}
	f.add(f.Bubbler.ChangeColumn(t))
}

func (f fizzer) AddColumn(table, name, ctype string, options Options) {
	t := Table{
		Name: table,
		Columns: []Column{
			{Name: name, ColType: ctype, Options: options},
		},
	}
	f.add(f.Bubbler.AddColumn(t))
}

func (f fizzer) DropColumn(table, name string) {
	t := Table{
		Name: table,
		Columns: []Column{
			{Name: name},
		},
	}
	f.add(f.Bubbler.DropColumn(t))
}

func (f fizzer) RenameColumn(table, old, new string) error {
	t := Table{
		Name: table,
		Columns: []Column{
			{Name: old},
			{Name: new},
		},
	}
	return f.add(f.Bubbler.RenameColumn(t))
}
