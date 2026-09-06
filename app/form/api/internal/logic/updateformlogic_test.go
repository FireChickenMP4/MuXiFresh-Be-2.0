package logic

import (
	"context"
	"testing"

	"MuXiFresh-Be-2.0/app/form/api/internal/svc"
	"MuXiFresh-Be-2.0/app/form/api/internal/types"
	usermodel "MuXiFresh-Be-2.0/app/userauth/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUpdateForm_OwnIdPasses(t *testing.T) {
	userID := primitive.NewObjectID()
	entryFormID := primitive.NewObjectID()
	svcCtx := &svc.ServiceContext{
		UserInfoModelClient: &fakeFormUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{ID: userID, EntryFormID: entryFormID}, nil
			},
		},
		FormClient: &fakeFormClient{},
	}
	l := NewUpdateFormLogic(formCtxWithUser(userID.Hex()), svcCtx)

	_, err := l.UpdateForm(&types.CreateReq{FormId: entryFormID.Hex()})
	if err != nil {
		t.Fatalf("own form id should pass, got err: %v", err)
	}
	fc := svcCtx.FormClient.(*fakeFormClient)
	if !fc.updateCalled {
		t.Fatal("UpdateForm should be called for authorized access")
	}
}

func TestUpdateForm_OtherIdRejected(t *testing.T) {
	userID := primitive.NewObjectID()
	entryFormID := primitive.NewObjectID()
	otherID := primitive.NewObjectID()
	svcCtx := &svc.ServiceContext{
		UserInfoModelClient: &fakeFormUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{ID: userID, EntryFormID: entryFormID}, nil
			},
		},
		FormClient: &fakeFormClient{},
	}
	l := NewUpdateFormLogic(formCtxWithUser(userID.Hex()), svcCtx)

	_, err := l.UpdateForm(&types.CreateReq{FormId: otherID.Hex()})
	if err == nil || err.Error() != "无权修改该报名表" {
		t.Fatalf("other form id should be rejected, got: %v", err)
	}
	fc := svcCtx.FormClient.(*fakeFormClient)
	if fc.updateCalled {
		t.Fatal("UpdateForm should not be called for unauthorized access")
	}
}

func TestUpdateForm_EmptyUserId(t *testing.T) {
	svcCtx := &svc.ServiceContext{FormClient: &fakeFormClient{}}
	l := NewUpdateFormLogic(formCtxWithUser(""), svcCtx)

	_, err := l.UpdateForm(&types.CreateReq{FormId: primitive.NewObjectID().Hex()})
	if err == nil || err.Error() != "身份缺失" {
		t.Fatalf("empty user id should be rejected, got: %v", err)
	}
}
