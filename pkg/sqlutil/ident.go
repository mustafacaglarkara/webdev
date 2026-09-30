package sqlutil

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Dialect: SQL diyalekti. Tanımlayıcı tırnaklama ve yer tutucu (placeholder)
// biçimi diyalekte göre değişir.
type Dialect string

const (
	// Generic: diyalekt bilinmiyor; tanımlayıcılar doğrulanır ama tırnaklanmaz,
	// yer tutucu olarak "?" kullanılır (GORM / database/sql + sqlite/mysql uyumlu).
	Generic   Dialect = ""
	Postgres  Dialect = "postgres"
	MySQL     Dialect = "mysql"
	SQLite    Dialect = "sqlite"
	SQLServer Dialect = "sqlserver"
)

// ErrInvalidIdentifier: geçersiz tablo/kolon/şema adı.
var ErrInvalidIdentifier = errors.New("sqlutil: geçersiz tanımlayıcı")

// ParseDialect: sürücü/diyalekt adını normalize eder ("pg", "postgresql",
// "mssql", "sqlite3" gibi takma adları kabul eder).
func ParseDialect(name string) (Dialect, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "postgres", "postgresql", "pg", "pgx":
		return Postgres, nil
	case "mysql", "mariadb":
		return MySQL, nil
	case "sqlite", "sqlite3":
		return SQLite, nil
	case "sqlserver", "mssql":
		return SQLServer, nil
	case "", "generic":
		return Generic, nil
	}
	return Generic, fmt.Errorf("sqlutil: bilinmeyen diyalekt: %q", name)
}

// DetectDialect: *sql.DB'nin sürücü tipinden diyalekti tahmin eder.
// Tanınmayan sürücüde Generic döner.
func DetectDialect(db *sql.DB) Dialect {
	if db == nil {
		return Generic
	}
	t := reflect.TypeOf(db.Driver())
	if t == nil {
		return Generic
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	full := strings.ToLower(t.PkgPath() + "." + t.Name())
	switch {
	case strings.Contains(full, "sqlite"):
		return SQLite
	case strings.Contains(full, "pgx"), strings.Contains(full, "lib/pq"), strings.Contains(full, "postgres"):
		return Postgres
	case strings.Contains(full, "mysql"):
		return MySQL
	case strings.Contains(full, "mssql"), strings.Contains(full, "sqlserver"):
		return SQLServer
	}
	return Generic
}

// identPart: tek bir tanımlayıcı parçası: harf/altçizgi ile başlar, harf/rakam/altçizgi devam eder.
var identRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,127}$`)

// ValidateIdent: tanımlayıcının katı kalıba uyduğunu doğrular. En fazla üç
// noktalı parça kabul edilir (şema.tablo.kolon). Tırnak, boşluk, yorum vb.
// içeren her şey reddedilir.
func ValidateIdent(name string) error {
	if name == "" {
		return fmt.Errorf("%w: boş", ErrInvalidIdentifier)
	}
	parts := strings.Split(name, ".")
	if len(parts) > 3 {
		return fmt.Errorf("%w: %q", ErrInvalidIdentifier, name)
	}
	for _, p := range parts {
		if !identRe.MatchString(p) {
			return fmt.Errorf("%w: %q", ErrInvalidIdentifier, name)
		}
	}
	return nil
}

// QuoteIdent: tanımlayıcıyı doğrular ve diyalekte göre tırnaklar.
//
//	postgres/sqlite: "schema"."table"
//	mysql:           `schema`.`table`
//	sqlserver:       [schema].[table]
//	Generic:         schema.table (yalnızca doğrulama)
func QuoteIdent(d Dialect, name string) (string, error) {
	if err := ValidateIdent(name); err != nil {
		return "", err
	}
	parts := strings.Split(name, ".")
	for i, p := range parts {
		parts[i] = quotePart(d, p)
	}
	return strings.Join(parts, "."), nil
}

// QuoteIdents: birden fazla tanımlayıcıyı tırnaklar.
func QuoteIdents(d Dialect, names []string) ([]string, error) {
	out := make([]string, len(names))
	for i, n := range names {
		q, err := QuoteIdent(d, n)
		if err != nil {
			return nil, err
		}
		out[i] = q
	}
	return out, nil
}

func quotePart(d Dialect, p string) string {
	switch d {
	case Postgres, SQLite:
		return `"` + p + `"`
	case MySQL:
		return "`" + p + "`"
	case SQLServer:
		return "[" + p + "]"
	}
	return p
}

// Placeholder: n. (1 tabanlı) parametre için diyalekte uygun yer tutucu döner:
// postgres "$n", sqlserver "@pn", diğerleri "?".
func Placeholder(d Dialect, n int) string {
	switch d {
	case Postgres:
		return "$" + strconv.Itoa(n)
	case SQLServer:
		return "@p" + strconv.Itoa(n)
	}
	return "?"
}

// sqlTypeRe: SQL Server OUT parametre tipi için katı kalıp: INT, NVARCHAR(100),
// NVARCHAR(MAX), DECIMAL(18, 2) ...
var sqlTypeRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,63}(\s*\(\s*(\d{1,5}|[Mm][Aa][Xx])\s*(,\s*\d{1,5}\s*)?\))?$`)

// ValidateSQLType: tip adının katı kalıba uyduğunu doğrular.
func ValidateSQLType(typ string) error {
	if !sqlTypeRe.MatchString(strings.TrimSpace(typ)) {
		return fmt.Errorf("sqlutil: geçersiz SQL tipi: %q", typ)
	}
	return nil
}
