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

func TestSetSubmissionComment_AdminPasses(t *testing.T) {
	adminID := primitive.NewObjectID()
	sub := &taskmodel.Submission{ID: primitive.NewObjectID(), UserId: primitive.NewObjectID()}
	// 捕获插入的评论，断言作者必须是 metadata callerID（伪造的 in.UserId 不生效）
	var inserted *taskmodel.Comment
	svcCtx := &svc.ServiceContext{
		SubmissionModel: &fakeSubmissionModel{
			findOneFn: func(ctx context.Context, id string) (*taskmodel.Submission, error) { return sub, nil },
		},
		UserInfoModel: &fakeUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{UserType: globalKey.Admin}, nil
			},
		},
		CommentModel: &fakeCommentModel{
			insertFn: func(ctx context.Context, data *taskmodel.Comment) error {
				inserted = data
				return nil
			},
		},
	}
	l := NewSetSubmissionCommentLogic(callerCtx(t, adminID.Hex()), svcCtx)

	_, err := l.SetSubmissionComment(&pb.SetSubmissionCommentReq{
		UserId:       primitive.NewObjectID().Hex(), // 伪造的 in.UserId，不应成为评论作者
		SubmissionID: sub.ID.Hex(),
		Content:      "批注",
	})
	if err != nil {
		t.Fatalf("admin should pass, got err: %v", err)
	}
	if inserted == nil || inserted.UserId.Hex() != adminID.Hex() {
		t.Fatalf("comment author must be metadata callerID (admin), got %+v", inserted)
	}
}

func TestSetSubmissionComment_OwnerRejected(t *testing.T) {
	// 提交者本人（普通学生）也不能发根评论——前端学生端无此入口，收紧隐藏能力
	ownerID := primitive.NewObjectID()
	sub := &taskmodel.Submission{ID: ownerID, UserId: ownerID}
	svcCtx := &svc.ServiceContext{
		SubmissionModel: &fakeSubmissionModel{
			findOneFn: func(ctx context.Context, id string) (*taskmodel.Submission, error) { return sub, nil },
		},
		CommentModel: &fakeCommentModel{},
	}
	l := NewSetSubmissionCommentLogic(callerCtx(t, ownerID.Hex()), svcCtx)

	_, err := l.SetSubmissionComment(&pb.SetSubmissionCommentReq{
		UserId:       ownerID.Hex(),
		SubmissionID: ownerID.Hex(),
		Content:      "x",
	})
	if err == nil {
		t.Fatal("owner (non-admin) should be rejected for root comment")
	}
}

func TestSetSubmissionComment_OtherUserRejected(t *testing.T) {
	// 无关用户被 checkSubmissionAccess 拒绝
	ownerID := primitive.NewObjectID()
	otherID := primitive.NewObjectID()
	sub := &taskmodel.Submission{ID: ownerID, UserId: ownerID}
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
	l := NewSetSubmissionCommentLogic(callerCtx(t, otherID.Hex()), svcCtx)

	_, err := l.SetSubmissionComment(&pb.SetSubmissionCommentReq{
		UserId:       otherID.Hex(),
		SubmissionID: ownerID.Hex(),
		Content:      "x",
	})
	if err == nil {
		t.Fatal("unrelated user should be rejected")
	}
}

func TestSetSubmissionComment_NoMetadataRejected(t *testing.T) {
	svcCtx := &svc.ServiceContext{CommentModel: &fakeCommentModel{}}
	l := NewSetSubmissionCommentLogic(context.Background(), svcCtx)

	_, err := l.SetSubmissionComment(&pb.SetSubmissionCommentReq{
		UserId:       primitive.NewObjectID().Hex(),
		SubmissionID: primitive.NewObjectID().Hex(),
		Content:      "x",
	})
	if err == nil {
		t.Fatal("missing caller identity should be rejected")
	}
}
