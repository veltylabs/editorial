import re

with open('view.go', 'r') as f:
    content = f.read()

content = content.replace(
'''func NewView(caller router.Caller) view.Presenter {
\tb := view.NewCallerLister(view.NewRecords().Caller(caller),
\t\tview.Ops{Module: ModelName, Allow: model.Create | model.Read | model.Update | model.Delete, Page: view.Page{First: 50, Size: 50, More: true}},
\t\tfunc() model.ModelSlice { return &PostList{} })
\treturn view.New(b, &Post{},
\t\tview.WithTitle(titlePosts),
\t\tview.WithSearchPlaceholder(searchPlaceholderPosts))
}''',
'''func NewView(caller router.Caller) view.Presenter {
\tb := view.NewCallerLister(view.NewRecords().Caller(caller),
\t\tview.Ops{Module: "editorial", Page: view.Page{First: 50, Size: 50, More: true}},
\t\tfunc() model.TableSlice { return &PostList{} })
\treturn view.New(b, &Post{},
\t\tview.WithTitle(titlePosts),
\t\tview.WithSearchPlaceholder(searchPlaceholderPosts))
}'''
)

with open('view.go', 'w') as f:
    f.write(content)
