package mcp

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/richer421/q-deploy/conf"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Run() error {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "q-deploy",
		Version: "1.0.0",
	}, nil)

	s.registerTools(server)

	return server.Run(context.Background(), &mcp.StdioTransport{})
}

func (s *Server) registerTools(server *mcp.Server) {
	// Tool: read_logs
	type readLogsArgs struct {
		Lines int `json:"lines,omitempty" jsonschema:"Number of lines to read (default 100)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_logs",
		Description: "Read last N lines from log file",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args readLogsArgs) (*mcp.CallToolResult, any, error) {
		lines := args.Lines
		if lines <= 0 {
			lines = 100
		}
		result, err := s.handleReadLogs(lines)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)}},
			}, nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: result}},
		}, nil, nil
	})
}

func (s *Server) handleReadLogs(lines int) (string, error) {
	logPath := conf.C.Log.File.Path
	if logPath == "" {
		logPath = "logs/app.log"
	}

	content, err := readLastLines(logPath, lines)
	if err != nil {
		return "", err
	}

	return content, nil
}

func readLastLines(path string, n int) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}

	return strings.Join(lines, "\n"), scanner.Err()
}
