package cmd

import (
	"FkAdBot/depend"

	"github.com/johnpoint/go-bootstrap/core"
	"github.com/johnpoint/go-bootstrap/log"

	"github.com/spf13/cobra"
)

var httpServerCommand = &cobra.Command{
	Use:   "api",
	Short: "Start http server",
	Run: func(cmd *cobra.Command, args []string) {
		err := core.NewBoot(
			&depend.Logger{},
			&depend.Bot{},
		).WithLogger(log.GetLogger()).Init()
		if err != nil {
			panic(err)
		}

		forever := make(chan struct{})
		<-forever
	},
}
