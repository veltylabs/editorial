//go:build wasm

package seed

import (
	"github.com/veltylabs/editorial"
	"github.com/veltylabs/editorial/ui"
	"webtyp.com/crud"
	"webtyp.com/ddl"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
	"webtyp.com/unixid"
	"webtyp.com/view"
)

// Demo crea una demostración ejecutable del módulo editorial sobre una base en memoria.
//
// Ejemplo: demo, err := seed.Demo()
func Demo() (view.Demo, error) {
	db := orm.New(mem.New())
	ids, err := unixid.NewUnixID()
	if err != nil {
		return view.Demo{}, err
	}
	mod, err := editorial.New(db, editorial.Deps{TenantID: "demo-tenant", IDs: ids})
	if err != nil {
		return view.Demo{}, err
	}
	return view.Demo{
		Label:    ui.DefaultLabel,
		Timezone: "America/Santiago",
		ActingAs: "u1",
		Modules:  []crud.Module{mod},
		Migrations: []ddl.Step{
			{Name: editorial.ModelName, Run: editorial.Migrate},
		},
		Screens: func(d view.BrowserDeps) []view.Screen {
			return ui.Screens(d.Caller)
		},
	}, nil
}
