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
	"github.com/chromedp/cdproto/cdp"
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
	}

    //promptScanner := bufio.NewScanner(os.Stdin)
    //fmt.Print("> ")
    //for promptScanner.Scan() {
    //    prompt := promptScanner.Text()

    //    // a channel to get back result
    //    type reply struct {
    //        text  string
    //    }
    //    replyCh := make(chan reply)

    //    go func(p string) {
    //        var res string
    //        switch {
    //        case p == "!login":
    //            loginProfile(defaultBrowser, profileDir)
    //        case p == "exit":
    //            os.Exit(0)
    //        default:
    //            modifiedPrompt := p + " (Make an answer in less than 5 lines)."
    //            res = runDefault(modifiedPrompt, defaultBrowser, defaultLLM, profileDir)
    //        }

    //        replyCh <- reply{text: res}
    //    }(prompt)

    //    // wait for it and then print
    //    r := <-replyCh
    //    fmt.Printf("%s\n\n", r.text)

    //    fmt.Print("> ")
    //}

	runChatGPT(defaultBrowser, profileDir)
}

func waitForStableText(ctx context.Context, sel string, timeout time.Duration) (string, error) {
    var lastText string
    stableCount := 0
    start := time.Now()

    for {
        var currentText string
        err := chromedp.Run(ctx,
            chromedp.Text(sel, &currentText, chromedp.NodeVisible),
        )
        if err != nil {
            return "", err
        }

        if currentText == lastText && len(lastText) > 0 {
            stableCount++
        } else {
            stableCount = 0
        }

        if stableCount >= 3 {
            return currentText, nil
        }

        if time.Since(start) > timeout {
			fmt.Println("gg")
        }

        lastText = currentText
        time.Sleep(500 * time.Millisecond) // check every 0.5s
    }
}

func runChatGPT(defaultBrowser string, profileDir string) {
	browserPath := defaultBrowser

	allocatorCtx, cancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.ExecPath(browserPath),
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
		)...,
	)

	defer cancel()

	ctx, cancel := chromedp.NewContext(allocatorCtx)
	defer cancel()

	taskCtx, taskCancel := context.WithTimeout(ctx, ctxTime*time.Second)
	defer taskCancel()

	var text string

	//outputDiv := `div[class="markdown prose dark:prose-invert w-full break-words dark markdown-new-styling"]`
	outputDiv := `div.markdown.prose.dark\:prose-invert.w-full.break-words.dark.markdown-new-styling`

	go func() {
		err := chromedp.Run(taskCtx,
			chromedp.Navigate(`https://chatgpt.com`),
		)

		if err != nil {
			log.Fatal(err)
		}
	}()

    fmt.Print("> ")
    promptScanner := bufio.NewScanner(os.Stdin)
    for promptScanner.Scan() {
		prompt := promptScanner.Text()
		modifiedPrompt := prompt + " (Make an answer in less than 5 lines)."
		//fmt.Printf("Prompt: %s\n\n", modifiedPrompt)

		err := chromedp.Run(taskCtx,
			chromedp.WaitVisible(`#prompt-textarea`, chromedp.ByID),
			chromedp.Click(`#prompt-textarea`, chromedp.ByID),
			chromedp.SendKeys(`#prompt-textarea`, modifiedPrompt, chromedp.ByID),
			chromedp.Click(`#composer-submit-button`, chromedp.ByID),
			chromedp.Click(`#prompt-textarea`, chromedp.ByID),

			chromedp.WaitVisible(outputDiv, chromedp.ByQueryAll),
			chromedp.Text(outputDiv, &text, chromedp.ByQueryAll),
		)

		if err != nil {
			log.Fatal(err)
		}

		var nodes []*cdp.Node
		err = chromedp.Run(taskCtx,
		    chromedp.Nodes(outputDiv, &nodes, chromedp.ByQueryAll),
		)

		if err != nil {
		    log.Fatal(err)
		}

		if len(nodes) > 0 {
			    lastNode := nodes[len(nodes)-1]
			    selector := lastNode.FullXPath()
			    ress, err := waitForStableText(ctx, selector, 20*time.Second)
			    if err != nil {
				log.Println("Warning:", err)
			    }
			    fmt.Printf("%s\n\n", ress)
		}

		fmt.Print("> ")
	}
}

func loginProfile(defaultBrowser string, profileDir string) {
	browserPath := defaultBrowser

	allocatorCtx, cancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.ExecPath(browserPath),
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
