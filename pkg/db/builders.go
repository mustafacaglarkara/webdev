package db

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/mustafacaglarkara/webdev/pkg/sqlutil"
)

// ----- ${name} bağlama -----

// emptyListRe: boş liste için "<ifade> [NOT] IN (" bağlamını yakalar.
// İfade basit (isteğe bağlı nitelikli / tırnaklı) bir kolon adı olmalıdır.
var emptyListRe = regexp.MustCompile("(?i)((?:[A-Za-z_][A-Za-z0-9_]*|\"[^\"]+\"|`[^`]+`|\\[[^\\]]+\\])(?:\\.(?:[A-Za-z_][A-Za-z0-9_]*|\"[^\"]+\"|`[^`]+`|\\[[^\\]]+\\]))*)\\s+(NOT\\s+)?IN\\s*\\(\\s*$")

var sqlKeywords = map[string]bool{"NOT": true, "AND": true, "OR": true, "WHERE": true, "ON": true, "WHEN": true, "THEN": true, "ELSE": true, "HAVING": true, "SELECT": true}

var notInTailRe = regexp.MustCompile(`(?i)\bNOT\s+IN\s*\(\s*$`)

// bindNamed: ${name} yer tutucularını ph(n) ile değiştirir ve argümanları üretir.
// Slice/array değerler virgüllü çoklu yer tutucuya açılır ([]byte tek değerdir).
// Boş liste:
//   - "kolon IN (${x})"     => "1=0"
//   - "kolon NOT IN (${x})" => "1=1"
//   - diğer IN bağlamları   => "NULL" (IN için doğru sonuç)
//   - karmaşık ifadeli NOT IN => hata (NOT IN (NULL) hiçbir satır döndürmez)
//
// Eksik parametre hata verir.
func bindNamed(sqlText string, params map[string]any, ph func(n int) string) (string, []any, error) {
	var out strings.Builder
	args := make([]any, 0, 8)
	next := func(v any) {
		args = append(args, v)
		out.WriteString(ph(len(args)))
	}
	for i := 0; i < len(sqlText); {
		if i+2 < len(sqlText) && sqlText[i] == '$' && sqlText[i+1] == '{' {
			j := strings.IndexByte(sqlText[i+2:], '}')
			if j < 0 {
				return "", nil, errors.New("parametre süslü parantez kapanmıyor")
			}
			j += i + 2
			name := sqlText[i+2 : j]
			val, ok := params[name]
			if !ok {
				return "", nil, fmt.Errorf("eksik parametre: %s", name)
			}
			i = j + 1
			vals, isList := expandList(val)
			if !isList {
				next(val)
				continue
			}
			if len(vals) == 0 {
				cur := out.String()
				rest := sqlText[i:]
				closeIdx := strings.IndexByte(rest, ')')
				m := emptyListRe.FindStringSubmatchIndex(cur)
				if m != nil && sqlKeywords[strings.ToUpper(cur[m[2]:m[3]])] {
					m = nil // "NOT IN (" içindeki NOT bir kolon adı değildir
				}
				if m != nil && closeIdx >= 0 && strings.TrimSpace(rest[:closeIdx]) == "" {
					not := m[4] >= 0
					out.Reset()
					out.WriteString(cur[:m[0]])
					if not {
						out.WriteString("1=1")
					} else {
						out.WriteString("1=0")
					}
					i += closeIdx + 1
					continue
				}
				if notInTailRe.MatchString(cur) {
					return "", nil, fmt.Errorf("boş liste NOT IN içinde kullanılamaz (parametre %s): basit bir kolon adı kullanın veya listeyi kontrol edin", name)
				}
				out.WriteString("NULL")
				continue
			}
			for k, v := range vals {
				if k > 0 {
					out.WriteString(",")
				}
				next(v)
			}
			continue
		}
		out.WriteByte(sqlText[i])
		i++
	}
	return out.String(), args, nil
}

func qmark(int) string { return "?" }

// bindNamedToQ: GORM için "?" yer tutucularıyla bağlar (GORM diyalekte çevirir).
func bindNamedToQ(sqlText string, params map[string]any) (string, []any, error) {
	return bindNamed(sqlText, params, qmark)
}

func expandList(v any) ([]any, bool) {
	if v == nil {
		return nil, false
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

// ----- diyalekt / batch -----

func dialectFromName(name string) sqlutil.Dialect {
	d, _ := sqlutil.ParseDialect(name)
	return d
}

// maxParamsFor: tek ifadede güvenli azami parametre sayısı.
func maxParamsFor(dialect string) int {
	switch dialect {
	case "sqlite":
		return 999
	case "postgres":
		return 65535
	case "sqlserver":
		return 2000 // gerçek sınır 2100; pay bırakılır
	case "mysql":
		return 65535
	default:
		return 999
	}
}

// defaultBatchSizeFor: satır başına paramsPerRow parametre için varsayılan batch boyutu.
func defaultBatchSizeFor(dialect string, paramsPerRow int) int {
	if paramsPerRow <= 0 {
		return 100
	}
	maxRows := maxParamsFor(dialect) / paramsPerRow
	capRows := 500
	switch dialect {
	case "sqlite":
		capRows = 300
	case "postgres":
		capRows = 2000
	case "mysql":
		switch {
		case paramsPerRow <= 8:
			capRows = 1000
		case paramsPerRow <= 25:
			capRows = 500
		default:
			capRows = 200
		}
	}
	if maxRows > capRows {
		maxRows = capRows
	}
	if maxRows <= 0 {
		maxRows = 1
	}
	return maxRows
}

// batchSizeFor: istenen batch boyutunu parametre sınırına göre kırpar.
func batchSizeFor(dialect string, paramsPerRow, requested int) (int, error) {
	limit := maxParamsFor(dialect)
	if paramsPerRow > limit {
		return 0, fmt.Errorf("satır başına %d parametre, %s sınırını (%d) aşıyor", paramsPerRow, dialect, limit)
	}
	bs := requested
	if bs <= 0 {
		bs = defaultBatchSizeFor(dialect, paramsPerRow)
	}
	if paramsPerRow > 0 && bs*paramsPerRow > limit {
		bs = limit / paramsPerRow
	}
	if bs <= 0 {
		bs = 1
	}
	return bs, nil
}

// ----- SQL builder'lar (tanımlayıcılar doğrulanır ve tırnaklanır) -----

func quoteAll(d sqlutil.Dialect, names []string) ([]string, error) {
	q, err := sqlutil.QuoteIdents(d, names)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func checkRows(cols []string, rows [][]any) error {
	for i, r := range rows {
		if len(r) != len(cols) {
			return fmt.Errorf("satır %d: %d değer var, %d kolon bekleniyor", i, len(r), len(cols))
		}
	}
	return nil
}

func buildInsertSQL(d sqlutil.Dialect, table string, cols []string, rows [][]any) (string, []any, error) {
	qt, err := sqlutil.QuoteIdent(d, table)
	if err != nil {
		return "", nil, err
	}
	qc, err := quoteAll(d, cols)
	if err != nil {
		return "", nil, err
	}
	if err := checkRows(cols, rows); err != nil {
		return "", nil, err
	}
	var b strings.Builder
	b.WriteString("INSERT INTO ")
	b.WriteString(qt)
	b.WriteString(" (")
	b.WriteString(strings.Join(qc, ","))
	b.WriteString(") VALUES ")
	args := make([]any, 0, len(rows)*len(cols))
	rowPH := "(" + strings.TrimSuffix(strings.Repeat("?,", len(cols)), ",") + ")"
	for i, r := range rows {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(rowPH)
		args = append(args, r...)
	}
	return b.String(), args, nil
}

func buildUpsertSuffix(d sqlutil.Dialect, cols, conflictCols, updateCols []string) (string, error) {
	qu, err := quoteAll(d, updateCols)
	if err != nil {
		return "", err
	}
	qk, err := quoteAll(d, conflictCols)
	if err != nil {
		return "", err
	}
	switch d {
	case sqlutil.Postgres, sqlutil.SQLite:
		if len(conflictCols) == 0 {
			return "", errors.New("postgres/sqlite için conflictCols gerekli")
		}
		s := "ON CONFLICT (" + strings.Join(qk, ",") + ")"
		if len(qu) == 0 {
			return s + " DO NOTHING", nil
		}
		parts := make([]string, len(qu))
		for i, c := range qu {
			parts[i] = c + " = EXCLUDED." + c
		}
		return s + " DO UPDATE SET " + strings.Join(parts, ","), nil
	case sqlutil.MySQL:
		if len(qu) == 0 {
			// değişiklik yapmayan güncelleme: çakışmada satırı olduğu gibi bırak
			first, err := sqlutil.QuoteIdent(d, cols[0])
			if err != nil {
				return "", err
			}
			return "ON DUPLICATE KEY UPDATE " + first + " = " + first, nil
		}
		parts := make([]string, len(qu))
		for i, c := range qu {
			parts[i] = c + " = VALUES(" + c + ")"
		}
		return "ON DUPLICATE KEY UPDATE " + strings.Join(parts, ","), nil
	case sqlutil.SQLServer:
		return "", errors.New("sqlserver için upsert desteklenmiyor (MERGE kullanın)")
	default:
		return "", fmt.Errorf("bilinmeyen dialector: %q", d)
	}
}

// buildBulkUpdateByKeySQL: UPDATE t SET c = CASE WHEN k = ? THEN ? ... ELSE c END WHERE k IN (...)
func buildBulkUpdateByKeySQL(d sqlutil.Dialect, table, keyCol string, updateCols []string, rows []map[string]any) (string, []any, error) {
	if len(rows) == 0 {
		return "", nil, nil
	}
	if len(updateCols) == 0 {
		return "", nil, errors.New("updateCols boş olamaz")
	}
	qt, err := sqlutil.QuoteIdent(d, table)
	if err != nil {
		return "", nil, err
	}
	qk, err := sqlutil.QuoteIdent(d, keyCol)
	if err != nil {
		return "", nil, err
	}
	qu, err := quoteAll(d, updateCols)
	if err != nil {
		return "", nil, err
	}
	var b strings.Builder
	b.WriteString("UPDATE ")
	b.WriteString(qt)
	b.WriteString(" SET ")
	args := make([]any, 0, len(rows)*(len(updateCols)*2+1))
	for i, c := range updateCols {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(qu[i])
		b.WriteString(" = CASE ")
		for ri, r := range rows {
			key, ok := r[keyCol]
			if !ok {
				return "", nil, fmt.Errorf("satır %d: key eksik: %s", ri, keyCol)
			}
			val, ok := r[c]
			if !ok {
				return "", nil, fmt.Errorf("satır %d: kolon eksik: %s", ri, c)
			}
			b.WriteString("WHEN ")
			b.WriteString(qk)
			b.WriteString(" = ? THEN ? ")
			args = append(args, key, val)
		}
		b.WriteString("ELSE ")
		b.WriteString(qu[i])
		b.WriteString(" END")
	}
	b.WriteString(" WHERE ")
	b.WriteString(qk)
	b.WriteString(" IN (")
	b.WriteString(strings.TrimSuffix(strings.Repeat("?,", len(rows)), ","))
	b.WriteString(")")
	for _, r := range rows {
		args = append(args, r[keyCol])
	}
	return b.String(), args, nil
}

// ----- SQL Server MERGE -----

// mergeHints: MERGE hedef tablosu için izinli tablo ipuçları.
var mergeHints = map[string]bool{
	"HOLDLOCK": true, "SERIALIZABLE": true, "UPDLOCK": true, "ROWLOCK": true, "PAGLOCK": true,
	"TABLOCK": true, "TABLOCKX": true, "XLOCK": true, "READCOMMITTED": true,
	"READCOMMITTEDLOCK": true, "REPEATABLEREAD": true,
}

// normalizeTableHint: "WITH (HOLDLOCK, ROWLOCK)", "HOLDLOCK" veya "(UPDLOCK)"
// biçimlerini kabul eder; yalnızca izinli ipuçlarından "WITH (...)" üretir.
func normalizeTableHint(h string) (string, error) {
	s := strings.ToUpper(strings.TrimSpace(h))
	if s == "" {
		return "", nil
	}
	if strings.HasPrefix(s, "WITH") {
		s = strings.TrimSpace(s[4:])
	}
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		s = s[1 : len(s)-1]
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if !mergeHints[p] {
			return "", fmt.Errorf("izin verilmeyen SQL Server tablo ipucu: %q", h)
		}
		out = append(out, p)
	}
	return "WITH (" + strings.Join(out, ", ") + ")", nil
}

var outputItemRe = regexp.MustCompile(`(?i)^(\$action|(INSERTED|DELETED)\.(\*|[A-Za-z_][A-Za-z0-9_]*))(?:\s+AS\s+([A-Za-z_][A-Za-z0-9_]*))?$`)

// normalizeOutputClause: ham OUTPUT dizesini ayrıştırır; yalnızca
// $action, INSERTED.kolon, DELETED.kolon (isteğe bağlı "AS takma_ad") öğelerini kabul eder.
func normalizeOutputClause(s string) (string, error) {
	t := strings.TrimSpace(s)
	if t == "" {
		return "", nil
	}
	if len(t) < 7 || !strings.EqualFold(t[:6], "OUTPUT") || (t[6] != ' ' && t[6] != '\t' && t[6] != '\n') {
		return "", fmt.Errorf("geçersiz OUTPUT ifadesi: %q", s)
	}
	items := strings.Split(t[7:], ",")
	out := make([]string, 0, len(items))
	for _, it := range items {
		m := outputItemRe.FindStringSubmatch(strings.TrimSpace(it))
		if m == nil {
			return "", fmt.Errorf("geçersiz OUTPUT öğesi: %q", strings.TrimSpace(it))
		}
		var r string
		if strings.EqualFold(m[1], "$action") {
			r = "$action"
		} else {
			col := m[3]
			if col != "*" {
				col = "[" + col + "]"
			}
			r = strings.ToUpper(m[2]) + "." + col
		}
		if m[4] != "" {
			r += " AS [" + m[4] + "]"
		}
		out = append(out, r)
	}
	return "OUTPUT " + strings.Join(out, ", "), nil
}

func (o *UpsertOutput) render() (string, error) {
	if o == nil {
		return "", nil
	}
	parts := make([]string, 0, 1+len(o.InsertedCols)+len(o.DeletedCols))
	if o.IncludeAction {
		parts = append(parts, "$action AS [action]")
	}
	for _, c := range o.InsertedCols {
		q, err := sqlutil.QuoteIdent(sqlutil.SQLServer, c)
		if err != nil {
			return "", err
		}
		parts = append(parts, "INSERTED."+q)
	}
	for _, c := range o.DeletedCols {
		q, err := sqlutil.QuoteIdent(sqlutil.SQLServer, c)
		if err != nil {
			return "", err
		}
		parts = append(parts, "DELETED."+q)
	}
	if len(parts) == 0 {
		return "", nil
	}
	return "OUTPUT " + strings.Join(parts, ", "), nil
}

// buildMergeSQLServerWithOptions: MERGE tabanlı upsert (sqlserver) + opsiyonlar.
func buildMergeSQLServerWithOptions(table string, cols []string, keyCols []string, updateCols []string, rows [][]any, opts *UpsertOptions) (string, []any, error) {
	if len(cols) == 0 || len(rows) == 0 {
		return "", nil, nil
	}
	if len(keyCols) == 0 {
		return "", nil, errors.New("sqlserver MERGE için keyCols gerekli")
	}
	const d = sqlutil.SQLServer
	qt, err := sqlutil.QuoteIdent(d, table)
	if err != nil {
		return "", nil, err
	}
	qc, err := quoteAll(d, cols)
	if err != nil {
		return "", nil, err
	}
	qk, err := quoteAll(d, keyCols)
	if err != nil {
		return "", nil, err
	}
	qu, err := quoteAll(d, updateCols)
	if err != nil {
		return "", nil, err
	}
	if err := checkRows(cols, rows); err != nil {
		return "", nil, err
	}
	hint, output := "", ""
	if opts != nil {
		if hint, err = normalizeTableHint(opts.SQLServerTableHint); err != nil {
			return "", nil, err
		}
		if opts.Output != nil {
			if output, err = opts.Output.render(); err != nil {
				return "", nil, err
			}
		} else if output, err = normalizeOutputClause(opts.SQLServerOutput); err != nil {
			return "", nil, err
		}
	}

	var b strings.Builder
	b.WriteString("MERGE INTO ")
	b.WriteString(qt)
	if hint != "" {
		b.WriteString(" ")
		b.WriteString(hint)
	}
	b.WriteString(" AS target USING (VALUES ")
	args := make([]any, 0, len(rows)*len(cols))
	rowPH := "(" + strings.TrimSuffix(strings.Repeat("?,", len(cols)), ",") + ")"
	for i, r := range rows {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(rowPH)
		args = append(args, r...)
	}
	b.WriteString(") AS src (")
	b.WriteString(strings.Join(qc, ","))
	b.WriteString(") ON (")
	for i, k := range qk {
		if i > 0 {
			b.WriteString(" AND ")
		}
		b.WriteString("target." + k + " = src." + k)
	}
	b.WriteString(") ")
	if len(qu) > 0 {
		b.WriteString("WHEN MATCHED THEN UPDATE SET ")
		for i, c := range qu {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString("target." + c + " = src." + c)
		}
		b.WriteString(" ")
	}
	b.WriteString("WHEN NOT MATCHED THEN INSERT (")
	b.WriteString(strings.Join(qc, ","))
	b.WriteString(") VALUES (")
	for i, c := range qc {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString("src." + c)
	}
	b.WriteString(") ")
	if output != "" {
		b.WriteString(output)
		b.WriteString(" ")
	}
	b.WriteString(";")
	return b.String(), args, nil
}

// buildMergeSQLServer: WithOptions sarmalayıcısı (opts=nil)
func buildMergeSQLServer(table string, cols []string, keyCols []string, updateCols []string, rows [][]any) (string, []any, error) {
	return buildMergeSQLServerWithOptions(table, cols, keyCols, updateCols, rows, nil)
}
