package svc

import (
	externalModel "MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/submission/internal/config"
	"MuXiFresh-Be-2.0/app/task/model"
	userauthModel "MuXiFresh-Be-2.0/app/userauth/model"
)

type ServiceContext struct {
	Config          config.Config
	SubmissionModel model.SubmissionModel
	EntryFormModel  externalModel.EntryFormModel
	UserInfoModel   userauthModel.UserInfoModel
	AssignmentModel model.AssignmentModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:          c,
		SubmissionModel: model.NewSubmissionModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "submission"),
		EntryFormModel:  externalModel.NewEntryFormModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "entry_form"),
		UserInfoModel:   userauthModel.NewUserInfoModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "userinfo"),
		AssignmentModel: model.NewAssignmentModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "assignment"),
	}
}
