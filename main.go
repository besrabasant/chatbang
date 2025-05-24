package main

import (
	"context"
	"fmt"
	"log"
	"time"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/chromedp/chromedp"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Enter Text")
	
	myWindow.Resize(fyne.NewSize(400, 120))
	myWindow.SetFixedSize(true)

	var storedText string

	textEntry := widget.NewEntry()
	textEntry.SetPlaceHolder("Type here and press Enter...")

	textEntry.OnSubmitted = func(text string) {
		storedText = text
		
		myWindow.Close()
	}

	content := container.NewVBox(
		widget.NewLabel("Enter text:"),
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
			runClaude(storedText)
		} else if (strings.HasSuffix(storedText, "!chatgpt")) {
			storedText = strings.TrimSuffix(storedText, "!chatgpt")
			storedText = strings.TrimRight(storedText, " ")
			runChatGPT(storedText)
		} else {
			runDefault(storedText)
		}
	}
}

func runDefault(userPrompt string) {
	runChatGPT(userPrompt)
}

func runClaude(userPrompt string) {
	edgePath := "/usr/bin/microsoft-edge"

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

	taskCtx, taskCancel := context.WithTimeout(ctx, 200*time.Second)
	defer taskCancel()

	err := chromedp.Run(taskCtx,
		chromedp.Navigate(`https://claude.ai/new`),
		chromedp.WaitVisible(`//div[contains(@class, "ProseMirror") and contains(@class, "break-words") and contains(@class, "max-w-[60ch")]`),
		chromedp.Click(`//div[contains(@class, "ProseMirror") and contains(@class, "break-words") and contains(@class, "max-w-[60ch")]`),
		chromedp.SendKeys(`//div[contains(@class, "ProseMirror") and contains(@class, "break-words") and contains(@class, "max-w-[60ch]") and @contenteditable="true"]`, userPrompt),
		chromedp.Click(`//button[@aria-label="Send message"]`),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Browser interaction completed. Press Enter to exit.")
	//select {}
	fmt.Scanln()
}

func runChatGPT(userPrompt string) {
	edgePath := "/usr/bin/microsoft-edge"

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

	taskCtx, taskCancel := context.WithTimeout(ctx, 200*time.Second)
	defer taskCancel()

	err := chromedp.Run(taskCtx,
		chromedp.Navigate(`https://chatgpt.com`),
		chromedp.WaitVisible(`#prompt-textarea`, chromedp.ByID),
		chromedp.Click(`#prompt-textarea`, chromedp.ByID),
		chromedp.SendKeys(`#prompt-textarea`, userPrompt, chromedp.ByID),
		chromedp.Click(`#composer-submit-button`, chromedp.ByID),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Browser interaction completed. Press Enter to exit.")
	//select {}
	fmt.Scanln()
}
