import re

with open('editorial.go', 'r') as f:
    content = f.read()

content = content.replace(
'''func (m *Module) ModelName() string {
\treturn "editorial"
}

func (m *Module) Definition() model.ModuleDefinition {
\treturn Definition
}
''', '')

with open('editorial.go', 'w') as f:
    f.write(content)
