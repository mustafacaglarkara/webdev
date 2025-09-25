package q

import (
	"fmt"
	"strings"
)

type Op string

const (
	OpEq   Op = "="
	OpNe   Op = "!="
	OpGt   Op = ">"
	OpGte  Op = ">="
	OpLt   Op = "<"
	OpLte  Op = "<="
	OpIn   Op = "IN"
	OpLike Op = "LIKE"
)

type Q struct {
	Field string
	Op    Op
	Value any
	And   []*Q
	Or    []*Q
	Not   *Q
}

// Eq, Gt, Lt, ...
func Eq(field string, value any) *Q   { return &Q{Field: field, Op: OpEq, Value: value} }
func Ne(field string, value any) *Q   { return &Q{Field: field, Op: OpNe, Value: value} }
func Gt(field string, value any) *Q   { return &Q{Field: field, Op: OpGt, Value: value} }
func Gte(field string, value any) *Q  { return &Q{Field: field, Op: OpGte, Value: value} }
func Lt(field string, value any) *Q   { return &Q{Field: field, Op: OpLt, Value: value} }
func Lte(field string, value any) *Q  { return &Q{Field: field, Op: OpLte, Value: value} }
func In(field string, value any) *Q   { return &Q{Field: field, Op: OpIn, Value: value} }
func Like(field string, value any) *Q { return &Q{Field: field, Op: OpLike, Value: value} }

func And(qs ...*Q) *Q { return &Q{And: qs} }
func Or(qs ...*Q) *Q  { return &Q{Or: qs} }
func Not(q *Q) *Q     { return &Q{Not: q} }

// ToSQL: SQL string ve arg dizisi üretir.
func (q *Q) ToSQL() (string, []any) {
	if q == nil {
		return "", nil
	}
	var parts []string
	var args []any
	switch {
	case q.And != nil:
		for _, sub := range q.And {
			w, a := sub.ToSQL()
			if w != "" {
				parts = append(parts, fmt.Sprintf("(%s)", w))
				args = append(args, a...)
			}
		}
		return strings.Join(parts, " AND "), args
	case q.Or != nil:
		for _, sub := range q.Or {
			w, a := sub.ToSQL()
			if w != "" {
				parts = append(parts, fmt.Sprintf("(%s)", w))
				args = append(args, a...)
			}
		}
		return strings.Join(parts, " OR "), args
	case q.Not != nil:
		w, a := q.Not.ToSQL()
		return fmt.Sprintf("NOT (%s)", w), a
	default:
		if q.Op == OpIn {
			vals, ok := q.Value.([]any)
			if !ok {
				return fmt.Sprintf("%s IN (?)", q.Field), []any{q.Value}
			}
			placeholders := make([]string, len(vals))
			for i := range vals {
				placeholders[i] = "?"
			}
			return fmt.Sprintf("%s IN (%s)", q.Field, strings.Join(placeholders, ", ")), vals
		}
		return fmt.Sprintf("%s %s ?", q.Field, q.Op), []any{q.Value}
	}
}
