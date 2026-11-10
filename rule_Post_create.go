package editorial

import "webtyp.com/crud"

func (p *Post) Create(a crud.CurrentUser, tx crud.Tx, m *Module) (crud.RuleAnswer, error) {
	p.State = int64(StateDraft)
	p.AuthorId = a.UserID
	return crud.ApplyAndPublish, nil
}
