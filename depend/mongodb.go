package depend

import (
	"FkAdBot/config"
	"FkAdBot/dao/mongoDao"
	"context"
)

type MongoDB struct{}

func (r *MongoDB) Init(ctx context.Context) error {
	mongoDao.InitMongoClient(config.Config.MongoDBConfig)
	return nil
}
