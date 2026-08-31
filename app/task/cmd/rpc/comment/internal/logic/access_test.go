package logic

import (
	"context"
	"errors"
	"testing"

	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/internal/svc"
	taskmodel "MuXiFresh-Be-2.0/app/task/model"
	usermodel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/globalKey"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/metadata"
)

// --- fakes（内嵌接口：只实现被测路径用到的方法，其余走 nil 接口不会被触发） ---

type fakeSubmissionModel struct {
	taskmodel.SubmissionModel
	findOneFn func(ctx context.Context, id string) (*taskmodel.Submission, error)
	updateFn  func(ctx context.Context, data *taskmodel.Submission) (*mongo.UpdateResult, error)
}

func (f *fakeSubmissionModel) FindOne(ctx context.Context, id string) (*taskmodel.Submission, error) {
	return f.findOneFn(ctx, id)
}

func (f *fakeSubmissionModel) Update(ctx context.Context, data *taskmodel.Submission) (*mongo.UpdateResult, error) {
	if f.updateFn != nil {
		return f.updateFn(ctx, data)
	}
	return &mongo.UpdateResult{}, nil
}

type fakeUserInfoModel struct {
	usermodel.UserInfoModel
	findOneFn func(ctx context.Context, id string) (*usermodel.UserInfo, error)
}

func (f *fakeUserInfoModel) FindOne(ctx context.Context, id string) (*usermodel.UserInfo, error) {
	return f.findOneFn(ctx, id)
}

func newTestSvcCtx(sub *taskmodel.Submission, subErr error, caller *usermodel.UserInfo, callerErr error) *svc.ServiceContext {
	return &svc.ServiceContext{
		SubmissionModel: &fakeSubmissionModel{
			findOneFn: func(ctx context.Context, id string) (*taskmodel.Submission, error) {
				return sub, subErr
			},
		},
		UserInfoModel: &fakeUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return caller, callerErr
			},
		},
	}
}

func TestCheckSubmissionAccess_OwnerPasses(t *testing.T) {
	ownerID := primitive.NewObjectID()
	sub := &taskmodel.Submission{ID: ownerID, UserId: ownerID}
	// 本人分支不应查 caller（caller 为 nil 也不 panic）
	svcCtx := newTestSvcCtx(sub, nil, nil, nil)

	ut, err := checkSubmissionAccess(context.Background(), svcCtx, ownerID.Hex(), ownerID.Hex())
	if err != nil {
		t.Fatalf("owner should pass, got err: %v", err)
	}
	if ut != "" {
		t.Fatalf("owner userType should be empty (non-admin), got %q", ut)
	}
}

func TestCheckSubmissionAccess_AdminPasses(t *testing.T) {
	ownerID := primitive.NewObjectID()
	sub := &taskmodel.Submission{ID: ownerID, UserId: ownerID}
	caller := &usermodel.UserInfo{UserType: globalKey.Admin}
	svcCtx := newTestSvcCtx(sub, nil, caller, nil)

	ut, err := checkSubmissionAccess(context.Background(), svcCtx, primitive.NewObjectID().Hex(), ownerID.Hex())
	if err != nil {
		t.Fatalf("admin should pass, got err: %v", err)
	}
	if ut != globalKey.Admin {
		t.Fatalf("admin userType mismatch, got %q", ut)
	}
}

func TestCheckSubmissionAccess_SuperAdminPasses(t *testing.T) {
	ownerID := primitive.NewObjectID()
	sub := &taskmodel.Submission{ID: ownerID, UserId: ownerID}
	caller := &usermodel.UserInfo{UserType: globalKey.SuperAdmin}
	svcCtx := newTestSvcCtx(sub, nil, caller, nil)

	ut, err := checkSubmissionAccess(context.Background(), svcCtx, primitive.NewObjectID().Hex(), ownerID.Hex())
	if err != nil {
		t.Fatalf("super_admin should pass, got err: %v", err)
	}
	if ut != globalKey.SuperAdmin {
		t.Fatalf("super_admin userType mismatch, got %q", ut)
	}
}

func TestCheckSubmissionAccess_OtherUserRejected(t *testing.T) {
	ownerID := primitive.NewObjectID()
	sub := &taskmodel.Submission{ID: ownerID, UserId: ownerID}
	caller := &usermodel.UserInfo{UserType: globalKey.Freshman}
	svcCtx := newTestSvcCtx(sub, nil, caller, nil)

	_, err := checkSubmissionAccess(context.Background(), svcCtx, primitive.NewObjectID().Hex(), ownerID.Hex())
	if err == nil {
		t.Fatal("unrelated user should be rejected")
	}
	if err.Error() != "无权访问该提交" {
		t.Fatalf("expected generic rejection, got: %v", err)
	}
}

func TestCheckSubmissionAccess_SubmissionNotFoundNormalized(t *testing.T) {
	// submission 不存在 → 归一为"无权访问该提交"，不泄露存在性/内部错误串
	svcCtx := newTestSvcCtx(nil, mon.ErrNotFound, nil, nil)

	_, err := checkSubmissionAccess(context.Background(), svcCtx, primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex())
	if err == nil {
		t.Fatal("missing submission should be rejected")
	}
	if err.Error() != "无权访问该提交" {
		t.Fatalf("not-found should normalize to generic error, got: %v", err)
	}
}

func TestCheckSubmissionAccess_CallerNotFoundNormalized(t *testing.T) {
	ownerID := primitive.NewObjectID()
	sub := &taskmodel.Submission{ID: ownerID, UserId: ownerID}
	svcCtx := newTestSvcCtx(sub, nil, nil, mon.ErrNotFound)

	_, err := checkSubmissionAccess(context.Background(), svcCtx, primitive.NewObjectID().Hex(), ownerID.Hex())
	if err == nil {
		t.Fatal("missing caller should be rejected")
	}
	if err.Error() != "无权访问该提交" {
		t.Fatalf("caller not-found should normalize, got: %v", err)
	}
}

func TestCheckSubmissionAccess_DBErrorPropagated(t *testing.T) {
	// 真实 DB 错误原样返回（可观测），不归一
	ownerID := primitive.NewObjectID()
	sub := &taskmodel.Submission{ID: ownerID, UserId: ownerID}
	svcCtx := newTestSvcCtx(sub, nil, nil, errors.New("mongo down"))

	_, err := checkSubmissionAccess(context.Background(), svcCtx, primitive.NewObjectID().Hex(), ownerID.Hex())
	if err == nil || err.Error() != "mongo down" {
		t.Fatalf("db error should propagate, got: %v", err)
	}
}

func TestCheckSubmissionAccess_SubmissionDBErrorPropagated(t *testing.T) {
	// submission 侧真实 DB 错误同样原样返回，不归一
	svcCtx := newTestSvcCtx(nil, errors.New("mongo down"), nil, nil)

	_, err := checkSubmissionAccess(context.Background(), svcCtx, primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex())
	if err == nil || err.Error() != "mongo down" {
		t.Fatalf("submission db error should propagate, got: %v", err)
	}
}

func TestCallerIDFromCtx(t *testing.T) {
	// 无 metadata → 拒绝
	if _, err := callerIDFromCtx(context.Background()); err == nil {
		t.Fatal("missing metadata should be rejected")
	}
	// 空值 → 拒绝
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(callerIDKey, ""))
	if _, err := callerIDFromCtx(ctx); err == nil {
		t.Fatal("empty caller id should be rejected")
	}
	// 正常 → 返回
	ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs(callerIDKey, "abc123"))
	id, err := callerIDFromCtx(ctx)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if id != "abc123" {
		t.Fatalf("expected abc123, got %q", id)
	}
}
