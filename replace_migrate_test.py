import re

with open('tests/migrate_test.go', 'r') as f:
    content = f.read()

content = content.replace('return storage.Result{}, nil', 'return nil, nil')


with open('tests/migrate_test.go', 'w') as f:
    f.write(content)
