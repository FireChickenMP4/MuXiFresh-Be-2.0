package config

import (
	"MuXiFresh-Be-2.0/common/infra"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Infra   infra.Config
	JwtAuth struct {
		AccessSecret string
		AccessExpire int64
	}
	AssignmentConf zrpc.RpcClientConf
	SubmissionConf zrpc.RpcClientConf
	CommentConf    zrpc.RpcClientConf
	UserConf       zrpc.RpcClientConf
}

func (c *Config) ApplyInfra() {
	c.Infra.ApplyEtcd(
		&c.AssignmentConf.Etcd,
		&c.SubmissionConf.Etcd,
		&c.CommentConf.Etcd,
		&c.UserConf.Etcd,
	)
}
