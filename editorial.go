package editorial

import (
	"webtyp.com/ddl"
	"webtyp.com/events"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/storage"
)

// domainError es el tipo concreto de los errores centinela de este paquete.
// El código los compara afirmando este tipo y comparando el valor: el uso de == entre
// dos valores de error se compila, bajo TinyGo, a runtime.interfaceEqual, lo que incluye
// internal/reflectlite en el binario de wasm.




const (
	ErrNotFound       model.NotFound = "editorial: not found"
	ErrAlreadyExists  model.Conflict = "editorial: already exists"
	ErrReasonRequired model.Invalid = "editorial: reason is required when requesting changes"
	ErrCannotRead     model.Invalid = "editorial: cannot read the request"
)

const (
	EventPostTransitioned = "editorial.post.transitioned"
	EventPostPublished    = "editorial.post.published"
)

type Deps struct {
	TenantID  string
	IDs       model.IDGenerator
	Publisher events.Publisher
}

type Module struct {
	tenantID string
	db   *orm.DB
	deps Deps
}

func New(db *orm.DB, deps Deps) (*Module, error) {
	if deps.IDs == nil {
		return nil, fmt.Err("editorial: IDs generator is required")
	}
	m := &Module{
		tenantID: deps.TenantID,
		db:   db,
		deps: deps,
	}

	if c, ok := db.RawConn().(ddl.Compiler); ok {
		if err := Migrate(db.RawConn(), c); err != nil {
			return nil, err
		}
	}

	return m, nil
}














func (m *Module) getPost(id string) (*Post, error) {
	post := &Post{}
	err := m.db.Query(post).
		Where("tenant_id").Eq(m.tenantID).
		Where("id").Eq(id).
		ReadOne()
	if err != nil {
		if orm.IsNotFound(err) || storage.IsNoRows(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return post, nil
}
