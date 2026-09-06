package logic

import (
	"context"
	"testing"

	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/pb"
	taskmodel "MuXiFresh-Be-2.0/app/task/model"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestIsMyComment_OwnerTrue(t *testing.T) {
	ownerID := primitive.NewObjectID()
	comment := &taskmodel.Comment{ID: primitive.NewObjectID(), UserId: ownerID}
	svcCtx := &svc.ServiceContext{
		CommentModel: &fakeCommentModel{
			findOneFn: func(ctx context.Context, id string) (*taskmodel.Comment, error) { return comment, nil },
		},
	}
	l := NewIsMyCommentLogic(callerCtx(t, ownerID.Hex()), svcCtx)

	resp, err := l.IsMyComment(&pb.IsMyCommentReq{CommentID: comment.ID.Hex()})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !resp.Flag {
		t.Fatal("owner should see flag=true")
	}
}

func TestIsMyComment_OtherFalse(t *testing.T) {
	ownerID := primitive.NewObjectID()
	comment := &taskmodel.Comment{ID: primitive.NewObjectID(), UserId: ownerID}
	svcCtx := &svc.ServiceContext{
		CommentModel: &fakeCommentModel{
			findOneFn: func(ctx context.Context, id string) (*taskmodel.Comment, error) { return comment, nil },
		},
	}
	// 身份来自 metadata，不再信任请求参数 in.UserId（可伪造）
	l := NewIsMyCommentLogic(callerCtx(t, primitive.NewObjectID().Hex()), svcCtx)

	resp, err := l.IsMyComment(&pb.IsMyCommentReq{UserId: ownerID.Hex(), CommentID: comment.ID.Hex()})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp.Flag {
		t.Fatal("unrelated user should see flag=false even with forged UserId param")
	}
}

func TestIsMyComment_NoMetadataRejected(t *testing.T) {
	svcCtx := &svc.ServiceContext{CommentModel: &fakeCommentModel{}}
	l := NewIsMyCommentLogic(context.Background(), svcCtx)

	_, err := l.IsMyComment(&pb.IsMyCommentReq{CommentID: primitive.NewObjectID().Hex()})
	if err == nil {
		t.Fatal("missing caller identity should be rejected")
	}
}

func TestIsMyComment_CommentNotFoundNormalized(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		CommentModel: &fakeCommentModel{
			findOneFn: func(ctx context.Context, id string) (*taskmodel.Comment, error) {
				return nil, mon.ErrNotFound
			},
		},
	}
	l := NewIsMyCommentLogic(callerCtx(t, primitive.NewObjectID().Hex()), svcCtx)

	_, err := l.IsMyComment(&pb.IsMyCommentReq{CommentID: primitive.NewObjectID().Hex()})
	if err == nil || err.Error() != "评论不存在" {
		t.Fatalf("not-found should normalize to 评论不存在, got: %v", err)
	}
}
