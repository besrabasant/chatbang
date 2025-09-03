# Chatbang

`Chatbang` is a simple tool to access ChatGPT from the terminal, without needing for an API key.

![Chatbang](./assets/chatbang.png)

## Installation

On Linux:

```bash
curl -L https://github.com/besrabasant/chatbang/releases/latest/download/chatbang -o chatbang
chmod +x chatbang
sudo mv chatbang /usr/bin/chatbang
```

Install from source:

```bash
git clone git@github.com:besrabasant/chatbang.git
cd chatbang
go mod tidy
# Option A: Makefile
make build
sudo mv bin/chatbang /usr/bin/chatbang
# Option B: plain go
go build -o chatbang ./cmd/chatbang
sudo mv chatbang /usr/bin/chatbang
```

## Configuration

Note: Run `chatbang login` once to set up your profile and create the config directory `$HOME/.config/chatbang`.

`Chatbang` requires a Chromium-based browser (e.g. Chrome, Edge, Brave) to work, so you need to have one installed. And then make sure that the config file points to the right path of your chosen browser in the default config path for `Chatbang`: `$HOME/.config/chatbang/chatbang`.

Its default is:
```
browser=/usr/bin/google-chrome-stable
```

Change it to the right path of your favorite Chromium-based browser.

Note: `Chatbang` doesn't work when the browser is installed with `Snap`, the only option right now is to install it in `/bin` or `/usr/bin`.

Then, log in to ChatGPT in Chatbang's Chromium profile and allow clipboard permission:
```bash
chatbang login
```
This opens ChatGPT in the managed profile and prompts for clipboard permission. For backward compatibility, `chatbang --config` does the same.

## MCP Configuration

Chatbang can read local files and directories using pluggable MCP servers. Configuration lives at `~/.config/chatbang/mcp.toml`.

- Create a minimal config (with commented options):
  ```bash
  chatbang mcp init
  # or customize
  chatbang mcp init --name fs-local --root . --root ./internal --force
  ```
  This writes `~/.config/chatbang/mcp.toml`. Use `--root` multiple times to add roots; `--force` overwrites existing files.

- Example generated config (you can edit/uncomment as needed):
  ```toml
  [[mcp.servers]]
  # name = "fs-local"
  provider = "fs"
  roots = ["./"]
  # max_bytes = 1048576
  # include_hidden = false
  # allow_binary = false
  ```

- Multiple servers: Add more `[[mcp.servers]]` blocks if you register additional providers.

With MCP enabled, use the in‑chat commands below to attach files or explore directories.

## Usage

It's very simple, just type `chatbang` in the terminal.
```bash
chatbang
```

You can also pass a one‑shot prompt:
```bash
chatbang "Summarize https://platform.openai.com/docs/mcp in 5 bullets"
```

In‑chat commands for attaching context:
- :attach <path> [limit=N]
- :list [path] [depth=N]
- :search <root> <query> [globs=pat1,pat2]
- :stat <path>
- :clear, :help

Build and development (Makefile):
- make build: build binary to bin/chatbang
- make build-mcp: build MCP provider tool to bin/mcp (optional)
- make run: run chatbang
- make login: run chatbang login
- make tidy | fmt | vet | test | clean

Debug logging
- Start from the sample: `cp .env.sample .env`
- Create a `.env` file in the project or your working directory with:
  - `DEBUG=true` to enable debug logs, or set `LOG_LEVEL=debug|info|warn|error|trace`.
  - Example:
    ```
    DEBUG=true
    LOG_LEVEL=debug
    ```
  Chatbang uses logrus for structured logging and loads `.env` via `joho/godotenv`.
  - Optional: write logs to a file with `LOG_FILE` (path can use `~`):
    ```
    LOG_FILE=~/.config/chatbang/chatbang.log
    ```

## Troubleshooting

- MCP not loading: Enable debug logs and check `~/.config/chatbang/mcp.toml`. Run `chatbang mcp init` to create it; verify `roots` exist and are readable.
- Clipboard blocked: Run `chatbang login` and allow clipboard permission in the browser prompt. Ensure the session uses the Chatbang profile.
- Browser path invalid: Edit `~/.config/chatbang/chatbang` and set `browser=/path/to/chrome`. Verify with `which google-chrome-stable` or `chromium`.
- No GUI session: Chatbang requires a graphical environment. If on SSH/WSL/CI, use an X server or run locally.
- Still stuck: Set `DEBUG=true` in `.env` and re-run to see detailed logs. Consider switching log format to JSON if ingesting elsewhere.

## How it works?

Read that article: [https://ahmedhosssam.github.io/posts/chatbang/](https://ahmedhosssam.github.io/posts/chatbang/)
