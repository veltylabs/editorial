package editorial

import (
	"webtyp.com/crud"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/storage"
)

func (m *Module) ReadPost(c crud.Caller, args crud.ReadArgs, filter model.Model) (PostList, error) {
	var p Post
	qb := m.db.Query(&p).Where("tenant_id").Eq(m.tenantID)

	if args.ID != "" {
		qb = qb.Where("id").Eq(args.ID)
	}
	if args.Search != "" {
		qb = qb.Where("title").Like("%" + args.Search + "%")
	}
	if args.After != "" {
		qb = qb.Where("id").Gt(args.After)
	}

	if filter != nil {
		f := filter.(*Post)
		if f.State != 0 {
			qb = qb.Where("state").Eq(f.State)
		}
		if f.AuthorId != "" {
			qb = qb.Where("author_id").Eq(f.AuthorId)
		}
	}

	qb = qb.OrderBy("id").Asc()
	if args.Limit > 0 {
		qb = qb.Limit(int(args.Limit))
	}

	var list PostList
	err := qb.ReadAll(
		func() model.Model { return &Post{} },
		func(rec model.Model) { list = append(list, rec.(*Post)) },
	)
	if err != nil && !orm.IsNotFound(err) && !storage.IsNoRows(err) {
		return nil, err
	}
	return list, nil
}

func (m *Module) ReadPostTransition(c crud.Caller, args crud.ReadArgs, filter model.Model) (PostTransitionList, error) {
	var pt PostTransition
	qb := m.db.Query(&pt).Where("tenant_id").Eq(m.tenantID)

	if args.ID != "" {
		qb = qb.Where("id").Eq(args.ID)
	}
	if args.After != "" {
		qb = qb.Where("id").Gt(args.After)
	}

	if filter != nil {
		f := filter.(*PostTransition)
		if f.PostId != "" {
			qb = qb.Where("post_id").Eq(f.PostId)
		}
	}

	qb = qb.OrderBy("id").Asc()
	if args.Limit > 0 {
		qb = qb.Limit(int(args.Limit))
	}

	var list PostTransitionList
	err := qb.ReadAll(
		func() model.Model { return &PostTransition{} },
		func(rec model.Model) { list = append(list, rec.(*PostTransition)) },
	)
	if err != nil && !orm.IsNotFound(err) && !storage.IsNoRows(err) {
		return nil, err
	}
	return list, nil
}

func (m *Module) ReadPublication(c crud.Caller, args crud.ReadArgs, filter model.Model) (PublicationList, error) {
	var p Publication
	qb := m.db.Query(&p).Where("tenant_id").Eq(m.tenantID)

	if args.ID != "" {
		qb = qb.Where("id").Eq(args.ID)
	}
	if args.After != "" {
		qb = qb.Where("id").Gt(args.After)
	}

	if filter != nil {
		f := filter.(*Publication)
		if f.PostId != "" {
			qb = qb.Where("post_id").Eq(f.PostId)
		}
		if f.Channel != "" {
			qb = qb.Where("channel").Eq(f.Channel)
		}
	}

	qb = qb.OrderBy("id").Asc()
	if args.Limit > 0 {
		qb = qb.Limit(int(args.Limit))
	}

	var list PublicationList
	err := qb.ReadAll(
		func() model.Model { return &Publication{} },
		func(rec model.Model) { list = append(list, rec.(*Publication)) },
	)
	if err != nil && !orm.IsNotFound(err) && !storage.IsNoRows(err) {
		return nil, err
	}
	return list, nil
}
