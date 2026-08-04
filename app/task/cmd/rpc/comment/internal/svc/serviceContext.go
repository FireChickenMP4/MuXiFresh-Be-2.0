package svc

import (
	entryformmodel "MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/internal/config"
	"MuXiFresh-Be-2.0/app/task/model"
	externalModel "MuXiFresh-Be-2.0/app/userauth/model"
)

type ServiceContext struct {
	Config          config.Config
	CommentModel    model.CommentModel
	UserInfoModel   externalModel.UserInfoModel
	SubmissionModel model.SubmissionModel
	EntryFormModel  entryformmodel.EntryFormModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:          c,
		CommentModel:    model.NewCommentModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "comment"),
		UserInfoModel:   externalModel.NewUserInfoModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "userinfo"),
		SubmissionModel: model.NewSubmissionModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "submission"),
		EntryFormModel:  entryformmodel.NewEntryFormModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "entry_form"),
	}
}
