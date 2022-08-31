package depend

import (
	"FkAdBot/config"
	"context"
	"math/rand"
	"time"

	"github.com/johnpoint/go-bootstrap/core"
)

type Config struct {
	Path string
}

var _ core.Component = (*Config)(nil)

func (d *Config) Init(ctx context.Context) error {
	rand.Seed(time.Now().UnixNano())
	return config.Config.SetPath(d.Path).ReadConfig()
}
