package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/richer421/q-deploy/app/deploy/vo"
	"github.com/richer421/q-deploy/conf"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type toolSpec struct {
	Name        string
	Description string
	Register    func(*mcp.Server)
}

type releaseService interface {
	ExecuteRelease(ctx context.Context, req *vo.ExecuteReleaseReq) (*vo.ReleaseDTO, error)
}

type Server struct {
	releaseSvc releaseService
}

func NewServer() *Server {
	return &Server{releaseSvc: &noopReleaseService{}}
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
	for _, spec := range s.toolSpecs() {
		spec.Register(server)
	}
}

func (s *Server) toolSpecs() []toolSpec {
	return []toolSpec{
		{
			Name:        "read_logs",
			Description: "Read last N lines from log file",
			Register: func(server *mcp.Server) {
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
						return s.errorResult(err), nil, nil
					}
					return &mcp.CallToolResult{
						Content: []mcp.Content{&mcp.TextContent{Text: result}},
					}, nil, nil
				})
			},
		},
		{
			Name:        "execute_release",
			Description: "Execute a release by deploy plan ID and build artifact ID",
			Register: func(server *mcp.Server) {
				type args struct {
					DeployPlanID    int64 `json:"deploy_plan_id"`
					BuildArtifactID int64 `json:"build_artifact_id"`
				}
				mcp.AddTool(server, &mcp.Tool{
					Name:        "execute_release",
					Description: "Execute a release by deploy plan ID and build artifact ID",
				}, func(ctx context.Context, req *mcp.CallToolRequest, in args) (*mcp.CallToolResult, any, error) {
					res, err := s.releaseSvc.ExecuteRelease(ctx, &vo.ExecuteReleaseReq{
						DeployPlanID:    in.DeployPlanID,
						BuildArtifactID: in.BuildArtifactID,
					})
					if err != nil {
						return s.errorResult(err), nil, nil
					}
					out, err := s.jsonResult(res)
					return out, nil, err
				})
			},
		},
	}
}

func (s *Server) jsonResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil
}

func (s *Server) errorResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(`{"error":%q}`, err.Error())}},
	}
}

type noopReleaseService struct{}

func (s *noopReleaseService) ExecuteRelease(_ context.Context, _ *vo.ExecuteReleaseReq) (*vo.ReleaseDTO, error) {
	return &vo.ReleaseDTO{}, nil
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
