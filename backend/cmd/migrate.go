package cmd

import (
	"github.com/richer421/q-deploy/conf"
	"github.com/richer421/q-deploy/infra/mysql"
	"github.com/richer421/q-deploy/pkg/logger"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		if err := mysql.Init(conf.C.MySQL); err != nil {
			logger.Fatalf("mysql init failed: %v", err)
		}
		defer mysql.Close()

		logger.Infof("Running migration...")
		if err := mysql.Migrate(); err != nil {
			logger.Fatalf("migration failed: %v", err)
		}
		logger.Infof("Migration completed!")
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
