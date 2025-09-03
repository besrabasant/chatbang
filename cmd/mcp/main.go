package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"gg/internal/mcp"
	_ "gg/internal/providers/fs" // register fs provider via init
)

func main() {
    var providerName string
    var rootsCSV string
    var maxBytes int
    var includeHidden bool
    var allowBinary bool

    flag.StringVar(&providerName, "provider", "fs", "MCP provider to run (e.g., fs)")
    flag.StringVar(&rootsCSV, "root", os.Getenv("FS_ROOTS"), "Colon- or comma-separated roots to allow")
    flag.IntVar(&maxBytes, "max-bytes", 1_048_576, "Max bytes to return for a single file read")
    flag.BoolVar(&includeHidden, "include-hidden", false, "Include dotfiles and hidden paths")
    flag.BoolVar(&allowBinary, "allow-binary", false, "Allow binary file reads (otherwise summarized)")
    flag.Parse()

    roots := []string{}
    if rootsCSV == "" {
        wd, err := os.Getwd()
        if err != nil {
            logrus.Fatalf("failed to get cwd: %v", err)
        }
        roots = []string{wd}
    } else {
        clean := strings.ReplaceAll(rootsCSV, ":", ",")
        for _, p := range strings.Split(clean, ",") {
            p = strings.TrimSpace(p)
            if p != "" {
                roots = append(roots, p)
            }
        }
    }

    logrus.WithFields(logrus.Fields{
        "provider": providerName,
        "roots":    roots,
        "maxBytes": maxBytes,
        "hidden":   includeHidden,
        "binary":   allowBinary,
    }).Info("starting MCP server")

    factory := mcp.Lookup(providerName)
    if factory == nil {
        logrus.Fatalf("unknown provider: %s", providerName)
    }

    opts := map[string]any{
        "roots":         roots,
        "maxBytes":      maxBytes,
        "includeHidden": includeHidden,
        "allowBinary":   allowBinary,
    }

    provider, err := factory(opts)
    if err != nil {
        logrus.Fatalf("provider init failed: %v", err)
    }

    srv := mcp.NewServer(provider)
    if err := srv.Serve(os.Stdin, os.Stdout); err != nil {
        fmt.Fprintf(os.Stderr, "mcp error: %v\n", err)
        os.Exit(1)
    }
}
