package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/richer421/q-deploy/app/hello_world"
	"github.com/richer421/q-deploy/app/hello_world/vo"
	"github.com/richer421/q-deploy/conf"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Server struct {
	helloWorld *hello_world.AppService
}

func NewServer() *Server {
	return &Server{
		helloWorld: hello_world.NewAppService(),
	}
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
	// Tool: call_api
	type callAPIArgs struct {
		Action string         `json:"action" jsonschema:"Action to call: hello_world.list, hello_world.get, hello_world.create, hello_world.update, hello_world.delete"`
		Params map[string]any `json:"params,omitempty" jsonschema:"Parameters for the action"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "call_api",
		Description: "Call q-deploy API directly",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args callAPIArgs) (*mcp.CallToolResult, any, error) {
		result, err := s.handleCallAPI(ctx, args.Action, args.Params)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)}},
			}, nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: result}},
		}, nil, nil
	})

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

func (s *Server) handleCallAPI(ctx context.Context, action string, params map[string]any) (string, error) {
	var result any
	var err error

	switch action {
	case "hello_world.list":
		req := &vo.ListReq{Page: 1, PageSize: 10}
		if p, ok := params["page"].(float64); ok {
			req.Page = int(p)
		}
		if p, ok := params["page_size"].(float64); ok {
			req.PageSize = int(p)
		}
		result, err = s.helloWorld.List(ctx, req)

	case "hello_world.get":
		id, ok := params["id"].(float64)
		if !ok {
			return "", fmt.Errorf("missing or invalid 'id' parameter")
		}
		result, err = s.helloWorld.Get(ctx, uint(id))

	case "hello_world.create":
		title, ok := params["title"].(string)
		if !ok {
			return "", fmt.Errorf("missing or invalid 'title' parameter")
		}
		desc, _ := params["description"].(string)
		req := &vo.CreateReq{Title: title, Description: desc}
		result, err = s.helloWorld.Create(ctx, req)

	case "hello_world.update":
		id, ok := params["id"].(float64)
		if !ok {
			return "", fmt.Errorf("missing or invalid 'id' parameter")
		}
		req := &vo.UpdateReq{}
		if v, ok := params["title"].(string); ok {
			req.Title = &v
		}
		if v, ok := params["description"].(string); ok {
			req.Description = &v
		}
		err = s.helloWorld.Update(ctx, uint(id), req)

	case "hello_world.delete":
		id, ok := params["id"].(float64)
		if !ok {
			return "", fmt.Errorf("missing or invalid 'id' parameter")
		}
		err = s.helloWorld.Delete(ctx, uint(id))

	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}

	if err != nil {
		return "", err
	}

	if result == nil {
		return "ok", nil
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data), nil
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
