# migrate

`database/sql` üzerinde çalışan, sürüm takipli SQL migration koşucusu. İşletim sistemi
dizini veya `fs.FS` (ör. `embed.FS`) kaynaklarını destekler.

```go
import "github.com/mustafacaglarkara/webdev/pkg/migrate"
```

## Dosya adlandırma kuralları

```
<sürüm>_<ad>.up.sql     ileri yönlü (zorunlu)
<sürüm>_<ad>.down.sql   geri alma (isteğe bağlı; yoksa o sürüm geri alınamaz)
```

- `<sürüm>` yalnızca rakamdır ve **sayısal** olarak sıralanır (`2_x` < `10_y`).
  Sıfır dolgulu (`0001_...`) veya zaman damgalı (`20240101120000_...`) biçim önerilir.
  `0001` ile `1` aynı sürümdür; aynı sürüm için iki farklı ad hata verir.
- `<ad>`: harf, rakam, `_`, `-`, `.`.
- `.up.sql` / `.down.sql` ile biten ama kurala uymayan dosya `ErrInvalidName` hatası verir.
  Diğer `.sql` dosyaları yok sayılır (ör. `notes.sql`); `WithPlainSQL(true)` ile
  `<sürüm>_<ad>.sql` dosyaları da ileri yönlü migration sayılır ve kurala uymayan her `.sql` hata verir.
- Alt dizinler taranmaz.

Örnek:

```
migrations/
  0001_create_users.up.sql
  0001_create_users.down.sql
  0002_add_email_index.up.sql
  0002_add_email_index.down.sql
```

## Davranış

- Uygulanan sürümler `schema_migrations (version, name, applied_at)` tablosunda tutulur
  (tablo yoksa oluşturulur; ad `WithTable` ile değiştirilebilir).
- `Up` yalnızca **bekleyen** dosyaları artan sırada çalıştırır. Tekrar çalıştırmak güvenlidir.
- Her dosya **kendi transaction'ında** çalışır; sürüm kaydı aynı transaction'da yazılır.
  Hata olursa o dosya geri alınır, kaydedilmez ve sonraki dosyalar çalışmaz.
- Geri alma **azalan** sürüm sırasıyla, **yalnızca uygulanmış** sürümler için yapılır;
  varsayılan olarak **tek adım**. `.down.sql` yoksa `ErrNoDownFile` döner ve kayıt silinmez.
- Transaction içinde çalışamayan dosyalar (ör. PostgreSQL `CREATE INDEX CONCURRENTLY`,
  SQLite `VACUUM`) için dosyanın ilk satırına `-- migrate:no-transaction` yazın.
- MySQL'de DDL örtük commit yaptığı için DDL içeren dosyalar atomik değildir. MySQL'de bir
  dosyada birden fazla ifade için DSN'de `multiStatements=true` gerekir.
- Diyalekt sürücüden tespit edilir (`sqlutil.DetectDialect`); `WithDialect` ile verilebilir.
- **Eşzamanlı koşucular:** `Up`/`Down` çalışırken veritabanı advisory kilidi tutulur
  (PostgreSQL `pg_advisory_lock`, MySQL `GET_LOCK`, SQL Server `sp_getapplock`); ikinci
  koşucu bekler, sonra bekleyen dosya kalmadığını görüp döner. Kilit ayrı bir bağlantıda
  oturum düzeyinde tutulur. SQLite ve tanınmayan diyalektlerde kilit yoktur.
  `WithLock(false)` kapatır; `WithLockTimeout(d)` bekleme süresini sınırlar (aşılırsa
  `ErrLockTimeout`). Varsayılan: context iptal edilene kadar bekle.

## Paket fonksiyonları (işletim sistemi dizini)

```go
db, _ := sql.Open("sqlite3", "app.db")

err := migrate.Migrate(db, "migrations")            // bekleyenleri uygula
err = migrate.RunMigrations(db, "migrations")        // Migrate ile aynı
err = migrate.RollbackMigrations(db, "migrations")   // SON TEK migration'ı geri al
err = migrate.RollbackSteps(db, "migrations", 3)     // son 3'ü geri al
err = migrate.RollbackAll(db, "migrations")          // hepsini geri al
st, err := migrate.StatusDir(db, "migrations")       // []migrate.Status

//go:embed migrations
var migFS embed.FS
err = migrate.MigrateFS(db, migFS, "migrations")
```

> Not: `RollbackMigrations` eskiden **tüm** `.down.sql` dosyalarını çalıştırıyordu;
> artık yalnızca son uygulanan migration'ı geri alır.

## Migrator

```go
m := migrate.New(db, migFS, "migrations",
	migrate.WithDialect(sqlutil.Postgres),
	migrate.WithTable("schema_migrations"),
	migrate.WithLockTimeout(2*time.Minute), // advisory kilit için azami bekleme (0: sınırsız)
)
ctx := context.Background()

applied, err := m.Up(ctx)          // []migrate.Migration
rolled, err := m.Down(ctx, 1)      // azalan sırada 1 adım
rolled, err = m.DownAll(ctx)
status, err := m.Status(ctx)
for _, s := range status {
	fmt.Println(s.Version, s.Name, s.Applied, s.AppliedAt, s.Missing)
}
```

`migrate.NewDir(db, "migrations")` işletim sistemi dizini için kısayoldur.

## Güvenlik

- Migration dosyaları güvenilir SQL'dir ve olduğu gibi çalışır; kullanıcı girdisinden üretmeyin.
- Tablo adı (`WithTable`) tanımlayıcı olarak doğrulanır ve tırnaklanır.
- `fs.FS` yolları `fs.ValidPath` ile doğrulanır (`..`, mutlak yol reddedilir).
