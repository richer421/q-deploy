package cmd

import (
	"github.com/richer421/q-deploy/conf"
	"github.com/richer421/q-deploy/infra/mysql"
	"github.com/richer421/q-deploy/pkg/logger"
	"github.com/richer421/q-deploy/pkg/mcp"

	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server for Claude integration",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize MySQL (required for app layer)
		if err := mysql.Init(conf.C.MySQL); err != nil {
			logger.Fatalf("mysql init failed: %v", err)
		}

		logger.Infof("Starting MCP server...")
		srv := mcp.NewServer()
		if err := srv.Run(); err != nil {
			logger.Fatalf("MCP server error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
