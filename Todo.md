#	Task	Why
2	Rate limit auth endpoints	Add stricter rate limiting on /v1/auth/login, /register, /forgot-password
Medium Priority (Quality)
#	Task	Why
3	E2E tests with Playwright	Test login, register, OAuth flows in real browser
4	Dark mode toggle	Settings page has appearance section but it's a stub
Lower Priority (Polish)
#	Task	Why
5	OpenAPI docs	Auto-generate API documentation
6	Database migrations	Proper migration system for schema changes




then workflows are failing; from PR.yml workflow; Run actions/github-script@v7
SyntaxError: Identifier 'github' has already been declared
    at new AsyncFunction (<anonymous>)
    at callAsyncFunction (/home/runner/work/_actions/actions/github-script/v7/dist/index.js:36187:16)
    at main (/home/runner/work/_actions/actions/github-script/v7/dist/index.js:36285:26)
    at /home/runner/work/_actions/actions/github-script/v7/dist/index.js:36260:1
    at /home/runner/work/_actions/actions/github-script/v7/dist/index.js:36317:3
    at Object.<anonymous> (/home/runner/work/_actions/actions/github-script/v7/dist/index.js:36320:12)
    at Module._compile (node:internal/modules/cjs/loader:1521:14)
    at Module._extensions..js (node:internal/modules/cjs/loader:1623:10)
    at Module.load (node:internal/modules/cjs/loader:1266:32)
    at Module._load (node:internal/modules/cjs/loader:1091:12)
Error: Unhandled error: SyntaxError: Identifier 'github' has already been declared


deploy.yml; Invalid workflow file: .github/workflows/deploy.yml#L1
(Line: 98, Col: 12): Unrecognized named-value: 'secrets'. Located at position 1 within expression: secrets.RENDER_SERVICE_URL, (Line: 108, Col: 13): Unrecognized named-value: 'secrets'. Located at position 1 within expression: secrets.RENDER_DEPLOY_HOOK_URL != '', (Line: 117, Col: 13): Unrecognized named-value: 'secrets'. Located at position 1 within expression: secrets.RENDER_SERVICE_URL != '', (Line: 141, Col: 12): Unrecognized named-value: 'secrets'. Located at position 1 within expression: secrets.PRODUCTION_URL, (Line: 151, Col: 13): Unrecognized named-value: 'secrets'. Located at position 1 within expression: secrets.PRODUCTION_DEPLOY_HOOK_URL != '', (Line: 160, Col: 13): Unrecognized named-value: 'secrets'. Located at position 1 within expression: secrets.PRODUCTION_URL != ''



then CI.yml all the jobs for linting and other jobs etc hit errors and all show that the codebase is full bugs🫠
