package svc

import (
	"MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/test/rpc/internal/config"
	userauthModel "MuXiFresh-Be-2.0/app/userauth/model"
)

type ServiceContext struct {
	Config          config.Config
	UserInfoClient  userauthModel.UserInfoModel
	EntryFormClient model.EntryFormModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:          c,
		UserInfoClient:  userauthModel.NewUserInfoModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "userinfo"),
		EntryFormClient: model.NewEntryFormModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "entry_form"),
	}
}
