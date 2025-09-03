package app

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "os"
    "os/user"
    "path/filepath"
    "strconv"
    "strings"
    "time"

    markdown "github.com/MichaelMure/go-term-markdown"
    "github.com/chromedp/cdproto/runtime"
    "github.com/chromedp/chromedp"
    "github.com/sirupsen/logrus"

    mcp "gg/internal/mcp"
    _ "gg/internal/providers/fs" // register fs provider via init
)

const ctxTime = 2000

type attachment struct{ path string; content string }

type App struct {
    defaultBrowser string
    profileDir     string
    configDir      string
    mcpMgr         *mcp.Manager
}

func New() *App {
    usr, err := user.Current()
    if err != nil {
        panic(fmt.Sprintf("Error fetching user info: %v", err))
    }
    configDir := usr.HomeDir + "/.config/chatbang"
    profileDir := usr.HomeDir + "/.config/chatbang/profile_data"
    configPath := configDir + "/chatbang"

    if err := os.MkdirAll(configDir, 0o755); err != nil {
        panic(fmt.Sprintf("Error creating config directory: %v", err))
    }

    configFile, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE, 0o644)
    if err != nil {
        panic(fmt.Sprintf("Error opening config file: %v", err))
    }
    defer configFile.Close()

    info, err := configFile.Stat()
    if err != nil {
        panic(fmt.Sprintf("Error getting file info: %v", err))
    }
    if info.Size() == 0 {
        defaults := "browser=/usr/bin/google-chrome-stable\n"
        if _, err = io.WriteString(configFile, defaults); err != nil {
            panic(fmt.Sprintf("Error writing default config: %v", err))
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
        if key == "browser" {
            defaultBrowser = value
        }
    }

    a := &App{defaultBrowser: defaultBrowser, profileDir: profileDir, configDir: configDir}
    logrus.WithFields(logrus.Fields{
        "configDir":   configDir,
        "profileDir":  profileDir,
        "browser":     defaultBrowser,
    }).Info("initialized app config")
    a.initMCPProviders()
    return a
}

// Run starts interactive chat (or uses a provided first prompt).
func (a *App) Run(firstPrompt string) error {
    attachments := []attachment{}

    allocatorCtx, cancel := chromedp.NewExecAllocator(context.Background(),
        append(chromedp.DefaultExecAllocatorOptions[:],
            chromedp.ExecPath(a.defaultBrowser),
            chromedp.Flag("disable-blink-features", "AutomationControlled"),
            chromedp.Flag("exclude-switches", "enable-automation"),
            chromedp.Flag("disable-extensions", false),
            chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
            chromedp.Flag("disable-default-apps", false),
            chromedp.Flag("disable-dev-shm-usage", false),
            chromedp.Flag("disable-gpu", false),
            chromedp.Flag("headless", false),
            chromedp.UserDataDir(a.profileDir),
            chromedp.Flag("profile-directory", "Default"),
        )...,
    )
    defer cancel()

    ctx, cancel := chromedp.NewContext(allocatorCtx)
    defer cancel()

    taskCtx, taskCancel := context.WithTimeout(ctx, ctxTime*time.Second)
    defer taskCancel()

    logrus.WithFields(logrus.Fields{"browser": a.defaultBrowser}).Info("starting chat session and navigating to chatgpt.com")
    if err := chromedp.Run(taskCtx, chromedp.Navigate(`https://chatgpt.com`)); err != nil {
        return err
    }

    // Scanner loop
    fmt.Print("> ")
    promptScanner := bufio.NewScanner(os.Stdin)

    // If a first prompt is supplied via args, handle it once then exit
    if strings.TrimSpace(firstPrompt) != "" {
        line := strings.TrimSpace(firstPrompt)
        var b strings.Builder
        if len(attachments) > 0 {
            b.WriteString("You have access to the following context files. Use them when answering.\n\n")
            for _, at := range attachments {
                b.WriteString("File: ")
                b.WriteString(at.path)
                b.WriteString("\n````\n")
                b.WriteString(at.content)
                b.WriteString("\n````\n\n")
            }
        }
        b.WriteString(line)
        a.runChatGPT(taskCtx, b.String())
        return nil
    }

    for promptScanner.Scan() {
        line := strings.TrimSpace(promptScanner.Text())
        if line == "" {
            fmt.Print("> ")
            continue
        }
        if strings.HasPrefix(line, ":") {
            if a.handleLocalCommand(line, &attachments) {
                fmt.Print("> ")
                continue
            }
            fmt.Println("Unknown command. Try :help")
            fmt.Print("> ")
            continue
        }
        var b strings.Builder
        if len(attachments) > 0 {
            b.WriteString("You have access to the following context files. Use them when answering.\n\n")
            for _, at := range attachments {
                b.WriteString("File: ")
                b.WriteString(at.path)
                b.WriteString("\n````\n")
                b.WriteString(at.content)
                b.WriteString("\n````\n\n")
            }
        }
        b.WriteString(line)
        a.runChatGPT(taskCtx, b.String())
        return nil
    }
    return nil
}

// Login opens ChatGPT and prompts for clipboard permission within the profile.
func (a *App) Login() error {
    allocatorCtx, cancel := chromedp.NewExecAllocator(context.Background(),
        append(chromedp.DefaultExecAllocatorOptions[:],
            chromedp.ExecPath(a.defaultBrowser),
            chromedp.Flag("disable-blink-features", "AutomationControlled"),
            chromedp.Flag("exclude-switches", "enable-automation"),
            chromedp.Flag("disable-extensions", false),
            chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
            chromedp.Flag("disable-default-apps", false),
            chromedp.Flag("disable-dev-shm-usage", false),
            chromedp.Flag("disable-gpu", false),
            chromedp.Flag("headless", false),
            chromedp.UserDataDir(a.profileDir),
            chromedp.Flag("profile-directory", "Default"),
        )...,
    )
    defer cancel()

    ctx, cancel := chromedp.NewContext(allocatorCtx)
    defer cancel()

    taskCtx, taskCancel := context.WithTimeout(ctx, ctxTime*time.Second)
    defer taskCancel()

    logrus.Info("opening login/profile setup flow")
    if err := chromedp.Run(taskCtx,
        chromedp.Navigate(`https://www.chatgpt.com/`),
        chromedp.Evaluate(`(async () => {
            const permName = 'clipboard-read';
            try {
                const p = await navigator.permissions.query({ name: permName });
                if (p.state !== 'granted') {
                    alert("Please allow clipboard access in the popup that will appear now.");
                }
            } catch (e) {
                try { await navigator.clipboard.readText(); } catch (_) {
                    alert("Please allow clipboard access in the popup that will appear now.");
                }
            }
        })();`, nil),
        chromedp.Evaluate(`navigator.clipboard.readText().catch(() => {});`, nil),
    ); err != nil {
        return err
    }
    done := make(chan bool)
    go func() {
        ticker := time.NewTicker(ctxTime * time.Second)
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                if err := chromedp.Run(ctx, chromedp.Evaluate(`document.readyState`, nil)); err != nil {
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
    return nil
}

// init MCP
func (a *App) initMCPProviders() {
    mgr := mcp.NewManager()
    cfgPath := mcp.DefaultMCPConfigPath(a.configDir)
    if _, err := os.Stat(cfgPath); err != nil {
        a.mcpMgr = mgr
        logrus.WithField("path", cfgPath).Info("no MCP config found; MCP disabled")
        return
    }
    if err := mgr.LoadFromConfig(cfgPath); err != nil {
        logrus.WithError(err).Error("failed to load MCP config")
    } else {
        logrus.WithField("path", cfgPath).Info("loaded MCP config")
    }
    a.mcpMgr = mgr
}

// Local ":" commands
func (a *App) handleLocalCommand(line string, attachments *[]attachment) bool {
    fields := strings.Fields(line)
    if len(fields) == 0 {
        return true
    }
    cmd := strings.TrimPrefix(strings.ToLower(fields[0]), ":")
    switch cmd {
    case "help":
        fmt.Println("Commands:\n  :attach <path> [limit=N]\n  :list [path] [depth=N]\n  :search <root> <query> [globs=pat1,pat2]\n  :stat <path>\n  :clear (clear attachments)")
        return true
    case "clear":
        *attachments = (*attachments)[:0]
        fmt.Println("Attachments cleared.")
        return true
    case "attach":
        if len(fields) < 2 {
            fmt.Println("Usage: :attach <path> [limit=N]")
            return true
        }
        path := fields[1]
        limit := 0
        if len(fields) >= 3 && strings.HasPrefix(fields[2], "limit=") {
            limit, _ = strconv.Atoi(strings.TrimPrefix(fields[2], "limit="))
        }
        contents, truncated, err := a.providerRead(path, 0, limit)
        if err != nil {
            fmt.Printf("attach error: %v\n", err)
            return true
        }
        if truncated {
            fmt.Println("Note: content truncated.")
        }
        *attachments = append(*attachments, attachment{path: path, content: contents})
        fmt.Printf("Attached %s (%d chars).\n", path, len(contents))
        logrus.WithFields(logrus.Fields{"path": path, "chars": len(contents), "truncated": truncated}).Info(":attach")
        return true
    case "list":
        path := "."
        depth := 1
        if len(fields) >= 2 {
            path = fields[1]
        }
        if len(fields) >= 3 && strings.HasPrefix(fields[2], "depth=") {
            depth, _ = strconv.Atoi(strings.TrimPrefix(fields[2], "depth="))
        }
        entries, err := a.providerList(path, depth)
        if err != nil {
            fmt.Printf("list error: %v\n", err)
            return true
        }
        logrus.WithFields(logrus.Fields{"path": path, "depth": depth, "count": len(entries)}).Info(":list")
        for _, e := range entries {
            name, _ := e["name"].(string)
            dir, _ := e["dir"].(bool)
            p, _ := e["path"].(string)
            if rel, err := filepath.Rel(".", p); err == nil {
                p = rel
            }
            if dir {
                fmt.Printf("[D] %s\t%s\n", name, p)
            } else {
                fmt.Printf("[F] %s\t%s\n", name, p)
            }
        }
        return true
    case "search":
        if len(fields) < 3 {
            fmt.Println("Usage: :search <root> <query> [globs=pat1,pat2]")
            return true
        }
        root := fields[1]
        query := strings.Join(fields[2:], " ")
        globs := []string{}
        if i := strings.Index(query, "globs="); i >= 0 {
            parts := strings.Fields(query)
            nq := []string{}
            for _, tok := range parts {
                if strings.HasPrefix(tok, "globs=") {
                    rest := strings.TrimPrefix(tok, "globs=")
                    for _, p := range strings.Split(rest, ",") {
                        p = strings.TrimSpace(p)
                        if p != "" {
                            globs = append(globs, p)
                        }
                    }
                } else {
                    nq = append(nq, tok)
                }
            }
            query = strings.Join(nq, " ")
        }
        matches, err := a.providerSearch(root, query, globs)
        if err != nil {
            fmt.Printf("search error: %v\n", err)
            return true
        }
        logrus.WithFields(logrus.Fields{"root": root, "query": query, "globs": globs, "count": len(matches)}).Info(":search")
        for _, m := range matches {
            p, _ := m["path"].(string)
            mt, _ := m["mimeType"].(string)
            if rel, err := filepath.Rel(".", p); err == nil {
                p = rel
            }
            fmt.Printf("- %s (%s)\n", p, mt)
        }
        return true
    case "stat":
        if len(fields) < 2 {
            fmt.Println("Usage: :stat <path>")
            return true
        }
        info, err := a.providerStat(fields[1])
        if err != nil {
            fmt.Printf("stat error: %v\n", err)
            return true
        }
        logrus.WithFields(logrus.Fields{"path": fields[1]}).Info(":stat")
        b, _ := json.MarshalIndent(info, "", "  ")
        fmt.Println(string(b))
        return true
    default:
        return false
    }
}

// Provider helpers
func (a *App) getDefaultProvider() mcp.Provider {
    if a.mcpMgr == nil {
        return nil
    }
    names := a.mcpMgr.List()
    if len(names) == 0 {
        return nil
    }
    p, _ := a.mcpMgr.Provider(names[0])
    return p
}

func (a *App) providerRead(path string, offset, limit int) (contents string, truncated bool, err error) {
    p := a.getDefaultProvider()
    if p == nil {
        return "", false, fmt.Errorf("no MCP providers configured")
    }
    start := time.Now()
    args := map[string]any{"path": path, "offset": offset}
    if limit > 0 {
        args["limit"] = limit
    }
    req := map[string]any{"name": "fs.read", "arguments": args}
    raw, _ := json.Marshal(req)
    res, mErr := p.Handle("tools/call", raw)
    if mErr != nil {
        logrus.WithFields(logrus.Fields{"tool": "fs.read", "path": path, "offset": offset, "limit": limit, "elapsed": time.Since(start)}).WithError(fmt.Errorf(mErr.Message)).Error("mcp call failed")
        return "", false, fmt.Errorf(mErr.Message)
    }
    m, ok := res.(map[string]any)
    if !ok {
        return "", false, fmt.Errorf("unexpected response type")
    }
    contents, _ = m["contents"].(string)
    truncated, _ = m["truncated"].(bool)
    logrus.WithFields(logrus.Fields{"tool": "fs.read", "path": path, "bytes": len(contents), "truncated": truncated, "elapsed": time.Since(start)}).Info("mcp call")
    return contents, truncated, nil
}

func (a *App) providerList(path string, depth int) ([]map[string]any, error) {
    p := a.getDefaultProvider()
    if p == nil {
        return nil, fmt.Errorf("no MCP providers configured")
    }
    start := time.Now()
    args := map[string]any{"path": path, "depth": depth}
    req := map[string]any{"name": "fs.list", "arguments": args}
    raw, _ := json.Marshal(req)
    res, mErr := p.Handle("tools/call", raw)
    if mErr != nil {
        logrus.WithFields(logrus.Fields{"tool": "fs.list", "path": path, "depth": depth, "elapsed": time.Since(start)}).WithError(fmt.Errorf(mErr.Message)).Error("mcp call failed")
        return nil, fmt.Errorf(mErr.Message)
    }
    arr, ok := res.([]any)
    if !ok {
        return nil, fmt.Errorf("unexpected response type")
    }
    out := make([]map[string]any, 0, len(arr))
    for _, it := range arr {
        if m, ok := it.(map[string]any); ok {
            out = append(out, m)
        }
    }
    logrus.WithFields(logrus.Fields{"tool": "fs.list", "path": path, "depth": depth, "count": len(out), "elapsed": time.Since(start)}).Info("mcp call")
    return out, nil
}

func (a *App) providerSearch(root, query string, globs []string) ([]map[string]any, error) {
    p := a.getDefaultProvider()
    if p == nil {
        return nil, fmt.Errorf("no MCP providers configured")
    }
    start := time.Now()
    args := map[string]any{"root": root, "query": query}
    if len(globs) > 0 {
        args["globs"] = globs
    }
    req := map[string]any{"name": "fs.search", "arguments": args}
    raw, _ := json.Marshal(req)
    res, mErr := p.Handle("tools/call", raw)
    if mErr != nil {
        logrus.WithFields(logrus.Fields{"tool": "fs.search", "root": root, "query": query, "globs": globs, "elapsed": time.Since(start)}).WithError(fmt.Errorf(mErr.Message)).Error("mcp call failed")
        return nil, fmt.Errorf(mErr.Message)
    }
    arr, ok := res.([]any)
    if !ok {
        return nil, fmt.Errorf("unexpected response type")
    }
    out := make([]map[string]any, 0, len(arr))
    for _, it := range arr {
        if m, ok := it.(map[string]any); ok {
            out = append(out, m)
        }
    }
    logrus.WithFields(logrus.Fields{"tool": "fs.search", "root": root, "query": query, "globs": globs, "count": len(out), "elapsed": time.Since(start)}).Info("mcp call")
    return out, nil
}

func (a *App) providerStat(path string) (map[string]any, error) {
    p := a.getDefaultProvider()
    if p == nil {
        return nil, fmt.Errorf("no MCP providers configured")
    }
    start := time.Now()
    args := map[string]any{"path": path}
    req := map[string]any{"name": "fs.stat", "arguments": args}
    raw, _ := json.Marshal(req)
    res, mErr := p.Handle("tools/call", raw)
    if mErr != nil {
        logrus.WithFields(logrus.Fields{"tool": "fs.stat", "path": path, "elapsed": time.Since(start)}).WithError(fmt.Errorf(mErr.Message)).Error("mcp call failed")
        return nil, fmt.Errorf(mErr.Message)
    }
    m, ok := res.(map[string]any)
    if !ok {
        return nil, fmt.Errorf("unexpected response type")
    }
    logrus.WithFields(logrus.Fields{"tool": "fs.stat", "path": path, "elapsed": time.Since(start)}).Info("mcp call")
    return m, nil
}

// ChatGPT interaction logic (copied from previous main.go)
func (a *App) runChatGPT(taskCtx context.Context, modifiedPrompt string) {
    js := `new Promise((resolve, reject) => {
        navigator.permissions.query({ name: 'clipboard-read' }).then(permissionStatus => {
            if (permissionStatus.state === 'granted' || permissionStatus.state === 'prompt') {
                navigator.clipboard.readText().then(text => {
                    resolve(text);
                }).catch(err => {
                    resolve('Error reading clipboard: ' + err);
                });
            } else {
                resolve('Clipboard read not allowed: ' + permissionStatus.state);
            }
        }).catch(err => {
            resolve('Error querying permissions: ' + err);
        });
    });`

    var copiedText string
    var result []byte

    // Log prompt send (avoid logging full content at info level)
    preview := modifiedPrompt
    if len(preview) > 120 { preview = preview[:120] + "..." }
    logrus.WithFields(logrus.Fields{"chars": len(modifiedPrompt), "preview": preview}).Info("sending prompt to ChatGPT")

    err := chromedp.Run(taskCtx,
        chromedp.WaitVisible(`#prompt-textarea`, chromedp.ByID),
        chromedp.Click(`#prompt-textarea`, chromedp.ByID),
        chromedp.SendKeys(`#prompt-textarea`, modifiedPrompt, chromedp.ByID),
        chromedp.Click(`#composer-submit-button`, chromedp.ByID),
        chromedp.Click(`#prompt-textarea`, chromedp.ByID),
    )
    if err != nil {
        logrus.WithError(err).Error("failed to send prompt")
        log.Fatal(err)
    }

    // Scrape answer with copy button polling
    buttonDiv := `.markdown.prose.w-full:not(.result-streaming) [data-testid*="button-to-copy"]` // Generic copy button selector
    fmt.Printf("[Thinking...]\n\n")

    for {
        if copiedText != modifiedPrompt && len(copiedText) > 0 {
            break
        }
        err = chromedp.Run(taskCtx,
            chromedp.WaitVisible(buttonDiv, chromedp.ByQuery),
            chromedp.Evaluate(fmt.Sprintf(`
                (() => {
                    let buttons = document.querySelectorAll('%s');
                    if (buttons.length > 0) {
                        buttons[buttons.length - 1].click();
                    }
                })()
                `, buttonDiv), nil),
            chromedp.Evaluate(js, &copiedText, func(p *runtime.EvaluateParams) *runtime.EvaluateParams { return p.WithAwaitPromise(true) }),
        )
        result = markdown.Render(string(copiedText), 80, 2)
    }
    if err != nil {
        logrus.WithError(err).Error("failed while fetching response")
        log.Fatal(err)
    }
    logrus.WithField("chars", len(copiedText)).Info("received response from ChatGPT")
    fmt.Println(string(result))
    fmt.Print("> ")

    // Follow-up loop
    promptScanner := bufio.NewScanner(os.Stdin)
    for promptScanner.Scan() {
        prompt := promptScanner.Text()
        modifiedPrompt = prompt + " (Make an answer in less than 5 lines)."
        if len(prompt) == 0 {
            fmt.Print("> ")
            continue
        }
        fmt.Printf("[Thinking...]\n\n")
        // Log next prompt
        preview = modifiedPrompt
        if len(preview) > 120 { preview = preview[:120] + "..." }
        logrus.WithFields(logrus.Fields{"chars": len(modifiedPrompt), "preview": preview}).Info("sending follow-up prompt to ChatGPT")

        err := chromedp.Run(taskCtx,
            chromedp.WaitVisible(`#prompt-textarea`, chromedp.ByID),
            chromedp.Click(`#prompt-textarea`, chromedp.ByID),
            chromedp.SendKeys(`#prompt-textarea`, modifiedPrompt, chromedp.ByID),
            chromedp.Click(`#composer-submit-button`, chromedp.ByID),
            chromedp.Click(`#prompt-textarea`, chromedp.ByID),
        )
        if err != nil {
            logrus.WithError(err).Error("failed to send follow-up prompt")
            log.Fatal(err)
        }
        result = markdown.Render(string(copiedText), 80, 2)
        copiedText = ""
        for {
            if copiedText != modifiedPrompt && len(copiedText) > 0 {
                break
            }
            err = chromedp.Run(taskCtx,
                chromedp.Sleep(3*time.Second),
                chromedp.Evaluate(fmt.Sprintf(`
                    (() => {
                        let buttons = document.querySelectorAll('%s');
                        if (buttons.length > 0) {
                            buttons[buttons.length - 1].click();
                        }
                    })()
                    `, buttonDiv), nil),
                chromedp.Sleep(1*time.Second),
                chromedp.Evaluate(js, &copiedText, func(p *runtime.EvaluateParams) *runtime.EvaluateParams { return p.WithAwaitPromise(true) }),
            )
            result = markdown.Render(string(copiedText), 80, 2)
        }
        logrus.WithField("chars", len(copiedText)).Info("received response from ChatGPT")
        fmt.Println(string(result))
        fmt.Print("> ")
    }
}
