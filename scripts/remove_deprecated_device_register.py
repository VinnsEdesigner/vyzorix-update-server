#!/usr/bin/env python3
"""Remove deprecated /v1/device/register endpoint and update all references."""

import os
import re

ROOT = "/workspace/vyzorix-update-server"

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()
    
    original = content
    
    if '/v1/device/register' not in content:
        return False
    
    if 'server_routes.go' in filepath:
        content = re.sub(
            r'public\.POST\("/v1/device/register",\s*\n\s*middleware\.ValidationMiddleware\(&middleware\.DeviceRegisterSchema\{\}\),\s*\n\s*s\.deviceRegisterHandler\.Handle,\s*\n\s*\)',
            '// DEPRECATED: /v1/device/register removed. Use /v1/device/inbox for new registration flow.',
            content
        )
    
    if filepath.endswith('.md'):
        content = content.replace('POST /v1/device/register', 'POST /v1/device/inbox (DEPRECATED: was /v1/device/register)')
    
    if 'comprehensive.go' in filepath or 'test' in filepath:
        content = re.sub(r'(.*testRequest\("POST".*device/register.*)', r'// DEPRECATED: \1', content)
        content = re.sub(r'(.*"/v1/device/register".*)', r'// DEPRECATED: \1', content)
    
    if 'tenant_api_key.go' in filepath:
        content = re.sub(r'("/v1/device/register":\s*PathTypePublic)', r'// DEPRECATED: \1 // Use /v1/device/inbox', content)
    
    if 'request_signing.go' in filepath:
        content = re.sub(r'(if path == "/v1/device/register")', r'// DEPRECATED: \1', content)
    
    if 'device_registration_api.go' in filepath:
        content = re.sub(r'(\{"POST", "/v1/device/register")', r'// DEPRECATED: \1', content)
    
    if 'security_config_test.go' in filepath:
        content = re.sub(r'(.*"/v1/device/register".*)', r'// DEPRECATED: \1', content)
    
    if 'rate_limit_test.go' in filepath:
        content = re.sub(r'(.*"/v1/device/register".*)', r'// DEPRECATED: \1', content)
    
    if original != content:
        with open(filepath, 'w') as f:
            f.write(content)
        return True
    return False

if __name__ == '__main__':
    files = []
    for dirpath, _, filenames in os.walk(ROOT):
        if '.git' in dirpath or 'node_modules' in dirpath:
            continue
        for fn in filenames:
            if fn.endswith(('.go', '.md')):
                fp = os.path.join(dirpath, fn)
                with open(fp, 'r') as f:
                    if '/v1/device/register' in f.read():
                        files.append(fp)
    
    updated = [f for f in files if process_file(f)]
    print(f"Updated {len(updated)} files:")
    for f in updated:
        print(f"  - {f.replace(ROOT, '')}")
