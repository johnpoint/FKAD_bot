package depend

import (
	"FkAdBot/app/controller"
	"FkAdBot/config"
	"FkAdBot/pkg/bootstrap"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
)

type Api struct{}

var _ bootstrap.Component = (*Api)(nil)

func (r *Api) Init(ctx context.Context) error {
	gin.SetMode(gin.ReleaseMode)
	routerGin := gin.New()
	routerGin.GET("/ping", controller.Pong)

	go func() {
		fmt.Println("[init] HTTP Listen at " + config.Config.HttpServerListen)
		err := routerGin.Run(config.Config.HttpServerListen)
		if err != nil {
			panic(err)
		}
	}()
	return nil
}
