import re

with open('view.go', 'r') as f:
    content = f.read()

search = """view.Ops{Module: ModelName, List: OpListPosts, Save: OpUpsertPost, Delete: OpDeletePost},"""
replace = """view.Ops{Module: ModelName, List: OpListPosts, Save: OpUpsertPost, Delete: OpDeletePost, Page: view.Page{First: 50, Size: 50, More: true}},"""

new_content = content.replace(search, replace)
if new_content == content:
    print("Failed to replace!")
else:
    with open('view.go', 'w') as f:
        f.write(new_content)
    print("Replaced successfully!")
