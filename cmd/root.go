package cmd

import (
	"log"

	"github.com/richer421/q-deploy/conf"
	"github.com/richer421/q-deploy/pkg/logger"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "q-deploy",
	Short: "q-deploy 后端服务",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := conf.Load(cfgFile); err != nil {
			return err
		}
		logger.Init(conf.C.Log)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "conf/config.yaml", "配置文件路径")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
