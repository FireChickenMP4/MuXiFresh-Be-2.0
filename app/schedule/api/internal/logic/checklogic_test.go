package logic

import (
	"context"
	"testing"

	"MuXiFresh-Be-2.0/app/schedule/api/internal/svc"
	"MuXiFresh-Be-2.0/app/schedule/api/internal/types"
	"MuXiFresh-Be-2.0/app/schedule/rpc/scheduleclient"
	usermodel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
)

type fakeScheduleUserInfoModel struct {
	usermodel.UserInfoModel
	findOneFn func(ctx context.Context, id string) (*usermodel.UserInfo, error)
}

func (f *fakeScheduleUserInfoModel) FindOne(ctx context.Context, id string) (*usermodel.UserInfo, error) {
	return f.findOneFn(ctx, id)
}

type fakeScheduleClient struct {
	scheduleclient.ScheduleClient
	checkCalled bool
}

func (f *fakeScheduleClient) Check(ctx context.Context, in *scheduleclient.CheckReq, opts ...grpc.CallOption) (*scheduleclient.CheckResp, error) {
	f.checkCalled = true
	return &scheduleclient.CheckResp{}, nil
}

func scheduleCtxWithUser(uid string) context.Context {
	return context.WithValue(context.Background(), ctxData.CtxKeyJwtUserID, uid)
}

func TestCheck_MyselfWithSchedule(t *testing.T) {
	userID := primitive.NewObjectID()
	scheduleID := primitive.NewObjectID()
	svcCtx := &svc.ServiceContext{
		UserInfoClient: &fakeScheduleUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{ID: userID, ScheduleID: scheduleID}, nil
			},
		},
		ScheduleClient: &fakeScheduleClient{},
	}
	l := NewCheckLogic(scheduleCtxWithUser(userID.Hex()), svcCtx)

	_, err := l.Check(&types.CheckReq{ScheduleID: globalKey.Myself})
	if err != nil {
		t.Fatalf("myself with schedule should pass, got err: %v", err)
	}
	fc := svcCtx.ScheduleClient.(*fakeScheduleClient)
	if !fc.checkCalled {
		t.Fatal("Check should be called for authorized access")
	}
}

func TestCheck_OtherIdRejected(t *testing.T) {
	userID := primitive.NewObjectID()
	scheduleID := primitive.NewObjectID()
	otherID := primitive.NewObjectID()
	svcCtx := &svc.ServiceContext{
		UserInfoClient: &fakeScheduleUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{ID: userID, ScheduleID: scheduleID}, nil
			},
		},
		ScheduleClient: &fakeScheduleClient{},
	}
	l := NewCheckLogic(scheduleCtxWithUser(userID.Hex()), svcCtx)

	_, err := l.Check(&types.CheckReq{ScheduleID: otherID.Hex()})
	if err == nil || err.Error() != "无权查看该进度" {
		t.Fatalf("other schedule id should be rejected, got: %v", err)
	}
	fc := svcCtx.ScheduleClient.(*fakeScheduleClient)
	if fc.checkCalled {
		t.Fatal("Check should not be called for unauthorized access")
	}
}

func TestCheck_MyselfWithoutSchedule(t *testing.T) {
	userID := primitive.NewObjectID()
	svcCtx := &svc.ServiceContext{
		UserInfoClient: &fakeScheduleUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{ID: userID}, nil
			},
		},
		ScheduleClient: &fakeScheduleClient{},
	}
	l := NewCheckLogic(scheduleCtxWithUser(userID.Hex()), svcCtx)

	_, err := l.Check(&types.CheckReq{ScheduleID: globalKey.Myself})
	if err == nil || err.Error() != "尚未创建进度" {
		t.Fatalf("myself without schedule should be rejected, got: %v", err)
	}
}

func TestCheck_EmptyUserId(t *testing.T) {
	svcCtx := &svc.ServiceContext{ScheduleClient: &fakeScheduleClient{}}
	l := NewCheckLogic(scheduleCtxWithUser(""), svcCtx)

	_, err := l.Check(&types.CheckReq{ScheduleID: globalKey.Myself})
	if err == nil || err.Error() != "身份缺失" {
		t.Fatalf("empty user id should be rejected, got: %v", err)
	}
}
