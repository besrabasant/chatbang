package main

import (
	"context"
	"fmt"
	"log"
	"time"

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
		runChromedp(storedText)
	}
}

func runChromedp(userPrompt string) {
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
			//chromedp.Flag("profile-directory", "Default"),
			chromedp.Flag("profile-directory", "Profile 1"),
		)...,
	)
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
	fmt.Scanln()
}
