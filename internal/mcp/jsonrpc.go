package mcp

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io"
)

// Minimal JSON-RPC 2.0 types

type Request struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      json.RawMessage `json:"id,omitempty"`
    Method  string          `json:"method"`
    Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      json.RawMessage `json:"id,omitempty"`
    Result  any             `json:"result,omitempty"`
    Error   *Error          `json:"error,omitempty"`
}

type Error struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

// readNDJSON reads one JSON value per line (NDJSON framing).
func readNDJSON(r *bufio.Reader) (*Request, error) {
    line, err := r.ReadBytes('\n')
    if err != nil {
        return nil, err
    }
    if len(line) == 0 {
        return nil, io.EOF
    }
    var req Request
    if err := json.Unmarshal(line, &req); err != nil {
        return nil, fmt.Errorf("invalid json-rpc frame: %w", err)
    }
    return &req, nil
}

func writeNDJSON(w io.Writer, resp *Response) error {
    resp.JSONRPC = "2.0"
    enc, err := json.Marshal(resp)
    if err != nil {
        return err
    }
    if _, err := w.Write(append(enc, '\n')) ; err != nil {
        return err
    }
    return nil
}

