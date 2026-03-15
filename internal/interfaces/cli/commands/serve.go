package commands

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	mcpserver "github.com/jorelcb/codify/internal/interfaces/mcp"
)

// NewServeCmd creates the serve command for MCP server mode
func NewServeCmd() *cobra.Command {
	var (
		transport string
		addr      string
	)

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start as MCP (Model Context Protocol) server",
		Long: `Start codify as an MCP server, exposing tools for:
  - generate_context: Generate AI context files from a description
  - generate_specs: Generate SDD specs from existing context
  - analyze_project: Scan a project and generate context files

Transports:
  stdio - Standard input/output (default, for Claude Desktop and similar)
  http  - HTTP server (for remote deployments)

Configuration for Claude Desktop (claude_desktop_config.json):
  {
    "mcpServers": {
      "codify": {
        "command": "codify",
        "args": ["serve"],
        "env": {
          "ANTHROPIC_API_KEY": "your-key-here"
        }
      }
    }
  }

Remote HTTP mode:
  codify serve --transport http --addr :8080

Requires ANTHROPIC_API_KEY environment variable.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(transport, addr)
		},
	}

	cmd.Flags().StringVar(&transport, "transport", "stdio", "Transport mode: stdio, http")
	cmd.Flags().StringVar(&addr, "addr", ":8080", "Listen address for HTTP transport")

	return cmd
}

func runServe(transportName, addr string) error {
	s := mcpserver.NewServer()

	transport, err := mcpserver.NewTransport(transportName, addr, os.Stdin, os.Stdout)
	if err != nil {
		return err
	}

	return transport.Serve(context.Background(), s)
}
