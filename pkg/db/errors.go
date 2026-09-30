package db

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"net"
	"strings"
	"syscall"

	"github.com/go-sql-driver/mysql"
	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"

	"github.com/mustafacaglarkara/webdev/pkg/resilience"
)

// Hata sınıflandırması. Sürücüye özgü kodlar:
//   - PostgreSQL (pgx): SQLSTATE (40001 serialization, 40P01 deadlock, 08xxx bağlantı, 23xxx kısıt)
//   - SQL Server: hata numarası (1205 deadlock, 2627/2601 unique, 547 FK ...)
//   - MySQL: hata numarası (1213 deadlock, 1205 lock wait, 1062 duplicate ...)
//   - SQLite: SQLITE_BUSY/SQLITE_LOCKED geçici, SQLITE_CONSTRAINT kısıt

type sqlStater interface{ SQLState() string }
type mssqlNumberer interface{ SQLErrorNumber() int32 }
type codeInt interface{ Code() int } // modernc.org/sqlite

// IsTransient: hatanın yeniden denemeyle düzelebilecek geçici bir hata olup
// olmadığını söyler: bağlantı kopması/reddi, deadlock, serialization hatası,
// SQLite busy/locked. Context hataları, sql.ErrNoRows ve kısıt ihlalleri
// asla geçici değildir. Zaman aşımları burada geçici SAYILMAZ (bkz. IsTimeout).
func IsTransient(err error) bool {
	if err == nil || resilience.IsContextError(err) || errors.Is(err, sql.ErrNoRows) ||
		errors.Is(err, gorm.ErrRecordNotFound) || IsConstraintViolation(err) {
		return false
	}
	if errors.Is(err, driver.ErrBadConn) || errors.Is(err, mysql.ErrInvalidConn) ||
		errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNABORTED) || errors.Is(err, syscall.EPIPE) ||
		errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	var st sqlStater
	if errors.As(err, &st) {
		code := st.SQLState()
		switch {
		case code == "40001", code == "40P01", strings.HasPrefix(code, "08"),
			code == "57P01", code == "57P02", code == "57P03", code == "53300":
			return true
		}
		return false
	}
	var ms mssqlNumberer
	if errors.As(err, &ms) {
		switch ms.SQLErrorNumber() {
		case 1205, 1222, 233, 64, 10053, 10054, 10928, 10929, 40197, 40501, 40613, 49918, 49919, 49920:
			return true
		}
		return false
	}
	var my *mysql.MySQLError
	if errors.As(err, &my) {
		switch my.Number {
		case 1213, 1205, 1040, 1053:
			return true
		}
		return false
	}
	var sq sqlite3.Error
	if errors.As(err, &sq) {
		return sq.Code == sqlite3.ErrBusy || sq.Code == sqlite3.ErrLocked
	}
	var mc codeInt
	if errors.As(err, &mc) {
		c := mc.Code() & 0xff
		return c == 5 || c == 6
	}
	msg := strings.ToLower(err.Error())
	for _, s := range []string{"database is locked", "database table is locked", "deadlock", "connection reset by peer", "broken pipe", "connection refused", "bad connection"} {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}

// IsTimeout: sorgu/ağ zaman aşımı (context.DeadlineExceeded HARİÇ; o asla
// yeniden denenmez). Yalnızca okuma işlemlerinde yeniden denenir.
func IsTimeout(err error) bool {
	if err == nil || resilience.IsContextError(err) {
		return false
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return true
	}
	var st sqlStater
	if errors.As(err, &st) && st.SQLState() == "57014" {
		return true
	}
	var ms mssqlNumberer
	if errors.As(err, &ms) && ms.SQLErrorNumber() == -2 {
		return true
	}
	return false
}

// IsConstraintViolation: unique/foreign key/not null/check kısıt ihlali.
func IsConstraintViolation(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) || errors.Is(err, gorm.ErrForeignKeyViolated) || errors.Is(err, gorm.ErrCheckConstraintViolated) {
		return true
	}
	var st sqlStater
	if errors.As(err, &st) {
		return strings.HasPrefix(st.SQLState(), "23")
	}
	var ms mssqlNumberer
	if errors.As(err, &ms) {
		switch ms.SQLErrorNumber() {
		case 2627, 2601, 547, 515, 2628:
			return true
		}
		return false
	}
	var my *mysql.MySQLError
	if errors.As(err, &my) {
		switch my.Number {
		case 1062, 1451, 1452, 1048, 1216, 1217, 3819, 1364:
			return true
		}
		return false
	}
	var sq sqlite3.Error
	if errors.As(err, &sq) {
		return sq.Code == sqlite3.ErrConstraint
	}
	var mc codeInt
	if errors.As(err, &mc) {
		return mc.Code()&0xff == 19
	}
	return false
}

// shouldRetry: okumalarda geçici hatalar + zaman aşımları, yazmalarda yalnızca geçici hatalar.
func shouldRetry(read bool) func(error) bool {
	return func(err error) bool {
		if IsTransient(err) {
			return true
		}
		return read && IsTimeout(err)
	}
}

// isBreakerFailure: devre kesicinin başarısızlık saydığı hatalar. Kayıt
// bulunamadı, kısıt ihlali ve context hataları sayılmaz.
func isBreakerFailure(err error) bool {
	if err == nil || resilience.IsContextError(err) || errors.Is(err, sql.ErrNoRows) ||
		errors.Is(err, gorm.ErrRecordNotFound) || IsConstraintViolation(err) {
		return false
	}
	return true
}
