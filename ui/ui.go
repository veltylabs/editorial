package ui

import (
	"github.com/veltylabs/editorial"
	"webtyp.com/router"
	"webtyp.com/view"
)

const DefaultLabel = "Editorial"

func Screens(caller router.Caller) []view.Screen {
	return []view.Screen{view.Screen{Presenter: editorial.NewView(caller)}}
}
