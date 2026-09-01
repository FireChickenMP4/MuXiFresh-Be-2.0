package logic

import (
	"context"
	"testing"

	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/pb"
	taskmodel "MuXiFresh-Be-2.0/app/task/model"
	usermodel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/globalKey"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGetSubmissionComment_OtherUserRejected(t *testing.T) {
	// 非提交者本人、非管理员 → 读取评论被拒（N-H2 越权读的修复）
	ownerID := primitive.NewObjectID()
	otherID := primitive.NewObjectID()
	sub := &taskmodel.Submission{ID: primitive.NewObjectID(), UserId: ownerID}
	svcCtx := &svc.ServiceContext{
		SubmissionModel: &fakeSubmissionModel{
			findOneFn: func(ctx context.Context, id string) (*taskmodel.Submission, error) { return sub, nil },
		},
		UserInfoModel: &fakeUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{UserType: globalKey.Freshman}, nil
			},
		},
		CommentModel: &fakeCommentModel{},
	}
	l := NewGetSubmissionCommentLogic(callerCtx(t, otherID.Hex()), svcCtx)

	_, err := l.GetSubmissionComment(&pb.GetSubmissionCommentReq{SubmissionID: sub.ID.Hex()})
	if err == nil {
		t.Fatal("unrelated user should be rejected")
	}
}

func TestGetSubmissionComment_OwnerPasses(t *testing.T) {
	// 提交者本人 → 可读评论（fake 无评论，返回空列表不报错）
	ownerID := primitive.NewObjectID()
	sub := &taskmodel.Submission{ID: primitive.NewObjectID(), UserId: ownerID}
	svcCtx := &svc.ServiceContext{
		SubmissionModel: &fakeSubmissionModel{
			findOneFn: func(ctx context.Context, id string) (*taskmodel.Submission, error) { return sub, nil },
		},
		UserInfoModel: &fakeUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{}, nil
			},
		},
		CommentModel: &fakeCommentModel{
			findBySubmissionIDFn: func(ctx context.Context, submissionID string) ([]*taskmodel.Comment, error) {
				return []*taskmodel.Comment{}, nil
			},
		},
	}
	l := NewGetSubmissionCommentLogic(callerCtx(t, ownerID.Hex()), svcCtx)

	resp, err := l.GetSubmissionComment(&pb.GetSubmissionCommentReq{SubmissionID: sub.ID.Hex()})
	if err != nil {
		t.Fatalf("owner should read comments, got err: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil resp")
	}
}
