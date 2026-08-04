package svc

import (
	"MuXiFresh-Be-2.0/app/schedule/api/internal/config"
	"MuXiFresh-Be-2.0/app/schedule/rpc/scheduleclient"
	userauthModel "MuXiFresh-Be-2.0/app/userauth/model"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	ScheduleClient scheduleclient.ScheduleClient
	UserInfoClient userauthModel.UserInfoModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:         c,
		ScheduleClient: scheduleclient.NewScheduleClient(zrpc.MustNewClient(c.ScheduleConf)),
		UserInfoClient: userauthModel.NewUserInfoModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "userinfo"),
	}
}
