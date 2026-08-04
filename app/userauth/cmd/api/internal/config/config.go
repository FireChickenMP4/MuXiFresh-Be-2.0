package config

import (
	"MuXiFresh-Be-2.0/common/infra"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type CaptchaConf struct {
	Height          int
	Width           int
	Length          int
	Maxskew         float64
	Dotcount        int
	ExpireTime      int
	DebugExpireTime int
	TestingKey      string
}

type Config struct {
	rest.RestConf
	Infra             infra.Config
	AccountCenterConf zrpc.RpcClientConf
	JwtAuth           struct {
		AccessSecret string
		AccessExpire int64
	}
	KqConf           kq.KqConf
	KqConsumerConf   kq.KqConf
	CaptchaConf      *CaptchaConf
	EmailCodeExpired int
	JwtAuthChPass    struct {
		AccessSecret string
		AccessExpire int64
	}
}

func (c *Config) ApplyInfra() {
	c.Infra.ApplyEtcd(&c.AccountCenterConf.Etcd)
	c.Infra.ApplyKafka(&c.KqConf, &c.KqConsumerConf)
}
