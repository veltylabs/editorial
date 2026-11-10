package editorial

import "webtyp.com/crud"

// StorePost is the store of Post.
func (m *Module) StorePost() crud.Store { return m.store }

// StorePostTransition is the store of PostTransition.
func (m *Module) StorePostTransition() crud.Store { return m.store }

// StorePublication is the store of Publication.
func (m *Module) StorePublication() crud.Store { return m.store }
