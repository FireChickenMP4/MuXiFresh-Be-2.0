package config

import (
	"MuXiFresh-Be-2.0/common/infra"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Infra infra.Config
}

func (c *Config) ApplyInfra() {
	c.Infra.ApplyEtcd(&c.RpcServerConf.Etcd)
}
