package logic

import (
	"context"
	"errors"
	"testing"

	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/internal/config"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/pb"
	"MuXiFresh-Be-2.0/app/userauth/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type fakeRegisterUserInfoModel struct {
	model.UserInfoModel
	insertFn func(ctx context.Context, data *model.UserInfo) error
	deleteFn func(ctx context.Context, id string) (int64, error)
}

func (f *fakeRegisterUserInfoModel) Insert(ctx context.Context, data *model.UserInfo) error {
	return f.insertFn(ctx, data)
}

func (f *fakeRegisterUserInfoModel) Delete(ctx context.Context, id string) (int64, error) {
	return f.deleteFn(ctx, id)
}

type fakeRegisterUserAuthModel struct {
	model.UserAuthModel
	insertFn func(ctx context.Context, data *model.UserAuth) error
}

func (f *fakeRegisterUserAuthModel) Insert(ctx context.Context, data *model.UserAuth) error {
	return f.insertFn(ctx, data)
}

func TestRegister_RollsBackUserInfoOnUserAuthFail(t *testing.T) {
	// UserAuth 写入失败时应回滚已插入的 userinfo（补偿删除，ID 一致）
	var insertedID, deletedID string
	userInfoClient := &fakeRegisterUserInfoModel{
		insertFn: func(ctx context.Context, data *model.UserInfo) error {
			data.ID = primitive.NewObjectID()
			insertedID = data.ID.Hex()
			return nil
		},
		deleteFn: func(ctx context.Context, id string) (int64, error) {
			deletedID = id
			return 1, nil
		},
	}
	authErr := errors.New("mongo down")
	userAuthClient := &fakeRegisterUserAuthModel{
		insertFn: func(ctx context.Context, data *model.UserAuth) error {
			return authErr
		},
	}
	svcCtx := &svc.ServiceContext{
		Config: config.Config{
			DefaultUserInfo: struct {
				Avatar   string
				NickName string
			}{Avatar: "a", NickName: "n"},
		},
		UserInfoClient: userInfoClient,
		UserAuthClient: userAuthClient,
	}
	l := NewRegisterLogic(context.Background(), svcCtx)

	_, err := l.Register(&pb.RegisterDataReq{Email: "x@x.com", Password: "p"})
	if err != authErr {
		t.Fatalf("expected UserAuth error propagated, got: %v", err)
	}
	if insertedID == "" {
		t.Fatal("userinfo should have been inserted")
	}
	if deletedID != insertedID {
		t.Fatalf("rollback should delete the inserted userinfo, inserted=%q deleted=%q", insertedID, deletedID)
	}
}
