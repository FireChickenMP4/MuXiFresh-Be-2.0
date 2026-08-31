package logic

import (
	"context"
	"testing"

	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/pb"
	taskmodel "MuXiFresh-Be-2.0/app/task/model"
	usermodel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/globalKey"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestDelSubmissionComment_OwnerPasses(t *testing.T) {
	ownerID := primitive.NewObjectID()
	comment := &taskmodel.Comment{ID: primitive.NewObjectID(), UserId: ownerID}
	svcCtx := &svc.ServiceContext{
		CommentModel: &fakeCommentModel{
			findOneFn: func(ctx context.Context, id string) (*taskmodel.Comment, error) { return comment, nil },
		},
	}
	l := NewDelSubmissionCommentLogic(callerCtx(t, ownerID.Hex()), svcCtx)

	_, err := l.DelSubmissionComment(&pb.DelSubmissionCommentReq{CommentID: comment.ID.Hex()})
	if err != nil {
		t.Fatalf("comment author should be able to delete, got err: %v", err)
	}
}

func TestDelSubmissionComment_AdminPasses(t *testing.T) {
	authorID := primitive.NewObjectID()
	adminID := primitive.NewObjectID()
	comment := &taskmodel.Comment{ID: primitive.NewObjectID(), UserId: authorID}
	svcCtx := &svc.ServiceContext{
		CommentModel: &fakeCommentModel{
			findOneFn: func(ctx context.Context, id string) (*taskmodel.Comment, error) { return comment, nil },
		},
		UserInfoModel: &fakeUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{UserType: globalKey.Admin}, nil
			},
		},
	}
	l := NewDelSubmissionCommentLogic(callerCtx(t, adminID.Hex()), svcCtx)

	_, err := l.DelSubmissionComment(&pb.DelSubmissionCommentReq{CommentID: comment.ID.Hex()})
	if err != nil {
		t.Fatalf("admin should be able to delete any comment, got err: %v", err)
	}
}

func TestDelSubmissionComment_AdminBranchCallerNotFoundNormalized(t *testing.T) {
	// admin 分支：caller 查不到（not-found）→ 归一为"无权删除该评论"，不泄露存在性
	authorID := primitive.NewObjectID()
	comment := &taskmodel.Comment{ID: primitive.NewObjectID(), UserId: authorID}
	svcCtx := &svc.ServiceContext{
		CommentModel: &fakeCommentModel{
			findOneFn: func(ctx context.Context, id string) (*taskmodel.Comment, error) { return comment, nil },
		},
		UserInfoModel: &fakeUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return nil, mon.ErrNotFound
			},
		},
	}
	l := NewDelSubmissionCommentLogic(callerCtx(t, primitive.NewObjectID().Hex()), svcCtx)

	_, err := l.DelSubmissionComment(&pb.DelSubmissionCommentReq{CommentID: comment.ID.Hex()})
	if err == nil || err.Error() != "无权删除该评论" {
		t.Fatalf("caller not-found should normalize to generic rejection, got: %v", err)
	}
}

func TestDelSubmissionComment_OtherUserRejected(t *testing.T) {
	authorID := primitive.NewObjectID()
	otherID := primitive.NewObjectID()
	comment := &taskmodel.Comment{ID: primitive.NewObjectID(), UserId: authorID}
	svcCtx := &svc.ServiceContext{
		CommentModel: &fakeCommentModel{
			findOneFn: func(ctx context.Context, id string) (*taskmodel.Comment, error) { return comment, nil },
		},
		UserInfoModel: &fakeUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{UserType: globalKey.Freshman}, nil
			},
		},
	}
	l := NewDelSubmissionCommentLogic(callerCtx(t, otherID.Hex()), svcCtx)

	_, err := l.DelSubmissionComment(&pb.DelSubmissionCommentReq{CommentID: comment.ID.Hex()})
	if err == nil {
		t.Fatal("unrelated user should not delete others' comments")
	}
}

func TestDelSubmissionComment_NotFoundNormalized(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		CommentModel: &fakeCommentModel{
			findOneFn: func(ctx context.Context, id string) (*taskmodel.Comment, error) {
				return nil, mon.ErrNotFound
			},
		},
	}
	l := NewDelSubmissionCommentLogic(callerCtx(t, primitive.NewObjectID().Hex()), svcCtx)

	_, err := l.DelSubmissionComment(&pb.DelSubmissionCommentReq{CommentID: primitive.NewObjectID().Hex()})
	if err == nil || err.Error() != "评论不存在" {
		t.Fatalf("not-found should normalize, got: %v", err)
	}
}

func TestDelSubmissionComment_NoMetadataRejected(t *testing.T) {
	svcCtx := &svc.ServiceContext{CommentModel: &fakeCommentModel{}}
	l := NewDelSubmissionCommentLogic(context.Background(), svcCtx)

	_, err := l.DelSubmissionComment(&pb.DelSubmissionCommentReq{CommentID: primitive.NewObjectID().Hex()})
	if err == nil {
		t.Fatal("missing caller identity should be rejected")
	}
}
