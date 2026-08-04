package svc

import (
	externalModel2 "MuXiFresh-Be-2.0/app/form/model"
	schedulemodel "MuXiFresh-Be-2.0/app/schedule/model"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/internal/config"
	externalModel "MuXiFresh-Be-2.0/app/userauth/model"
)

type ServiceContext struct {
	Config         config.Config
	UserInfoModel  externalModel.UserInfoModel
	EntryFormModel externalModel2.EntryFormModel
	ScheduleModel  schedulemodel.ScheduleModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:         c,
		UserInfoModel:  externalModel.NewUserInfoModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "userinfo"),
		EntryFormModel: externalModel2.NewEntryFormModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "entry_form"),
		ScheduleModel:  schedulemodel.NewScheduleModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "schedule"),
	}
}
