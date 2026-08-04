package svc

import (
	"MuXiFresh-Be-2.0/app/form/api/internal/config"
	"MuXiFresh-Be-2.0/app/form/rpc/entryformclient"
	schedulemodel "MuXiFresh-Be-2.0/app/schedule/model"
	externalModel "MuXiFresh-Be-2.0/app/userauth/model"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config              config.Config
	FormClient          entryformclient.EntryFormClient
	UserInfoModelClient externalModel.UserInfoModel
	ScheduleModel       schedulemodel.ScheduleModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:              c,
		FormClient:          entryformclient.NewEntryFormClient(zrpc.MustNewClient(c.FormConf)),
		UserInfoModelClient: externalModel.NewUserInfoModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "userinfo"),
		ScheduleModel:       schedulemodel.NewScheduleModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "schedule"),
	}
}
