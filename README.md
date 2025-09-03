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
- Create a `.env` file in the project or your working directory with:
  - `DEBUG=true` to enable debug logs, or set `LOG_LEVEL=debug|info|warn|error|trace`.
  - Example:
    ```
    DEBUG=true
    LOG_LEVEL=debug
    ```
  Chatbang uses logrus for structured logging and loads `.env` via `joho/godotenv`.

## How it works?

Read that article: [https://ahmedhosssam.github.io/posts/chatbang/](https://ahmedhosssam.github.io/posts/chatbang/)
