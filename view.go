package editorial

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

const titlePosts = "Posts"
const searchPlaceholderPosts = "Search posts..."

// NewView construye el Presenter de artículos — el motor independiente de la tecnología
// que envuelve un renderizador (crudview, o cualquier otro). Este módulo lo construye
// (solo view + model + router); la aplicación decide qué renderizador lo dibuja.
//
// Ejemplo: presenter := editorial.NewView(caller)
func NewView(caller router.Caller) view.Presenter {
	b := view.NewCallerLister(view.NewRecords().Caller(caller),
		view.Ops{Module: ModelName, Allow: model.Create | model.Read | model.Update | model.Delete, Page: view.Page{First: 50, Size: 50, More: true}},
		func() model.ModelSlice { return &PostList{} })
	return view.New(b, &Post{},
		view.WithTitle(titlePosts),
		view.WithSearchPlaceholder(searchPlaceholderPosts))
}
