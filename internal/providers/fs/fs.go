package fs

import (
    "encoding/json"
    "errors"
    "io"
    "io/fs"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "unicode/utf8"

    mcp "gg/internal/mcp"
)

type provider struct {
    roots         []string
    maxBytes      int
    includeHidden bool
    allowBinary   bool
}

func (p *provider) Name() string { return "fs" }

func (p *provider) Capabilities() mcp.Capabilities {
    return mcp.Capabilities{Resources: true, Tools: true}
}

// Registration
func init() {
    mcp.Register("fs", func(opts map[string]any) (mcp.Provider, error) {
        pr := &provider{
            roots:         getSlice[string](opts, "roots", nil),
            maxBytes:      get[int](opts, "maxBytes", 1_048_576),
            includeHidden: get[bool](opts, "includeHidden", false),
            allowBinary:   get[bool](opts, "allowBinary", false),
        }
        if len(pr.roots) == 0 {
            wd, _ := os.Getwd()
            pr.roots = []string{wd}
        }
        return pr, nil
    })
}

func get[T any](m map[string]any, k string, def T) T {
    if v, ok := m[k]; ok {
        if cast, ok := v.(T); ok {
            return cast
        }
    }
    return def
}

func getSlice[T any](m map[string]any, k string, def []T) []T {
    v, ok := m[k]
    if !ok {
        return def
    }
    // Accept []any from JSON-y callers
    switch vv := v.(type) {
    case []T:
        return vv
    case []any:
        out := make([]T, 0, len(vv))
        for _, it := range vv {
            if cast, ok := it.(T); ok {
                out = append(out, cast)
            }
        }
        return out
    default:
        return def
    }
}

// JSON helpers
type listResourcesResult struct {
    Resources []resource `json:"resources"`
}

type resource struct {
    URI         string `json:"uri"`
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
}

type readResourceParams struct {
    URI string `json:"uri"`
}

type readResourceResult struct {
    Contents  string `json:"contents"`
    MimeType  string `json:"mimeType"`
    Truncated bool   `json:"truncated"`
    Bytes     int    `json:"bytes"`
}

type toolsListResult struct {
    Tools []toolDesc `json:"tools"`
}

type toolDesc struct {
    Name        string         `json:"name"`
    Description string         `json:"description"`
    InputSchema map[string]any `json:"inputSchema"`
}

type toolCallParams struct {
    Name string         `json:"name"`
    Args map[string]any `json:"arguments"`
}

// Provider entry
func (p *provider) Handle(method string, params []byte) (any, *mcp.Error) {
    switch method {
    case "resources/list":
        items := make([]resource, 0, len(p.roots))
        for _, r := range p.roots {
            abs, _ := filepath.Abs(r)
            items = append(items, resource{
                URI:         "file://" + abs,
                Name:        filepath.Base(abs),
                Description: "Root directory",
            })
        }
        return &listResourcesResult{Resources: items}, nil

    case "resources/read":
        var in readResourceParams
        if err := json.Unmarshal(params, &in); err != nil || in.URI == "" {
            return nil, &mcp.Error{Code: -32602, Message: "invalid params"}
        }
        path := strings.TrimPrefix(in.URI, "file://")
        if !p.allowed(path) {
            return nil, &mcp.Error{Code: -32000, Message: "access denied"}
        }
        data, mt, truncated, n, err := p.readFile(path)
        if err != nil {
            return nil, &mcp.Error{Code: -32001, Message: err.Error()}
        }
        return &readResourceResult{Contents: data, MimeType: mt, Truncated: truncated, Bytes: n}, nil

    case "tools/list":
        return &toolsListResult{Tools: []toolDesc{
            {
                Name:        "fs.list",
                Description: "List directory entries under a path",
                InputSchema: schema(map[string]string{"path": "string", "depth?": "number"}),
            },
            {
                Name:        "fs.stat",
                Description: "Get file or directory metadata",
                InputSchema: schema(map[string]string{"path": "string"}),
            },
            {
                Name:        "fs.read",
                Description: "Read text file content (chunkable)",
                InputSchema: schema(map[string]string{"path": "string", "offset?": "number", "limit?": "number"}),
            },
            {
                Name:        "fs.search",
                Description: "Search for text inside files",
                InputSchema: schema(map[string]string{"root?": "string", "query": "string", "globs?": "array:string"}),
            },
        }}, nil

    case "tools/call":
        var in toolCallParams
        if err := json.Unmarshal(params, &in); err != nil {
            return nil, &mcp.Error{Code: -32602, Message: "invalid params"}
        }
        switch in.Name {
        case "fs.list":
            return p.toolList(in.Args)
        case "fs.stat":
            return p.toolStat(in.Args)
        case "fs.read":
            return p.toolRead(in.Args)
        case "fs.search":
            return p.toolSearch(in.Args)
        default:
            return nil, &mcp.Error{Code: -32601, Message: "unknown tool"}
        }

    case "ping":
        return map[string]string{"status": "ok"}, nil
    case "shutdown":
        return map[string]string{"status": "bye"}, nil
    default:
        return nil, &mcp.Error{Code: -32601, Message: "method not found"}
    }
}

// Tools implementations
func (p *provider) toolList(args map[string]any) (any, *mcp.Error) {
    path, _ := strArg(args, "path")
    depth := intArg(args, "depth", 1)
    if path == "" {
        path = p.roots[0]
    }
    if !p.allowed(path) {
        return nil, &mcp.Error{Code: -32000, Message: "access denied"}
    }
    out := []map[string]any{}
    baseDepth := strings.Count(filepath.Clean(path), string(os.PathSeparator))
    err := filepath.WalkDir(path, func(pth string, d fs.DirEntry, err error) error {
        if err != nil { return err }
        if !p.includeHidden && isHidden(pth) { if pth != path { return filepath.SkipDir } }
        curDepth := strings.Count(filepath.Clean(pth), string(os.PathSeparator)) - baseDepth
        if curDepth > depth { if d.IsDir() { return filepath.SkipDir } ; return nil }
        if pth != path {
            info, _ := d.Info()
            out = append(out, map[string]any{
                "path":  pth,
                "name":  d.Name(),
                "dir":   d.IsDir(),
                "size":  sizeOf(info),
                "mode":  info.Mode().String(),
            })
        }
        return nil
    })
    if err != nil { return nil, &mcp.Error{Code: -32002, Message: err.Error()} }
    return map[string]any{"entries": out}, nil
}

func (p *provider) toolStat(args map[string]any) (any, *mcp.Error) {
    path, _ := strArg(args, "path")
    if path == "" || !p.allowed(path) { return nil, &mcp.Error{Code: -32000, Message: "access denied"} }
    info, err := os.Stat(path)
    if err != nil { return nil, &mcp.Error{Code: -32002, Message: err.Error()} }
    return map[string]any{
        "path": path,
        "name": info.Name(),
        "dir":  info.IsDir(),
        "size": sizeOf(info),
        "mode": info.Mode().String(),
        "mod":  info.ModTime().UTC().Format("2006-01-02T15:04:05Z"),
    }, nil
}

func (p *provider) toolRead(args map[string]any) (any, *mcp.Error) {
    path, _ := strArg(args, "path")
    if path == "" || !p.allowed(path) { return nil, &mcp.Error{Code: -32000, Message: "access denied"} }
    offset := intArg(args, "offset", 0)
    limit := intArg(args, "limit", p.maxBytes)
    if limit > p.maxBytes { limit = p.maxBytes }
    f, err := os.Open(path)
    if err != nil { return nil, &mcp.Error{Code: -32002, Message: err.Error()} }
    defer f.Close()
    if offset > 0 {
        if _, err := f.Seek(int64(offset), io.SeekStart); err != nil {
            return nil, &mcp.Error{Code: -32002, Message: err.Error()}
        }
    }
    buf := make([]byte, limit)
    n, err := io.ReadFull(f, buf)
    if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
        if !errors.Is(err, io.EOF) { return nil, &mcp.Error{Code: -32002, Message: err.Error()} }
    }
    buf = buf[:n]
    // Binary/text heuristic
    if !p.allowBinary {
        if !utf8.Valid(buf) {
            return map[string]any{"path": path, "contents": "<binary omitted>", "truncated": true}, nil
        }
    }
    mt := http.DetectContentType(buf)
    return map[string]any{"path": path, "contents": string(buf), "mimeType": mt, "truncated": n == limit}, nil
}

func (p *provider) toolSearch(args map[string]any) (any, *mcp.Error) {
    root, _ := strArg(args, "root")
    if root == "" { root = p.roots[0] }
    if !p.allowed(root) { return nil, &mcp.Error{Code: -32000, Message: "access denied"} }
    query, _ := strArg(args, "query")
    if query == "" { return map[string]any{"matches": []any{}}, nil }
    globsArg, _ := args["globs"].([]any)
    globs := make([]string, 0, len(globsArg))
    for _, g := range globsArg { if s, ok := g.(string); ok { globs = append(globs, s) } }

    matches := []map[string]any{}
    filepath.WalkDir(root, func(pth string, d fs.DirEntry, err error) error {
        if err != nil { return nil }
        if d.IsDir() {
            if !p.includeHidden && isHidden(pth) && pth != root { return filepath.SkipDir }
            return nil
        }
        if !p.includeHidden && isHidden(pth) { return nil }
        if len(globs) > 0 && !anyGlob(globs, pth) { return nil }
        data, mt, _, n, err := p.readFile(pth)
        if err != nil || n == 0 { return nil }
        if strings.Contains(strings.ToLower(data), strings.ToLower(query)) {
            matches = append(matches, map[string]any{"path": pth, "mimeType": mt})
        }
        return nil
    })
    return map[string]any{"matches": matches}, nil
}

// Internals
func (p *provider) readFile(path string) (string, string, bool, int, error) {
    fi, err := os.Stat(path)
    if err != nil { return "", "", false, 0, err }
    if fi.IsDir() { return "", "", false, 0, errors.New("is a directory") }
    f, err := os.Open(path)
    if err != nil { return "", "", false, 0, err }
    defer f.Close()
    limit := p.maxBytes
    buf := make([]byte, limit)
    n, err := io.ReadFull(f, buf)
    if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
        if !errors.Is(err, io.EOF) { return "", "", false, 0, err }
    }
    buf = buf[:n]
    mt := http.DetectContentType(buf)
    if !p.allowBinary && !utf8.Valid(buf) {
        return "<binary omitted>", mt, true, n, nil
    }
    return string(buf), mt, n == limit, n, nil
}

func (p *provider) allowed(path string) bool {
    abs, err := filepath.Abs(path)
    if err != nil { return false }
    abs, _ = filepath.EvalSymlinks(abs)
    for _, r := range p.roots {
        rootAbs, _ := filepath.Abs(r)
        rootAbs, _ = filepath.EvalSymlinks(rootAbs)
        if strings.HasPrefix(abs+string(os.PathSeparator), rootAbs+string(os.PathSeparator)) || abs == rootAbs {
            return true
        }
    }
    return false
}

func isHidden(path string) bool {
    base := filepath.Base(path)
    return strings.HasPrefix(base, ".")
}

func anyGlob(globs []string, path string) bool {
    for _, g := range globs {
        if ok, _ := filepath.Match(g, path); ok { return true }
        if ok, _ := filepath.Match(g, filepath.Base(path)); ok { return true }
    }
    return false
}

func sizeOf(info os.FileInfo) int64 {
    if info == nil { return 0 }
    return info.Size()
}

// arg helpers
func strArg(m map[string]any, k string) (string, bool) {
    if v, ok := m[k]; ok { if s, ok := v.(string); ok { return s, true } }
    return "", false
}
func intArg(m map[string]any, k string, def int) int {
    if v, ok := m[k]; ok {
        switch vv := v.(type) {
        case float64:
            return int(vv)
        case int:
            return vv
        }
    }
    return def
}

// schema creates a minimal JSON schema object for tool inputs.
// Supported field types: string, number, boolean, array (of strings).
func schema(fields map[string]string) map[string]any {
    props := map[string]any{}
    required := []string{}
    for rawName, typ := range fields {
        name := rawName
        opt := false
        if strings.HasSuffix(name, "?") {
            name = strings.TrimSuffix(name, "?")
            opt = true
        }
        var prop map[string]any
        // Support array:<type> form
        if strings.HasPrefix(typ, "array") {
            itemType := "string"
            if i := strings.Index(typ, ":"); i >= 0 && i+1 < len(typ) {
                itemType = typ[i+1:]
            }
            prop = map[string]any{"type": "array", "items": map[string]any{"type": itemType}}
        } else {
            switch typ {
            case "string", "number", "boolean", "integer":
                prop = map[string]any{"type": typ}
            default:
                prop = map[string]any{"type": "string"}
            }
        }
        props[name] = prop
        if !opt { required = append(required, name) }
    }
    return map[string]any{
        "type":                 "object",
        "properties":           props,
        "required":             required,
        "additionalProperties": false,
    }
}
