import re

with open('tests/editorial_test.go', 'r') as f:
    content = f.read()

content = content.replace(
'''	trans := editorial.PostTransition{PostId: post.Id, ToState: int64(editorial.StateApproved)}''',
'''	trans := editorial.PostTransition{Id: "t1", PostId: post.Id, ToState: int64(editorial.StateApproved)}'''
)


with open('tests/editorial_test.go', 'w') as f:
    f.write(content)
