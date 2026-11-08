package editorial

import (
	"webtyp.com/crud"
	"webtyp.com/model"
)

var _ crud.Module = (*Module)(nil)

// ModelName es la identidad de este módulo: mcp.HarvestOps califica cada operación
// a continuación como "editorial.<name>" en la red, y NewView en view.go pasa esta
// misma constante como view.Ops.Module para que el cliente componga el mismo nombre
// calificado idéntico.
const ModelName = "editorial"

func (m *Module) ModelName() string {
	return ModelName
}

func (m *Module) MountOperations(reg crud.Registry) {
	MountPostRead(reg, func() model.Model { return &Post{} }, m.ReadPost)
	MountPostCreate(reg, m.CreatePost)
	MountPostUpdate(reg, m.UpdatePost)
	MountPostDelete(reg, m.DeletePost)

	MountPostTransitionRead(reg, func() model.Model { return &PostTransition{} }, m.ReadPostTransition)
	MountPostTransitionCreate(reg, m.CreatePostTransition)
	MountPostTransitionUpdate(reg, m.UpdatePostTransition)
	MountPostTransitionDelete(reg, m.DeletePostTransition)

	MountPublicationRead(reg, func() model.Model { return &Publication{} }, m.ReadPublication)
	MountPublicationCreate(reg, m.CreatePublication)
	MountPublicationUpdate(reg, m.UpdatePublication)
	MountPublicationDelete(reg, m.DeletePublication)
}
