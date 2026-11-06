import re
import os

for root, _, files in os.walk('tests'):
    for file in files:
        if file.endswith('.go'):
            path = os.path.join(root, file)
            with open(path, 'r') as f:
                content = f.read()

            new_content = content.replace("Deps{", "Deps{\n\t\tTenantID: \"tenant-test\",")
            if 'tenant-ops' in content:
                new_content = new_content.replace("Deps{\n\t\tTenantID: \"tenant-test\",", "Deps{\n\t\tTenantID: \"tenant-ops\",")
            elif 'tenant-a' in content:
                new_content = new_content.replace("Deps{\n\t\tTenantID: \"tenant-test\",", "Deps{\n\t\tTenantID: \"tenant-a\",")

            if new_content != content:
                with open(path, 'w') as f:
                    f.write(new_content)
                print(f"Patched {path}")
