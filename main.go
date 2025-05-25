package main

import (
	"bufio"
	"os"
	"context"
	"log"
	"time"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/chromedp/chromedp"
)

const ctxTime = 2000

func main() {
	configFile, err := os.Open("/home/ahmed/.config/chatbang/chatbang")
	var defaultLLM string
	var defaultBrowser string

	if err != nil {
		defaultLLM = "ChatGPT"
		defaultBrowser = "chrome"
	}

	defer configFile.Close()


	scanner := bufio.NewScanner(configFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if (key == "browser") {
			defaultBrowser = value
		}

		if (key == "llm") {
			defaultLLM = value
		}
	}


	myApp := app.New()
	myWindow := myApp.NewWindow("Chatbang")
	
	myWindow.Resize(fyne.NewSize(400, 120))
	myWindow.SetFixedSize(true)

	var storedText string

	textEntry := widget.NewEntry()
	textEntry.SetPlaceHolder("Ask anything and press Enter...")

	textEntry.OnSubmitted = func(text string) {
		storedText = text
		
		myWindow.Close()
	}

	content := container.NewVBox(
		widget.NewLabel(strings.Title(defaultLLM)),
		textEntry,
	)

	myWindow.SetContent(content)
	myWindow.CenterOnScreen()
	
	myWindow.Canvas().Focus(textEntry)
	
	myWindow.ShowAndRun()
	
	if (len(storedText) > 0) {
		if (strings.HasSuffix(storedText, "!claude")) {
			storedText = strings.TrimSuffix(storedText, "!claude")
			storedText = strings.TrimRight(storedText, " ")
			runClaude(storedText, defaultBrowser)
		} else if (strings.HasSuffix(storedText, "!chatgpt")) {
			storedText = strings.TrimSuffix(storedText, "!chatgpt")
			storedText = strings.TrimRight(storedText, " ")
			runChatGPT(storedText, defaultBrowser)
		} else if (strings.HasSuffix(storedText, "!grok")) {
			storedText = strings.TrimSuffix(storedText, "!grok")
			storedText = strings.TrimRight(storedText, " ")
			runGrok(storedText, defaultBrowser)
		} else if (strings.HasSuffix(storedText, "!p")) {
			// p for perplexity
			storedText = strings.TrimSuffix(storedText, "!p")
			storedText = strings.TrimRight(storedText, " ")
			runPerplexity(storedText, defaultBrowser)
		} else {
			runDefault(storedText, defaultBrowser, defaultLLM)
		}
	}
}

func runDefault(userPrompt string, defaultBrowser string, defaultLLM string) {
	if (defaultLLM == "chatgpt") {
		runChatGPT(userPrompt, defaultBrowser)
	}
	if (defaultLLM == "claude") {
		runClaude(userPrompt, defaultBrowser)
	}
	if (defaultLLM == "perplexity") {
		runPerplexity(userPrompt, defaultBrowser)
	}
	if (defaultLLM == "grok") {
		runGrok(userPrompt, defaultBrowser)
	}
}

func runPerplexity(userPrompt string, defaultBrowser string) {
	edgePath := defaultBrowser

	allocatorCtx, cancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.ExecPath(edgePath),
			chromedp.Flag("disable-blink-features", "AutomationControlled"),
			chromedp.Flag("exclude-switches", "enable-automation"),
			chromedp.Flag("disable-extensions", false),
			chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
			chromedp.Flag("disable-default-apps", false),
			chromedp.Flag("disable-dev-shm-usage", false),
			chromedp.Flag("disable-gpu", false),
			chromedp.Flag("headless", false),
			chromedp.UserDataDir("/home/ahmed/config/microsoft-edge"),
			chromedp.Flag("profile-directory", "Default"),
			//chromedp.Flag("profile-directory", "Profile 1"),
		)...,
	)

	defer cancel()

	ctx, cancel := chromedp.NewContext(allocatorCtx)
	defer cancel()

	taskCtx, taskCancel := context.WithTimeout(ctx, ctxTime*time.Second)
	defer taskCancel()

	err := chromedp.Run(taskCtx,
		chromedp.Navigate(`https://www.perplexity.ai/`),
		chromedp.WaitVisible(`#ask-input`, chromedp.ByID),
		chromedp.Click(`#ask-input`, chromedp.ByID),
		chromedp.SendKeys(`#ask-input`, userPrompt, chromedp.ByID),
		chromedp.WaitVisible(`//button[@aria-label="Submit"]`),
		chromedp.Click(`//button[@aria-label="Submit"]`),
		chromedp.Click(`#ask-input`, chromedp.ByID),
	)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(ctxTime * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				// Try to execute a simple JavaScript command to check if browser is still alive
				err := chromedp.Run(ctx, chromedp.Evaluate(`document.readyState`, nil))
				if err != nil {
					// Browser is closed or context is invalid
					done <- true
					return
				}
			case <-ctx.Done():
				// Context was cancelled
				done <- true
				return
			}
		}
	}()

	<-done
}

func runClaude(userPrompt string, defaultBrowser string) {
	edgePath := defaultBrowser

	allocatorCtx, cancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.ExecPath(edgePath),
			chromedp.Flag("disable-blink-features", "AutomationControlled"),
			chromedp.Flag("exclude-switches", "enable-automation"),
			chromedp.Flag("disable-extensions", false),
			chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
			chromedp.Flag("disable-default-apps", false),
			chromedp.Flag("disable-dev-shm-usage", false),
			chromedp.Flag("disable-gpu", false),
			chromedp.Flag("headless", false),
			chromedp.UserDataDir("/home/ahmed/config/microsoft-edge"),
			chromedp.Flag("profile-directory", "Default"),
			//chromedp.Flag("profile-directory", "Profile 1"),
		)...,
	)

	defer cancel()

	ctx, cancel := chromedp.NewContext(allocatorCtx)
	defer cancel()

	taskCtx, taskCancel := context.WithTimeout(ctx, ctxTime*time.Second)
	defer taskCancel()

	err := chromedp.Run(taskCtx,
		chromedp.Navigate(`https://claude.ai/new`),
		chromedp.WaitVisible(`//div[contains(@class, "ProseMirror") and contains(@class, "break-words") and contains(@class, "max-w-[60ch")]`),
		chromedp.Click(`//div[contains(@class, "ProseMirror") and contains(@class, "break-words") and contains(@class, "max-w-[60ch")]`),
		chromedp.SendKeys(`//div[contains(@class, "ProseMirror") and contains(@class, "break-words") and contains(@class, "max-w-[60ch]") and @contenteditable="true"]`, userPrompt),
		chromedp.Click(`//button[@aria-label="Send message"]`),
		chromedp.Click(`//div[contains(@class, "ProseMirror") and contains(@class, "break-words") and contains(@class, "max-w-[60ch")]`),
	)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(ctxTime * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				err := chromedp.Run(ctx, chromedp.Evaluate(`document.readyState`, nil))
				if err != nil {
					done <- true
					return
				}
			case <-ctx.Done():
				done <- true
				return
			}
		}
	}()

	<-done
}

func runGrok(userPrompt string, defaultBrowser string) {
	edgePath := defaultBrowser

	allocatorCtx, cancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.ExecPath(edgePath),
			chromedp.Flag("disable-blink-features", "AutomationControlled"),
			chromedp.Flag("exclude-switches", "enable-automation"),
			chromedp.Flag("disable-extensions", false),
			chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
			chromedp.Flag("disable-default-apps", false),
			chromedp.Flag("disable-dev-shm-usage", false),
			chromedp.Flag("disable-gpu", false),
			chromedp.Flag("headless", false),
			chromedp.UserDataDir("/home/ahmed/config/microsoft-edge"),
			chromedp.Flag("profile-directory", "Default"),
			//chromedp.Flag("profile-directory", "Profile 1"),
		)...,
	)

	defer cancel()

	ctx, cancel := chromedp.NewContext(allocatorCtx)
	defer cancel()

	taskCtx, taskCancel := context.WithTimeout(ctx, ctxTime*time.Second)
	defer taskCancel()

	err := chromedp.Run(taskCtx,
		chromedp.Navigate(`https://grok.com/`),
		chromedp.WaitVisible(`textarea[aria-label="Ask Grok anything"]`), 
		chromedp.Click(`textarea[aria-label="Ask Grok anything"]`), 
		chromedp.SendKeys(`textarea[aria-label="Ask Grok anything"]`, userPrompt),
		chromedp.Click(`button[aria-label="Submit"]`),
		chromedp.Click(`textarea[aria-label="Ask Grok anything"]`), 
	)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(ctxTime * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				// Try to execute a simple JavaScript command to check if browser is still alive
				err := chromedp.Run(ctx, chromedp.Evaluate(`document.readyState`, nil))
				if err != nil {
					// Browser is closed or context is invalid
					done <- true
					return
				}
			case <-ctx.Done():
				// Context was cancelled
				done <- true
				return
			}
		}
	}()

	<-done
}

func runChatGPT(userPrompt string, defaultBrowser string) {
	edgePath := defaultBrowser

	allocatorCtx, cancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.ExecPath(edgePath),
			chromedp.Flag("disable-blink-features", "AutomationControlled"),
			chromedp.Flag("exclude-switches", "enable-automation"),
			chromedp.Flag("disable-extensions", false),
			chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
			chromedp.Flag("disable-default-apps", false),
			chromedp.Flag("disable-dev-shm-usage", false),
			chromedp.Flag("disable-gpu", false),
			chromedp.Flag("headless", false),
			chromedp.UserDataDir("/home/ahmed/config/microsoft-edge"),
			chromedp.Flag("profile-directory", "Default"),
			//chromedp.Flag("profile-directory", "Profile 1"),
		)...,
	)

	// allocatorCtx, cancel := chromedp.NewRemoteAllocator(context.Background(), "http://127.0.0.1:9222/")
	// that's an option if I want to use the default browser,
	// but i need to convince the user to open the browser using this script:
	// microsoft-edge --remote-debugging-port=9222 --user-data-dir="/home/ahmed/config/microsoft-edge" --profile-directory="Default" --disable-blink-features=AutomationControlled --exclude-switches=enable-automation --user-agent="Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36" --enable-extensions --enable-default-apps --enable-dev-shm-usage --enable-gpu

	defer cancel()

	ctx, cancel := chromedp.NewContext(allocatorCtx)
	defer cancel()

	taskCtx, taskCancel := context.WithTimeout(ctx, ctxTime*time.Second)
	defer taskCancel()

	err := chromedp.Run(taskCtx,
		chromedp.Navigate(`https://chatgpt.com`),
		chromedp.WaitVisible(`#prompt-textarea`, chromedp.ByID),
		chromedp.Click(`#prompt-textarea`, chromedp.ByID),
		chromedp.SendKeys(`#prompt-textarea`, userPrompt, chromedp.ByID),
		chromedp.Click(`#composer-submit-button`, chromedp.ByID),
		chromedp.Click(`#prompt-textarea`, chromedp.ByID),
	)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(ctxTime * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				// Try to execute a simple JavaScript command to check if browser is still alive
				err := chromedp.Run(ctx, chromedp.Evaluate(`document.readyState`, nil))
				if err != nil {
					// Browser is closed or context is invalid
					done <- true
					return
				}
			case <-ctx.Done():
				// Context was cancelled
				done <- true
				return
			}
		}
	}()

	<-done
}
