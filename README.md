# Chatbang

`Chatbang` is a simple tool to access ChatGPT from the terminal.
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
