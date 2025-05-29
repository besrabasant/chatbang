package main

import (
	"bufio"
	"os"
	"io"
	"os/user"
	"context"
	"log"
	"time"
	"strings"
	"fmt"

	"github.com/chromedp/chromedp"
)

const ctxTime = 2000

func main() {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error fetching user info:", err)
		return
	}

	configDir := usr.HomeDir + "/.config/chatbang"
	configPath := configDir + "/chatbang"
	profileDir := usr.HomeDir + "/.config/chatbang/profile_data"

	err = os.MkdirAll(configDir, 0o755)
	if err != nil {
		fmt.Println("Error creating config directory:", err)
		return
	}

	configFile, err := os.OpenFile(configPath,
		os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		fmt.Println("Error opening config file:", err)
		return
	}
	defer configFile.Close()

	info, err := configFile.Stat()
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}

	if info.Size() == 0 {
		defaults := "browser=/usr/bin/google-chrome\n" +
			"llm=chatgpt\n"
		_, err = io.WriteString(configFile, defaults)
		if err != nil {
			fmt.Println("Error writing default config:", err)
			return
		}
		configFile.Seek(0, 0)
	}

	var defaultLLM string
	var defaultBrowser string

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

    promptScanner := bufio.NewScanner(os.Stdin)
    fmt.Print("> ")
    for promptScanner.Scan() {
        prompt := promptScanner.Text()

        // a channel to get back result
        type reply struct {
            text  string
        }
        replyCh := make(chan reply)

        go func(p string) {
            var res string

            switch {
            case p == "!login":
		loginProfile(defaultBrowser, profileDir)
            case p == "exit":
                os.Exit(0)

            case strings.HasSuffix(p, "!claude"):
                base := strings.TrimSpace(strings.TrimSuffix(p, "!claude"))
		modifiedPrompt := base + " (Make an answer in less than 5 lines)."
                res = runClaude(modifiedPrompt, defaultBrowser, profileDir)

            case strings.HasSuffix(p, "!chatgpt"):
                base := strings.TrimSpace(strings.TrimSuffix(p, "!chatgpt"))
		modifiedPrompt := base + " (Make an answer in less than 5 lines)."
                res = runChatGPT(modifiedPrompt, defaultBrowser, profileDir)

            case strings.HasSuffix(p, "!grok"):
                base := strings.TrimSpace(strings.TrimSuffix(p, "!grok"))
		modifiedPrompt := base + " (Make an answer in less than 5 lines)."
                runGrok(modifiedPrompt, defaultBrowser, profileDir)

            case strings.HasSuffix(p, "!p"):
                base := strings.TrimSpace(strings.TrimSuffix(p, "!p"))
		modifiedPrompt := base + " (Make an answer in less than 5 lines)."
                res = runPerplexity(modifiedPrompt, defaultBrowser, profileDir)

            default:
				modifiedPrompt := p + " (Make an answer in less than 5 lines)."
                res = runDefault(modifiedPrompt, defaultBrowser, defaultLLM, profileDir)
            }

            replyCh <- reply{text: res}
        }(prompt)

        // wait for it and then print
        r := <-replyCh
        fmt.Printf("%s\n\n", r.text)

        fmt.Print("> ")
    }
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

func runPerplexity(userPrompt string, defaultBrowser string, profileDir string) string {
	return "Perplexity is not implemented yet"
	edgePath := defaultBrowser

	userPrompt += " (Just answer with text, no images and no table)."

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

	defer cancel()

	ctx, cancel := chromedp.NewContext(allocatorCtx)
	defer cancel()

	taskCtx, taskCancel := context.WithTimeout(ctx, ctxTime*time.Second)
	defer taskCancel()

	var text string


	err := chromedp.Run(taskCtx,
		chromedp.Navigate(`https://www.perplexity.ai/`),
		chromedp.WaitVisible(`#ask-input`, chromedp.ByID),
		chromedp.Click(`#ask-input`, chromedp.ByID),
		chromedp.SendKeys(`#ask-input`, userPrompt, chromedp.ByID),
		chromedp.WaitVisible(`//button[@aria-label="Submit"]`),
		chromedp.Click(`//button[@aria-label="Submit"]`),
		chromedp.Click(`#ask-input`, chromedp.ByID),
		chromedp.WaitVisible(`div[class="prose text-pretty dark:prose-invert inline leading-normal break-words min-w-0 [word-break:break-word]"]`, chromedp.ByQuery),
		chromedp.Text(`div[class="prose text-pretty dark:prose-invert inline leading-normal break-words min-w-0 [word-break:break-word]"]`, &text, chromedp.ByQuery),
	)

	for i := 1; i <= 10000; i++ {
		var exists bool
		err := chromedp.Run(ctx,
			chromedp.EvaluateAsDevTools(`
				document.querySelector('div.prose.text-pretty') !== null
			`, &exists),
		)
		if err != nil {
			log.Fatal(err)
		}

		if !exists {
			log.Println("Div not found, breaking loop.")
			break
		}

		// Element exists, so extract its text
		var newText string
		if err := chromedp.Run(ctx,
			chromedp.Text(`div[class="prose text-pretty dark:prose-invert inline leading-normal break-words min-w-0 [word-break:break-word]"]`, &newText, chromedp.ByQuery),
		); err != nil {
			log.Fatal(err)
		}

		if newText == text {
			break
		}
		text = newText
	}

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

func runClaude(userPrompt string, defaultBrowser string, profileDir string) string {
	return "Claude is not implemented yet"
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

	var text string

	err := chromedp.Run(taskCtx,
		chromedp.Navigate(`https://claude.ai/new`),
		chromedp.WaitVisible(`//div[contains(@class, "ProseMirror") and contains(@class, "break-words") and contains(@class, "max-w-[60ch")]`),
		chromedp.Click(`//div[contains(@class, "ProseMirror") and contains(@class, "break-words") and contains(@class, "max-w-[60ch")]`),
		chromedp.SendKeys(`//div[contains(@class, "ProseMirror") and contains(@class, "break-words") and contains(@class, "max-w-[60ch]") and @contenteditable="true"]`, userPrompt),
		chromedp.Click(`//button[@aria-label="Send message"]`),
		chromedp.Click(`//div[contains(@class, "ProseMirror") and contains(@class, "break-words") and contains(@class, "max-w-[60ch")]`),
		chromedp.WaitVisible(`div[class="group  relative  -tracking-[0.015em] pb-3"]`, chromedp.ByQuery),
		chromedp.Text(`div[class="group  relative  -tracking-[0.015em] pb-3"]`, &text, chromedp.ByQuery),
	)
	if err != nil {
		log.Fatal(err)
	}

	sel := `div.group.relative.-tracking-\[0\.015em\].pb-3`

	for i := 1; i <= 10; i++ {
		// grab the latest text
		if err := chromedp.Run(taskCtx,
			// you could also wrap this in WaitVisible if needed
			chromedp.Text(sel, &text, chromedp.ByQuery),
		); err != nil {
			log.Fatal(err)
		}
		fmt.Println(text)
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

	return text
}

func runGrok(userPrompt string, defaultBrowser string, profileDir string) string {
	return "Grok is not implemented yet"
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
	return "Grok is not implemented yet"
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

func loginProfile(defaultBrowser string, profileDir string) {
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
		chromedp.Navigate(`https://www.chatgpt.com/`),
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
