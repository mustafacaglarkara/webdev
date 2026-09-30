package policy

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	stringadapter "github.com/casbin/casbin/v2/persist/string-adapter"
)

// ModelLoader casbin modelini üretir; Reload her çağrıldığında yeniden çalıştırılır.
// Böylece gömülü dosya (embed.FS) veya metin tabanlı modeller de yeniden yüklenebilir (A5-3).
type ModelLoader func() (model.Model, error)

// ModelFromText metin (model.conf içeriği) tabanlı ModelLoader döner.
func ModelFromText(text string) ModelLoader {
	return func() (model.Model, error) {
		if strings.TrimSpace(text) == "" {
			return nil, errors.New("policy: empty model text")
		}
		return model.NewModelFromString(text)
	}
}

// ModelFromFS fsys içindeki path dosyasını her yüklemede yeniden okuyan ModelLoader döner.
func ModelFromFS(fsys fs.FS, path string) ModelLoader {
	return func() (model.Model, error) {
		if fsys == nil {
			return nil, errors.New("policy: nil fs")
		}
		if !fs.ValidPath(path) {
			return nil, fmt.Errorf("policy: invalid model path %q", path)
		}
		b, err := fs.ReadFile(fsys, path)
		if err != nil {
			return nil, fmt.Errorf("policy: read model: %w", err)
		}
		return model.NewModelFromString(string(b))
	}
}

// NewWithModel bir ModelLoader ve isteğe bağlı adaptör ile enforcer başlatır. a nil ise
// politikalar yalnızca bellekte tutulur (AddPolicy çalışır, kalıcı olmaz). Reload modeli
// yükleyiciden, politikayı adaptörden yeniler.
func NewWithModel(load ModelLoader, a persist.Adapter) (*Manager, error) {
	if load == nil {
		return nil, errors.New("policy: nil model loader")
	}
	mdl, err := load()
	if err != nil {
		return nil, err
	}
	var e *casbin.SyncedEnforcer
	if a == nil {
		e, err = casbin.NewSyncedEnforcer(mdl)
	} else {
		e, err = casbin.NewSyncedEnforcer(mdl, a)
	}
	if err != nil {
		return nil, err
	}
	m := newManager(e)
	m.loadModel = load
	return m, nil
}

// NewFromText model ve politika (CSV) metinleriyle enforcer başlatır. policyCSV boşsa
// politikalar bellekte başlar. Gömülü/üretilmiş yapılandırmalar için uygundur.
func NewFromText(modelText, policyCSV string) (*Manager, error) {
	var a persist.Adapter
	if strings.TrimSpace(policyCSV) != "" {
		a = stringadapter.NewAdapter(policyCSV)
	}
	return NewWithModel(ModelFromText(modelText), a)
}

// NewFS fsys (ör. embed.FS) içindeki model ve politika dosyalarıyla enforcer başlatır.
// policyPath boşsa politikalar bellekte başlar. Reload her iki dosyayı da yeniden okur.
//
//	//go:embed policy/model.conf policy/policy.csv
//	var policyFS embed.FS
//	m, err := policy.NewFS(policyFS, "policy/model.conf", "policy/policy.csv")
func NewFS(fsys fs.FS, modelPath, policyPath string) (*Manager, error) {
	var a persist.Adapter
	if policyPath != "" {
		a = &fsAdapter{fsys: fsys, path: policyPath}
	}
	return NewWithModel(ModelFromFS(fsys, modelPath), a)
}

// NewFSWithAdapter fsys içindeki model dosyası ve bir persist.Adapter (ör. gormpolicy)
// ile enforcer başlatır.
func NewFSWithAdapter(fsys fs.FS, modelPath string, a persist.Adapter) (*Manager, error) {
	if a == nil {
		return nil, errors.New("policy: nil adapter")
	}
	return NewWithModel(ModelFromFS(fsys, modelPath), a)
}

// fsAdapter salt okunur fs.FS politika adaptörüdür (CSV). Yazma işlemleri casbin'in
// yok saydığı "not implemented" hatasını döner; AddPolicy/RemovePolicy bellekte çalışır.
type fsAdapter struct {
	fsys fs.FS
	path string
}

var errNotImplemented = errors.New("not implemented")

func (a *fsAdapter) LoadPolicy(m model.Model) error {
	if a.fsys == nil {
		return errors.New("policy: nil fs")
	}
	if !fs.ValidPath(a.path) {
		return fmt.Errorf("policy: invalid policy path %q", a.path)
	}
	f, err := a.fsys.Open(a.path)
	if err != nil {
		return fmt.Errorf("policy: read policy: %w", err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for sc.Scan() {
		if err := persist.LoadPolicyLine(strings.TrimSpace(sc.Text()), m); err != nil {
			return err
		}
	}
	return sc.Err()
}

func (a *fsAdapter) SavePolicy(model.Model) error                { return errNotImplemented }
func (a *fsAdapter) AddPolicy(string, string, []string) error    { return errNotImplemented }
func (a *fsAdapter) RemovePolicy(string, string, []string) error { return errNotImplemented }
func (a *fsAdapter) RemoveFilteredPolicy(string, string, int, ...string) error {
	return errNotImplemented
}
