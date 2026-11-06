import json

with open('lang.json', 'r') as f:
    data = json.load(f)

for key, translations in data["keys"].items():
    if not translations[0]:
        translations[0] = key

with open('lang.json', 'w') as f:
    json.dump(data, f, indent=2, ensure_ascii=False)
