package config

import (
    "bufio"
    "errors"
    "os"
    "path/filepath"
    "strconv"
    "strings"
)

type MCPServerEntry struct {
    Name          string
    Provider      string
    Roots         []string
    MaxBytes      int
    IncludeHidden bool
    AllowBinary   bool
}

type MCPConfig struct {
    Servers []MCPServerEntry
}

// LoadMCPConfig reads a minimal TOML config at path. Supported schema:
// [[mcp.servers]] tables with keys: name, provider, roots (array of strings),
// max_bytes (int), include_hidden (bool), allow_binary (bool).
func LoadMCPConfig(path string) (MCPConfig, error) {
    var cfg MCPConfig
    f, err := os.Open(path)
    if err != nil {
        return cfg, err
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
    inServers := false
    var cur *MCPServerEntry
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "#") { continue }
        switch {
        case line == "[[mcp.servers]]":
            inServers = true
            cfg.Servers = append(cfg.Servers, MCPServerEntry{})
            cur = &cfg.Servers[len(cfg.Servers)-1]
        case strings.HasPrefix(line, "["):
            // other tables not supported; ignore
            inServers = false
            cur = nil
        default:
            if !inServers || cur == nil { continue }
            k, v, ok := splitKV(line)
            if !ok { continue }
            switch k {
            case "name":
                cur.Name = trimQuotes(v)
            case "provider":
                cur.Provider = trimQuotes(v)
            case "roots":
                cur.Roots = parseStringArray(v)
            case "max_bytes":
                cur.MaxBytes = atoi(v)
            case "include_hidden":
                cur.IncludeHidden = parseBool(v)
            case "allow_binary":
                cur.AllowBinary = parseBool(v)
            }
        }
    }
    if err := scanner.Err(); err != nil {
        return cfg, err
    }
    // Normalize: if any relative roots are provided, make absolute
    for i := range cfg.Servers {
        for j, r := range cfg.Servers[i].Roots {
            if r == "" { continue }
            if !filepath.IsAbs(r) {
                abs, _ := filepath.Abs(r)
                cfg.Servers[i].Roots[j] = abs
            }
        }
    }
    if len(cfg.Servers) == 0 {
        return cfg, errors.New("no mcp servers configured")
    }
    return cfg, nil
}

func splitKV(line string) (string, string, bool) {
    idx := strings.Index(line, "=")
    if idx <= 0 { return "", "", false }
    k := strings.TrimSpace(line[:idx])
    v := strings.TrimSpace(line[idx+1:])
    return k, v, true
}

func trimQuotes(s string) string {
    s = strings.TrimSpace(s)
    if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") && len(s) >= 2 {
        return s[1:len(s)-1]
    }
    return s
}

func parseStringArray(s string) []string {
    s = strings.TrimSpace(s)
    if !strings.HasPrefix(s, "[") || !strings.HasSuffix(s, "]") { return nil }
    inner := strings.TrimSpace(s[1:len(s)-1])
    if inner == "" { return []string{} }
    parts := splitTopLevel(inner, ',')
    out := make([]string, 0, len(parts))
    for _, p := range parts {
        out = append(out, trimQuotes(strings.TrimSpace(p)))
    }
    return out
}

// splitTopLevel splits by sep ignoring separators inside quotes.
func splitTopLevel(s string, sep rune) []string {
    var out []string
    cur := strings.Builder{}
    inQ := false
    for _, r := range s {
        switch r {
        case '"':
            inQ = !inQ
            cur.WriteRune(r)
        default:
            if r == sep && !inQ { out = append(out, cur.String()); cur.Reset(); continue }
            cur.WriteRune(r)
        }
    }
    out = append(out, cur.String())
    return out
}

func atoi(s string) int {
    i, _ := strconv.Atoi(strings.TrimSpace(s))
    return i
}

func parseBool(s string) bool {
    s = strings.ToLower(strings.TrimSpace(s))
    return s == "true" || s == "1" || s == "yes"
}

