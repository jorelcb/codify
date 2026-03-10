package mcp

import (
	"context"
	"fmt"
	"io"

	"github.com/mark3labs/mcp-go/server"
)

// Transport defines the strategy for serving an MCP server.
type Transport interface {
	// Serve starts the MCP server and blocks until the context is cancelled or an error occurs.
	Serve(ctx context.Context, s *server.MCPServer) error
	// Name returns the transport identifier.
	Name() string
}

// StdioTransport serves the MCP server over stdin/stdout.
type StdioTransport struct {
	In  io.Reader
	Out io.Writer
}

func (t *StdioTransport) Serve(ctx context.Context, s *server.MCPServer) error {
	stdio := server.NewStdioServer(s)
	return stdio.Listen(ctx, t.In, t.Out)
}

func (t *StdioTransport) Name() string { return "stdio" }

// HTTPTransport serves the MCP server over HTTP (Streamable HTTP).
type HTTPTransport struct {
	Addr string // e.g. ":8080"
}

func (t *HTTPTransport) Serve(ctx context.Context, s *server.MCPServer) error {
	httpServer := server.NewStreamableHTTPServer(s)

	fmt.Printf("MCP server listening on %s\n", t.Addr)

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.Start(t.Addr)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return httpServer.Shutdown(context.Background())
	}
}

func (t *HTTPTransport) Name() string { return "http" }

// transports maps transport names to their factory functions.
var transports = map[string]func(addr string, in io.Reader, out io.Writer) Transport{
	"stdio": func(_ string, in io.Reader, out io.Writer) Transport {
		return &StdioTransport{In: in, Out: out}
	},
	"http": func(addr string, _ io.Reader, _ io.Writer) Transport {
		return &HTTPTransport{Addr: addr}
	},
}

// NewTransport resolves a transport strategy by name.
func NewTransport(name, addr string, in io.Reader, out io.Writer) (Transport, error) {
	factory, ok := transports[name]
	if !ok {
		available := make([]string, 0, len(transports))
		for k := range transports {
			available = append(available, k)
		}
		return nil, fmt.Errorf("unsupported transport: %s (available: %v)", name, available)
	}
	return factory(addr, in, out), nil
}
