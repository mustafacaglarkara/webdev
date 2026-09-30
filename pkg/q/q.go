// Package q, Django Q benzeri, iç içe kullanılabilen WHERE koşulu oluşturucusudur.
//
// Alan adları katı bir kalıpla doğrulanır (harf/rakam/altçizgi, isteğe bağlı
// "tablo.kolon"), operatörler izinli listeden gelir ve değerler her zaman yer
// tutucu olarak bağlanır. Geçersiz alan adı veya operatör SQL'e ulaşmaz; hata döner.
package q

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mustafacaglarkara/webdev/pkg/sqlutil"
)

type Op string

const (
	OpEq        Op = "="
	OpNe        Op = "!="
	OpGt        Op = ">"
	OpGte       Op = ">="
	OpLt        Op = "<"
	OpLte       Op = "<="
	OpIn        Op = "IN"
	OpNotIn     Op = "NOT IN"
	OpLike      Op = "LIKE"
	OpNotLike   Op = "NOT LIKE"
	OpIsNull    Op = "IS NULL"
	OpIsNotNull Op = "IS NOT NULL"
)

// allowedOps: izinli operatörler (büyük/küçük harf duyarsız eşleştirilir).
var allowedOps = map[string]Op{
	"=": OpEq, "!=": OpNe, "<>": OpNe, ">": OpGt, ">=": OpGte, "<": OpLt, "<=": OpLte,
	"IN": OpIn, "NOT IN": OpNotIn, "LIKE": OpLike, "NOT LIKE": OpNotLike,
	"IS NULL": OpIsNull, "IS NOT NULL": OpIsNotNull,
}

// ParseOp: operatörü izinli listeye göre doğrular.
func ParseOp(s string) (Op, error) {
	norm := strings.ToUpper(strings.Join(strings.Fields(s), " "))
	if op, ok := allowedOps[norm]; ok {
		return op, nil
	}
	return "", fmt.Errorf("q: izin verilmeyen operatör: %q", s)
}

type kind uint8

const (
	kindLeaf kind = iota
	kindAnd
	kindOr
	kindNot
)

// Q: tek bir koşul (Field/Op/Value) veya And/Or/Not grubu.
type Q struct {
	Field string
	Op    Op
	Value any
	And   []*Q
	Or    []*Q
	Not   *Q

	kind kind // constructor ile belirlenir; boş And()/Or() için gereklidir
}

func Eq(field string, value any) *Q      { return &Q{Field: field, Op: OpEq, Value: value} }
func Ne(field string, value any) *Q      { return &Q{Field: field, Op: OpNe, Value: value} }
func Gt(field string, value any) *Q      { return &Q{Field: field, Op: OpGt, Value: value} }
func Gte(field string, value any) *Q     { return &Q{Field: field, Op: OpGte, Value: value} }
func Lt(field string, value any) *Q      { return &Q{Field: field, Op: OpLt, Value: value} }
func Lte(field string, value any) *Q     { return &Q{Field: field, Op: OpLte, Value: value} }
func Like(field string, value any) *Q    { return &Q{Field: field, Op: OpLike, Value: value} }
func NotLike(field string, value any) *Q { return &Q{Field: field, Op: OpNotLike, Value: value} }
func IsNull(field string) *Q             { return &Q{Field: field, Op: OpIsNull} }
func IsNotNull(field string) *Q          { return &Q{Field: field, Op: OpIsNotNull} }

// In: value herhangi bir slice/array olabilir ([]int, []string, []any ...).
// Boş liste "1=0" üretir. Liste olmayan değer tek elemanlı liste sayılır.
func In(field string, value any) *Q { return &Q{Field: field, Op: OpIn, Value: value} }

// NotIn: boş liste "1=1" üretir.
func NotIn(field string, value any) *Q { return &Q{Field: field, Op: OpNotIn, Value: value} }

// Where: operatörü dizgi olarak alır (ör. kullanıcı arayüzünden); izinli listede
// yoksa ToSQL hata döner.
func Where(field, op string, value any) *Q {
	o, err := ParseOp(op)
	if err != nil {
		o = Op(op) // ToSQL'de reddedilir
	}
	return &Q{Field: field, Op: o, Value: value}
}

// And: boş And() "1=1" üretir.
func And(qs ...*Q) *Q { return &Q{And: qs, kind: kindAnd} }

// Or: boş Or() "1=0" üretir.
func Or(qs ...*Q) *Q { return &Q{Or: qs, kind: kindOr} }

// Not: Not(nil) "NOT (1=1)" üretir.
func Not(q *Q) *Q { return &Q{Not: q, kind: kindNot} }

func (q *Q) resolvedKind() kind {
	switch {
	case q.kind != kindLeaf:
		return q.kind
	case q.And != nil:
		return kindAnd
	case q.Or != nil:
		return kindOr
	case q.Not != nil:
		return kindNot
	}
	return kindLeaf
}

// ToSQL: tırnaksız (doğrulanmış) alan adları ve "?" yer tutucularla WHERE
// ifadesi ve argümanları üretir. nil Q "1=1" döner.
func (q *Q) ToSQL() (string, []any, error) {
	return q.Build(sqlutil.Generic)
}

// Build: diyalekte göre tırnaklanmış alan adları ve yer tutucularla ($1, @p1, ?)
// WHERE ifadesi üretir.
func (q *Q) Build(d sqlutil.Dialect) (string, []any, error) {
	b := &builder{d: d}
	s, err := b.build(q)
	if err != nil {
		return "", nil, err
	}
	return s, b.args, nil
}

type builder struct {
	d    sqlutil.Dialect
	args []any
}

func (b *builder) ph(v any) string {
	b.args = append(b.args, v)
	return sqlutil.Placeholder(b.d, len(b.args))
}

func (b *builder) build(q *Q) (string, error) {
	if q == nil {
		return "1=1", nil
	}
	switch q.resolvedKind() {
	case kindAnd, kindOr:
		subs, sep, empty := q.And, " AND ", "1=1"
		if q.resolvedKind() == kindOr {
			subs, sep, empty = q.Or, " OR ", "1=0"
		}
		parts := make([]string, 0, len(subs))
		for _, sub := range subs {
			if sub == nil {
				continue
			}
			w, err := b.build(sub)
			if err != nil {
				return "", err
			}
			parts = append(parts, "("+w+")")
		}
		if len(parts) == 0 {
			return empty, nil
		}
		return strings.Join(parts, sep), nil
	case kindNot:
		w, err := b.build(q.Not)
		if err != nil {
			return "", err
		}
		return "NOT (" + w + ")", nil
	}
	return b.leaf(q)
}

func (b *builder) leaf(q *Q) (string, error) {
	field, err := sqlutil.QuoteIdent(b.d, q.Field)
	if err != nil {
		return "", fmt.Errorf("q: alan adı: %w", err)
	}
	op, err := ParseOp(string(q.Op))
	if err != nil {
		return "", err
	}
	switch op {
	case OpIsNull, OpIsNotNull:
		return field + " " + string(op), nil
	case OpIn, OpNotIn:
		vals, ok := expand(q.Value)
		if !ok {
			vals = []any{q.Value}
		}
		if len(vals) == 0 {
			if op == OpIn {
				return "1=0", nil
			}
			return "1=1", nil
		}
		ps := make([]string, len(vals))
		for i, v := range vals {
			ps[i] = b.ph(v)
		}
		return field + " " + string(op) + " (" + strings.Join(ps, ", ") + ")", nil
	case OpEq, OpNe:
		if q.Value == nil {
			if op == OpEq {
				return field + " IS NULL", nil
			}
			return field + " IS NOT NULL", nil
		}
	}
	return field + " " + string(op) + " " + b.ph(q.Value), nil
}

// expand: slice/array değerini []any'e açar; []byte tek değer sayılır.
func expand(v any) ([]any, bool) {
	if v == nil {
		return nil, true
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, false
	}
	if rv.Kind() == reflect.Slice && rv.Type().Elem().Kind() == reflect.Uint8 {
		return nil, false
	}
	out := make([]any, rv.Len())
	for i := range out {
		out[i] = rv.Index(i).Interface()
	}
	return out, true
}
