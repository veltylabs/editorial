package editorial

import (
	"webtyp.com/crud"
	"webtyp.com/storage"
)

func (p *Post) Search(a crud.CurrentUser, filter *Post, text string, m *Module) ([]storage.Condition, error) {
	var conds []storage.Condition
	if text != "" {
        conds = append(conds, storage.Like("title", "%"+text+"%"))
	}
	if filter != nil {
		if filter.State != 0 {
			conds = append(conds, storage.Eq("state", filter.State))
		}
		if filter.AuthorId != "" {
			conds = append(conds, storage.Eq("author_id", filter.AuthorId))
		}
	}
	return conds, nil
}
