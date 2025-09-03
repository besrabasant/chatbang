package mcp

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"

    cfgpkg "gg/internal/config"
)

type Manager struct {
    providers map[string]Provider // name -> instance
}

func NewManager() *Manager {
    return &Manager{providers: map[string]Provider{}}
}

// LoadFromConfig reads TOML at path and initializes providers via registry.
func (m *Manager) LoadFromConfig(path string) error {
    cfg, err := cfgpkg.LoadMCPConfig(path)
    if err != nil {
        return err
    }
    for _, s := range cfg.Servers {
        name := s.Name
        if name == "" { name = s.Provider }
        f := Lookup(s.Provider)
        if f == nil {
            return fmt.Errorf("unknown provider: %s", s.Provider)
        }
        opts := map[string]any{
            "roots":         s.Roots,
            "maxBytes":      fallbackInt(s.MaxBytes, 1_048_576),
            "includeHidden": s.IncludeHidden,
            "allowBinary":   s.AllowBinary,
        }
        p, err := f(opts)
        if err != nil { return err }
        m.providers[name] = p
    }
    return nil
}

func fallbackInt(v, def int) int { if v <= 0 { return def } ; return v }

// Default config path
func DefaultMCPConfigPath(configDir string) string {
    if configDir == "" {
        wd, _ := os.Getwd()
        return filepath.Join(wd, "mcp.toml")
    }
    return filepath.Join(configDir, "mcp.toml")
}

// Provider returns a provider instance by name.
func (m *Manager) Provider(name string) (Provider, error) {
    p, ok := m.providers[name]
    if !ok { return nil, errors.New("provider not found") }
    return p, nil
}

// List returns registered provider names.
func (m *Manager) List() []string {
    names := make([]string, 0, len(m.providers))
    for k := range m.providers { names = append(names, k) }
    return names
}

