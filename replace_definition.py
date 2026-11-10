import re

with open('definition.go', 'r') as f:
    content = f.read()

content = content.replace(
    'var PostModel = model.Definition{',
    '''var PostModel = model.Definition{
\tAllow: model.Create | model.Read | model.Update | model.Delete,
\tRules: model.Rules{
\t\tCreate: model.AddMyBusinessRule,
\t\tRead:   model.AddMyBusinessRule,
\t\tUpdate: model.AddMyBusinessRule,
\t\tDelete: model.HaveNoBusinessRule,
\t},'''
)

content = content.replace(
    'var PostTransitionModel = model.Definition{',
    '''var PostTransitionModel = model.Definition{
\tAllow: model.Create | model.Read,
\tRules: model.Rules{
\t\tCreate: model.AddMyBusinessRule,
\t\tRead:   model.HaveNoBusinessRule,
\t},'''
)

content = content.replace(
    'var PublicationModel = model.Definition{',
    '''var PublicationModel = model.Definition{
\tAllow: model.Read,
\tRules: model.Rules{
\t\tRead: model.HaveNoBusinessRule,
\t},'''
)

content += '''
var Definition = model.ModuleDefinition{
\tName: "editorial",
\tEntities: []model.Definition{
\t\tPostModel,
\t\tPostTransitionModel,
\t\tPublicationModel,
\t},
}
'''

content = content.replace(
    '{Name: "author_id", Type: model.Text(), NotNull: true, DB: &model.FieldDB{}},',
    '{Name: "author_id", Type: model.Text(), NotNull: true, ServerOnly: true, DB: &model.FieldDB{}},'
)

content = content.replace(
    '{Name: "published_at", Type: model.NanoTime(), OmitEmpty: true, DB: &model.FieldDB{}},',
    '{Name: "published_at", Type: model.NanoTime(), OmitEmpty: true, ServerOnly: true, DB: &model.FieldDB{}},'
)

content = content.replace(
    '{Name: "created_at", Type: model.NanoTime(), DB: &model.FieldDB{}},',
    '{Name: "created_at", Type: model.NanoTime(), ServerOnly: true, DB: &model.FieldDB{}},'
)

content = content.replace(
    '{Name: "updated_at", Type: model.NanoTime(), DB: &model.FieldDB{}},',
    '{Name: "updated_at", Type: model.NanoTime(), ServerOnly: true, DB: &model.FieldDB{}},'
)

with open('definition.go', 'w') as f:
    f.write(content)
