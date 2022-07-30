package cmd

import (
	"FkAdBot/depend"
	"FkAdBot/pkg/log"
	"context"
	"github.com/johnpoint/go-bootstrap"
	"github.com/spf13/cobra"
)

var httpServerCommand = &cobra.Command{
	Use:   "api",
	Short: "Start http server",
	Run: func(cmd *cobra.Command, args []string) {
		err := bootstrap.NewBoot(
			context.Background(),
			&depend.Logger{},
			&depend.Bot{},
		).WithLogger(log.GetLogger()).Init()
		if err != nil {
			panic(err)
			return
		}

		forever := make(chan struct{})
		<-forever
	},
}
