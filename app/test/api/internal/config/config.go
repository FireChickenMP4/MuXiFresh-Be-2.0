package config

import (
	"MuXiFresh-Be-2.0/common/infra"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Infra    infra.Config
	TestConf zrpc.RpcClientConf
	JwtAuth  struct {
		AccessSecret string
		AccessExpire int64
	}
	UserConf zrpc.RpcClientConf
}

func (c *Config) ApplyInfra() {
	c.Infra.ApplyMiddlewares(&c.RestConf)
	c.Infra.ApplyEtcd(&c.TestConf.Etcd, &c.UserConf.Etcd)
}
