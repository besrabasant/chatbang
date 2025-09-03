# Repository Guidelines

## Project Structure & Module Organization
- `main.go`: CLI entrypoint; orchestrates Chromium (chromedp) and terminal rendering.
- `internal/`: Private application code (keep non-public helpers here).
- `pkg/`: Reusable packages intended for import by other modules.
- `assets/`: Images and static assets (e.g., `assets/chatbang.png`).
- `go.mod` / `go.sum`: Module definition and dependencies.

## Build, Test, and Development Commands
- Build local binary: `go build -o bin/chatbang .`
- Run without building: `go run .`
- Configure/login flow: `./bin/chatbang --config`
- Dependency tidy: `go mod tidy`
- Lint/vet (recommended): `go vet ./...`
- Format: `gofmt -s -w .`

## Coding Style & Naming Conventions
- Follow standard Go style; run `gofmt` before committing.
- Use Go naming: exported `UpperCamelCase`, unexported `lowerCamelCase`.
- Keep functions small; prefer clear names over comments.
- Place internal-only code in `internal/`; shared libraries in `pkg/`.
- Avoid global state; pass context and dependencies explicitly.

## Testing Guidelines
- Framework: standard `testing` package; table-driven tests encouraged.
- File naming: `*_test.go` in the same package as code under test.
- Run tests: `go test ./... -v`
- Coverage (goal): `go test ./... -cover`
- Prefer deterministic tests; avoid requiring a live browser. Abstract chromedp calls behind interfaces and use fakes.

## Commit & Pull Request Guidelines
- Commits: imperative mood, concise subject (<= 72 chars), descriptive body when needed.
  - Example: `fix: handle clipboard permission errors during copy`
- PRs: include summary, rationale, and testing steps; link related issues.
- Screenshots/recordings: add when UX changes affect terminal output.
- Ensure CI passes (build, vet, tests) and code is formatted.

## Security & Configuration Tips
- Config path: `$HOME/.config/chatbang/chatbang`; set `browser=/usr/bin/<chromium>` (avoid Snap installs).
- Do not commit secrets or local paths. Use `.gitignore` for artifacts (e.g., `bin/`).
- When adding browser automation, request minimal permissions and document why.

