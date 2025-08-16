# Chatbang

`Chatbang` is a simple tool to access ChatGPT from the terminal, without needing for an API key.
## Installation

On Linux:

```bash
curl -L https://github.com/ahmedhosssam/chatbang/releases/download/v0.0.1/myprogram -o chatbang
chmod +x chatbang
sudo mv chatbang /usr/bin/chatbang
```

Install from source:

```bash
git clone git@github.com:ahmedhosssam/chatbang.git
cd chatbang
go mod init chatbang
go mod tidy
go build main.go
sudo mv main /usr/bin/chatbang
```


## Usage

It's very simple, just type `chatbang` in the terminal.
```bash
chatbang
```

## Configuration

`Chatbang` requires a Chromium-based browser (e.g. Chrome, Edge, Brave) to work, so you need to have one. And then make sure that it points to the right path to your chosen browser in the default config path for `Chatbang`: `$HOME/.config/chatbang/chatbang`.

It's default is:
```
browser=/usr/bin/google-chrome
```

Change it to your favorite Chromium-based browser.

You also need to log in to ChatGPT in `Chatbang`'s Chromium session, so you need to do:
```bash
chatbang --config
```
That will open `Chatbang`'s Chromium session on ChatGPT's website, log in with your account.

Then, you will need to allow the clipboard permission for ChatGPT's website (on the same session).

## How it works?

`Chatbang` works by scraping ChatGPT's website. It opens ChatGPT when you execute `Chatbang` and take your prompt and paste it in ChatGPT's prompt area and execute that prompt, then it waits for the response and take it as a markdown format and then render it in the terminal.
