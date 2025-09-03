package root

import (
    "fmt"
    "os"
    "os/user"
    "path/filepath"
    "strings"

    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
)

var (
    mcpName  string
    mcpRoots []string
    mcpForce bool
)

var mcpCmd = &cobra.Command{
    Use:   "mcp",
    Short: "Manage MCP configuration",
}

var mcpInitCmd = &cobra.Command{
    Use:   "init",
    Short: "Generate a minimal MCP config at the default location",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Resolve config dir and path: ~/.config/chatbang/mcp.toml
        usr, err := user.Current()
        if err != nil { return err }
        configDir := filepath.Join(usr.HomeDir, ".config", "chatbang")
        cfgPath := filepath.Join(configDir, "mcp.toml")

        // Prepare content
        if len(mcpRoots) == 0 { mcpRoots = []string{"./"} }
        // format roots as TOML array
        quoted := make([]string, 0, len(mcpRoots))
        for _, r := range mcpRoots { quoted = append(quoted, fmt.Sprintf("\"%s\"", r)) }
        rootsLine := fmt.Sprintf("roots = [%s]", strings.Join(quoted, ", "))

        content := fmt.Sprintf(`# Minimal MCP config for Chatbang
# Uncomment and adjust options as needed

[[mcp.servers]]
# name = "%s"
provider = "fs"
%s
# max_bytes = 1048576
# include_hidden = false
# allow_binary = false
`, mcpName, rootsLine)

        if err := os.MkdirAll(configDir, 0o755); err != nil {
            return err
        }
        if _, err := os.Stat(cfgPath); err == nil && !mcpForce {
            return fmt.Errorf("config already exists at %s (use --force to overwrite)", cfgPath)
        }
        if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
            return err
        }
        logrus.WithField("path", cfgPath).Info("wrote MCP config")
        fmt.Printf("Created %s\n", cfgPath)
        return nil
    },
}

func init() {
    // Parent command
    rootCmd.AddCommand(mcpCmd)
    // Init subcommand
    mcpCmd.AddCommand(mcpInitCmd)

    // Flags
    mcpInitCmd.Flags().StringVar(&mcpName, "name", "fs-local", "Name for this MCP server entry")
    mcpInitCmd.Flags().StringSliceVar(&mcpRoots, "root", nil, "One or more roots to allow (repeat or comma-separated)")
    mcpInitCmd.Flags().BoolVar(&mcpForce, "force", false, "Overwrite existing config if present")
}

