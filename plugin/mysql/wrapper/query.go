package wrapper

import (
	"reflect"
	"strings"
)

type (
	Query struct {
		fragments []fragment
		args      []interface{}
		order     []string
		group     []string
	}

	fragment struct {
		and   bool
		query string
	}
)

func Q() *Query {
	return new(Query)
}

func (q *Query) expr(field string, val ...interface{}) *Query {
	q.fragments = append(q.fragments, fragment{and: true, query: field})
	q.args = append(q.args, val...)
	return q
}

func (q *Query) ExprIf(test bool, field string, val ...interface{}) *Query {
	if test {
		return q.Expr(field, val...)
	}
	return q
}

func (q *Query) Expr(field string, val ...interface{}) *Query {
	return q.expr(field, val...)
}

func (q *Query) LikeIf(test bool, field string, val interface{}) *Query {
	if test {
		return q.Like(field, val)
	}
	return q
}

func (q *Query) Like(field string, val interface{}) *Query {
	return q.Expr(field+" like ?", val)
}

func (q *Query) EqIf(test bool, field string, val interface{}) *Query {
	if test {
		return q.Eq(field, val)
	}
	return q
}

func (q *Query) Eq(field string, val interface{}) *Query {
	return q.Expr(field+" = ?", val)
}

func (q *Query) GtIf(test bool, field string, val interface{}) *Query {
	if test {
		return q.Gt(field, val)
	}
	return q
}

func (q *Query) Gt(field string, val interface{}) *Query {
	return q.Expr(field+" > ?", val)
}

func (q *Query) GteIf(test bool, field string, val interface{}) *Query {
	if test {
		return q.Gte(field, val)
	}
	return q
}

func (q *Query) Gte(field string, val interface{}) *Query {
	return q.Expr(field+" >= ?", val)
}

func (q *Query) LtIf(test bool, field string, val interface{}) *Query {
	if test {
		return q.Lt(field, val)
	}
	return q
}

func (q *Query) Lt(field string, val interface{}) *Query {
	return q.Expr(field+" < ?", val)
}

func (q *Query) LteIf(test bool, field string, val interface{}) *Query {
	if test {
		return q.Lte(field, val)
	}
	return q
}

func (q *Query) Lte(field string, val interface{}) *Query {
	return q.Expr(field+" <= ?", val)
}

func (q *Query) In(field string, val ...interface{}) *Query {
	if val == nil {
		return q
	}
	switch reflect.TypeOf(val).Kind() {
	case reflect.Array, reflect.Slice:
		s := reflect.ValueOf(val)
		if s.Len() == 0 {
			return q
		} else if s.Len() == 1 {
			v := s.Index(0)
			switch v.Elem().Kind() {
			case reflect.Array, reflect.Slice:
				if v.Elem().Len() == 0 {
					return q
				} else if v.Elem().Len() == 1 {
					return q.Eq(field, v.Elem().Index(0).Interface())
				}
				return q.Expr(field+" in ?", v.Interface())
			}
			return q.Eq(field, v.Interface())
		}
	}
	return q.Expr(field+" in ?", val)
}

func (q *Query) InIf(test bool, field string, val ...interface{}) *Query {
	if test {
		return q.In(field, val...)
	}
	return q
}

func (q *Query) InSql(field string, sql string, val ...interface{}) *Query {
	return q.expr(field+" in ("+sql+")", val...)
}

func (q *Query) InSqlIf(test bool, field string, sql string, val ...interface{}) *Query {
	if test {
		return q.InSql(field, sql, val...)
	}
	return q
}

func (q *Query) And(causes ...*Query) *Query {
	if len(causes) == 0 {
		return q
	}
	for _, cause := range causes {
		fragments := cause.Build()
		fragmentCause := fragments[0].(string)
		args := fragments[1].([]interface{})
		if fragmentCause == "" {
			return q
		}
		q.fragments = append(q.fragments, fragment{and: true, query: "( " + fragmentCause + " )"})
		q.args = append(q.args, args...)
	}
	return q
}

func (q *Query) Or(cause *Query) *Query {
	fragments := cause.Build()
	fragmentCause := fragments[0].(string)
	args := fragments[1].([]interface{})
	if fragmentCause == "" {
		return q
	}
	q.fragments = append(q.fragments, fragment{and: false, query: "( " + fragmentCause + " )"})
	q.args = append(q.args, args...)
	return q
}

func (q *Query) Asc(fields ...string) *Query {
	q.order = append(q.order, fields...)
	return q
}

func (q *Query) Desc(fields ...string) *Query {
	for _, field := range fields {
		q.order = append(q.order, field+" desc")
	}
	return q
}

func (q *Query) GroupBy(fields ...string) *Query {
	q.group = append(q.group, fields...)
	return q
}

func (q *Query) Build() []interface{} {
	sqlFragment, args, groupBy, orderBy := q.build()
	return []interface{}{sqlFragment, args, groupBy, orderBy}
}

func (q *Query) ForUpdate() *Update {
	return U(q)
}

func (q *Query) build() (string, []interface{}, []string, string) {
	orderStr := strings.Join(q.order, ", ")
	elems := q.fragments
	andSep := " and "
	orSep := " or "
	switch len(elems) {
	case 0:
		return "", q.args, q.group, orderStr
	case 1:
		return elems[0].query, q.args, q.group, orderStr
	}
	n := 0
	for i := 0; i < len(elems); i++ {
		n += len(elems[i].query)
		if elems[i].and {
			n += len(andSep)
		} else {
			n += len(orSep)
		}
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(elems[0].query)
	for _, s := range elems[1:] {
		if s.and {
			b.WriteString(andSep)
			b.WriteString(s.query)
		} else {
			b.WriteString(orSep)
			b.WriteString(s.query)
		}
	}
	return b.String(), q.args, q.group, orderStr
}
