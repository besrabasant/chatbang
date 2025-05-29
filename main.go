package main

import (
	"bufio"
	"os"
	"os/user"
	"context"
	"log"
	"time"
	"strings"
	"fmt"
	"image/color"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2"

	"github.com/chromedp/chromedp"
)

const ctxTime = 2000

type chatbangTheme struct{}

func (t chatbangTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{15, 23, 42, 255} // Dark slate background
	case theme.ColorNameInputBackground:
		return color.NRGBA{15, 23, 42, 255} // Dark slate background
	case theme.ColorNameButton:
		return color.NRGBA{59, 130, 246, 255} // Blue accent
	case theme.ColorNameDisabledButton:
		return color.NRGBA{71, 85, 105, 255}
	case theme.ColorNameForeground:
		return color.NRGBA{241, 245, 249, 255} // Light text
		//return color.NRGBA{0, 0, 0, 50}
	case theme.ColorNameDisabled:
		return color.NRGBA{241, 245, 249, 255} // Light text
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{148, 163, 184, 255} // Muted text
	case theme.ColorNamePrimary:
		return color.NRGBA{139, 92, 246, 255} // Purple accent
	case theme.ColorNameHover:
		return color.NRGBA{99, 102, 241, 255} // Indigo hover
	case theme.ColorNameFocus:
		return color.NRGBA{168, 85, 247, 255} // Purple focus
	case theme.ColorNameSelection:
		return color.NRGBA{59, 130, 246, 100} // Semi-transparent blue
	case theme.ColorNameShadow:
		return color.NRGBA{0, 0, 0, 50}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (t chatbangTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t chatbangTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t chatbangTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 14
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNameHeadingText:
		return 18
	case theme.SizeNameSubHeadingText:
		return 16
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInnerPadding:
		return 6
	case theme.SizeNameSeparatorThickness:
		return 1
	}
	return theme.DefaultTheme().Size(name)
}

// Custom entry widget with rounded corners effect
func createStyledEntry() *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("✨ Ask anything and press Enter...")
	return entry
}

// Custom rich text widget for better output display
func createStyledOutput(defaultLLM string) *widget.Entry {
	output := widget.NewMultiLineEntry()
	output.SetText("🤖 " + strings.Title(defaultLLM) + " Ready")
	output.Wrapping = fyne.TextWrapWord
	output.Disable() // Make it read-only but still selectable
	return output
}

// Create animated status indicator
func createStatusIndicator() *canvas.Circle {
	circle := canvas.NewCircle(color.NRGBA{34, 197, 94, 255}) // Green
	circle.Resize(fyne.NewSize(12, 12))
	return circle
}

// Create gradient background
func createGradientBackground() *canvas.LinearGradient {
	gradient := canvas.NewLinearGradient(
		color.NRGBA{15, 23, 42, 255},   // Dark slate
		color.NRGBA{30, 41, 59, 255},   // Slightly lighter slate
		90, // Vertical gradient
	)
	return gradient
}

func main() {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var defaultLLM string
	var defaultBrowser string
	profileDir := usr.HomeDir + "/.config/chatbang/profile_data"
	//profileDir := "/home/ahmed/.config/microsoft-edge"

	configFile, err := os.Open(usr.HomeDir + "/.config/chatbang/chatbang")

	// TODO: if the config directory is not created, create it.
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
	myApp.Settings().SetTheme(&chatbangTheme{})
	
	myWindow := myApp.NewWindow("Chatbang")
	myWindow.Resize(fyne.NewSize(480, 300))
	myWindow.SetFixedSize(true) // Allow resizing for better UX
	
	// Create gradient background
	gradient := createGradientBackground()
	
	// Create status indicator
	statusIndicator := createStatusIndicator()
	
	// Create styled components
	textEntry := createStyledEntry()
	output := createStyledOutput(defaultLLM)
	
	// Create header with status
	headerText := canvas.NewText("Chatbang", color.NRGBA{241, 245, 249, 255})
	headerText.TextStyle = fyne.TextStyle{Bold: true}
	headerText.TextSize = 20
	headerText.Alignment = fyne.TextAlignCenter
	
	statusText := canvas.NewText("● Ready", color.NRGBA{34, 197, 94, 255})
	statusText.TextSize = 12
	statusText.Alignment = fyne.TextAlignCenter
	
	// Create help text
	helpText := widget.NewLabel("💡 Commands: !claude, !chatgpt, !grok, !p (perplexity)")
	helpText.Wrapping = fyne.TextWrapWord
	
	// Create styled card-like container for the input
	inputCard := container.NewBorder(nil, nil, nil, nil, textEntry)
	
	// Create styled output container with scroll
	outputScroll := container.NewScroll(output)
	outputScroll.SetMinSize(fyne.NewSize(440, 440))
	
	// Create main content layout
	header := container.NewVBox(
		headerText,
		statusText,
		widget.NewSeparator(),
	)
	
	body := container.NewVBox(
		helpText,
		widget.NewCard("", "", outputScroll),
		widget.NewCard("", "", inputCard),
	)
	
	content := container.NewBorder(header, nil, nil, nil, body)
	
	// Add some padding around the main content
	paddedContent := container.NewPadded(content)
	
	// Stack gradient background with content
	finalContent := container.NewStack(gradient, paddedContent)
	
	myWindow.SetContent(finalContent)
	myWindow.CenterOnScreen()
	myWindow.Canvas().Focus(textEntry)
	
	// Animation function for status indicator
	animateStatus := func(isProcessing bool) {
		if isProcessing {
			statusText.Text = "● Thinking..."
			statusText.Color = color.NRGBA{251, 191, 36, 255} // Yellow
			statusIndicator.FillColor = color.NRGBA{251, 191, 36, 255}
		} else {
			statusText.Text = "● Ready"
			statusText.Color = color.NRGBA{34, 197, 94, 255} // Green
			statusIndicator.FillColor = color.NRGBA{34, 197, 94, 255}
		}
		statusText.Refresh()
		statusIndicator.Refresh()
	}
	
	// Enhanced Enter handler with animations
	textEntry.OnSubmitted = func(text string) {
		if strings.TrimSpace(text) == "" {
			return
		}
		
		// Start processing animation
		animateStatus(true)
		
		// Show user's input in a styled way
		userPrompt := "💬 You: " + text + "\n\n⏳ Thinking..."
		output.SetText(userPrompt)
		output.Refresh()
		
		go func(prompt string) {
			var result string
			var emoji string
			
			// Add small delay for better UX
			//time.Sleep(300 * time.Millisecond)
			
			switch {
			case strings.HasSuffix(prompt, "!claude"):
				p := strings.TrimSpace(strings.TrimSuffix(prompt, "!claude"))
				runClaude(p, defaultBrowser, profileDir)
				result = "Claude launched successfully!"
				emoji = "🎯"
			case strings.HasSuffix(prompt, "!chatgpt"):
				p := strings.TrimSpace(strings.TrimSuffix(prompt, "!chatgpt"))
				runChatGPT(p, defaultBrowser, profileDir)
				result = "ChatGPT launched successfully!"
				emoji = "🤖"
			case strings.HasSuffix(prompt, "!grok"):
				p := strings.TrimSpace(strings.TrimSuffix(prompt, "!grok"))
				runGrok(p, defaultBrowser, profileDir)
				result = "Grok launched successfully!"
				emoji = "🚀"
			case strings.HasSuffix(prompt, "!p"):
				p := strings.TrimSpace(strings.TrimSuffix(prompt, "!p"))
				runPerplexity(p, defaultBrowser, profileDir)
				result = "Perplexity launched successfully!"
				emoji = "🔍"
			default:
				result = runDefault(prompt, defaultBrowser, defaultLLM, profileDir)
				emoji = "✨"
			}
			
			// Safely update UI on main thread
			fyne.Do(func() {
				finalResponse := "💬 You: " + prompt + "\n\n" + emoji + " " + result
				output.SetText(finalResponse)
				output.Refresh()
				
				// Scroll to bottom
				outputScroll.ScrollToBottom()
				
				// Stop processing animation
				animateStatus(false)
				
				// Auto-resize window if needed
				if myWindow.Content().Size().Height < 350 {
					myWindow.Resize(fyne.NewSize(480, 350))
				}
			})
		}(text)
		
		// Clear entry for next prompt
		textEntry.SetText("")
	}
	
	myWindow.Show()
	myApp.Run()
}

func runDefault(userPrompt string, defaultBrowser string, defaultLLM string, profileDir string) string {
	var outputText string 
	if (defaultLLM == "chatgpt") {
		outputText = runChatGPT(userPrompt, defaultBrowser, profileDir)
	}
	if (defaultLLM == "claude") {
		runClaude(userPrompt, defaultBrowser, profileDir)
	}
	if (defaultLLM == "perplexity") {
		runPerplexity(userPrompt, defaultBrowser, profileDir)
	}
	if (defaultLLM == "grok") {
		runGrok(userPrompt, defaultBrowser, profileDir)
	}
	return outputText
}

func runPerplexity(userPrompt string, defaultBrowser string, profileDir string) {
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
			chromedp.UserDataDir(profileDir),
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

func runClaude(userPrompt string, defaultBrowser string, profileDir string) {
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
			chromedp.UserDataDir(profileDir),
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

func runGrok(userPrompt string, defaultBrowser string, profileDir string) {
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
			chromedp.UserDataDir(profileDir),
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

func runChatGPT(userPrompt string, defaultBrowser string, profileDir string) string {
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
			//chromedp.Flag("headless", false),
			chromedp.UserDataDir(profileDir),
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

	var text string

	err := chromedp.Run(taskCtx,
		chromedp.Navigate(`https://chatgpt.com`),
		chromedp.WaitVisible(`#prompt-textarea`, chromedp.ByID),
		chromedp.Click(`#prompt-textarea`, chromedp.ByID),
		chromedp.SendKeys(`#prompt-textarea`, userPrompt, chromedp.ByID),
		chromedp.Click(`#composer-submit-button`, chromedp.ByID),
		chromedp.Click(`#prompt-textarea`, chromedp.ByID),
		chromedp.WaitVisible(`div[class="markdown prose dark:prose-invert w-full break-words dark"]`, chromedp.ByQuery),
		chromedp.Text(`div[class="markdown prose dark:prose-invert w-full break-words dark"]`, &text, chromedp.ByQuery),
	)

	if err != nil {
		log.Fatal(err)
	}

	//done := make(chan bool)
	//go func() {
	//	ticker := time.NewTicker(ctxTime * time.Second)
	//	defer ticker.Stop()
	//	
	//	for {
	//		select {
	//		case <-ticker.C:
	//			// Try to execute a simple JavaScript command to check if browser is still alive
	//			err := chromedp.Run(ctx, chromedp.Evaluate(`document.readyState`, nil))
	//			if err != nil {
	//				// Browser is closed or context is invalid
	//				done <- true
	//				return
	//			}
	//		case <-ctx.Done():
	//			// Context was cancelled
	//			done <- true
	//			return
	//		}
	//	}
	//}()

	//<-done
	return text
}
