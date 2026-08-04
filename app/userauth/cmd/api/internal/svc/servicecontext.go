package svc

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/producer"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/config"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/accountcenterclient"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config              config.Config
	KqPusher            *producer.Pusher
	RedisClient         *redis.Redis
	AccountCenterClient accountcenterclient.AccountCenterClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:              c,
		KqPusher:            producer.NewPusher(c.KqConf),
		RedisClient:         redis.MustNewRedis(c.Infra.Redis),
		AccountCenterClient: accountcenterclient.NewAccountCenterClient(zrpc.MustNewClient(c.AccountCenterConf)),
	}
}
