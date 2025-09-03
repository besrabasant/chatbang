package mcp

import (
    "bufio"
    "encoding/json"
    "errors"
    "io"
)

// Capabilities mirrors the high-level MCP capabilities advertised by a provider.
type Capabilities struct {
    Resources bool `json:"resources"`
    Tools     bool `json:"tools"`
}

type Server struct {
    provider Provider
}

func NewServer(p Provider) *Server { return &Server{provider: p} }

// Serve processes JSON-RPC (NDJSON) over r/w stdio.
func (s *Server) Serve(r io.Reader, w io.Writer) error {
    br := bufio.NewReader(r)
    for {
        req, err := readNDJSON(br)
        if err != nil {
            if errors.Is(err, io.EOF) {
                return nil
            }
            return err
        }

        switch req.Method {
        case "initialize":
            // Return basic capabilities and provider name
            result := map[string]any{
                "serverInfo": map[string]any{
                    "name": s.provider.Name(),
                    "version": "0.1.0",
                },
                "capabilities": s.provider.Capabilities(),
            }
            _ = writeNDJSON(w, &Response{ID: req.ID, Result: result})

        case "resources/list", "resources/read", "tools/list", "tools/call", "ping", "shutdown":
            res, perr := s.provider.Handle(req.Method, req.Params)
            if perr != nil {
                _ = writeNDJSON(w, &Response{ID: req.ID, Error: perr})
                continue
            }
            _ = writeNDJSON(w, &Response{ID: req.ID, Result: res})

        default:
            _ = writeNDJSON(w, &Response{ID: req.ID, Error: &Error{Code: -32601, Message: "method not found"}})
        }
    }
}

// Helpers
func decodeParams[T any](raw []byte, dst *T) *Error {
    if len(raw) == 0 {
        return &Error{Code: -32602, Message: "missing params"}
    }
    if err := json.Unmarshal(raw, dst); err != nil {
        return &Error{Code: -32602, Message: "invalid params"}
    }
    return nil
}

