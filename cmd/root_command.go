package cmd

import (
	"FkAdBot/depend"
	"fmt"
	"os"

	"github.com/johnpoint/go-bootstrap/core"
	"github.com/spf13/cobra"
)

var (
	rootCmd    = &cobra.Command{}
	configPath string
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(func() {
		if configPath == "" {
			configPath = "config_local.yaml"
		}
		core.AddGlobalComponent(
			&depend.Config{
				Path: configPath,
			},
		)
	})
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "config_local.yaml", "config file (default is ./config_local.yaml)")

	rootCmd.AddCommand(httpServerCommand)
}
