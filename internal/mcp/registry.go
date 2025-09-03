package mcp

// Provider is the pluggable MCP provider interface.
// It exposes MCP capabilities and handles method dispatch.
type Provider interface {
    // Name returns the provider name (e.g., "fs").
    Name() string
    // Capabilities returns the provider's advertised capabilities.
    Capabilities() Capabilities
    // Handle handles a JSON-RPC method with raw params; returns result or error.
    Handle(method string, params []byte) (any, *Error)
}

// Factory creates a Provider with implementation-specific options.
type Factory func(opts map[string]any) (Provider, error)

var registry = map[string]Factory{}

// Register makes a provider available by name.
func Register(name string, f Factory) {
    registry[name] = f
}

// Lookup finds a provider factory by name.
func Lookup(name string) Factory {
    return registry[name]
}

