package editorial

import "webtyp.com/ddl"

// Migrate crea o pone al día las tablas de este módulo. Usa Sync y no CreateTable: crea la tabla
// que falta y además agrega las columnas que un modelo ganó, así un campo nuevo llega a una base ya
// instalada. Flujo: https://github.com/webtyp/ddl/blob/main/docs/diagrams/SYNC_FLOW.md
//
// Ejemplo: err := editorial.Migrate(conn, compiler)
func Migrate(conn ddl.Execer, ddlCompiler ddl.Compiler) error {
	return ddl.New(conn, ddlCompiler).Sync(
		&Post{},
		&PostTransition{},
		&Publication{},
	)
}
