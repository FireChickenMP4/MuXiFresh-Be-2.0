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

func newReplySvcCtx(sub *taskmodel.Submission, caller *usermodel.UserInfo, father *taskmodel.Comment) *svc.ServiceContext {
	return &svc.ServiceContext{
		SubmissionModel: &fakeSubmissionModel{
			findOneFn: func(ctx context.Context, id string) (*taskmodel.Submission, error) { return sub, nil },
		},
		UserInfoModel: &fakeUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) { return caller, nil },
		},
		CommentModel: &fakeCommentModel{
			findOneFn: func(ctx context.Context, id string) (*taskmodel.Comment, error) { return father, nil },
		},
	}
}

func TestReplySubmissionComment_OwnerPasses(t *testing.T) {
	ownerID := primitive.NewObjectID()
	sub := &taskmodel.Submission{ID: primitive.NewObjectID(), UserId: ownerID}
	father := &taskmodel.Comment{ID: primitive.NewObjectID(), SubmissionID: sub.ID, UserId: primitive.NewObjectID()}
	// 捕获插入的回复，断言作者必须是 metadata callerID（伪造的 in.UserId 不生效）
	var inserted *taskmodel.Comment
	svcCtx := &svc.ServiceContext{
		SubmissionModel: &fakeSubmissionModel{
			findOneFn: func(ctx context.Context, id string) (*taskmodel.Submission, error) { return sub, nil },
		},
		CommentModel: &fakeCommentModel{
			findOneFn: func(ctx context.Context, id string) (*taskmodel.Comment, error) { return father, nil },
			insertFn: func(ctx context.Context, data *taskmodel.Comment) error {
				inserted = data
				return nil
			},
		},
	}
	l := NewReplySubmissionCommentLogic(callerCtx(t, ownerID.Hex()), svcCtx)

	_, err := l.ReplySubmissionComment(&pb.ReplySubmissionCommentReq{
		UserId:       primitive.NewObjectID().Hex(), // 伪造的 in.UserId 不应影响作者身份
		SubmissionID: sub.ID.Hex(),
		FatherID:     father.ID.Hex(),
		Content:      "收到，已修改",
	})
	if err != nil {
		t.Fatalf("owner should be able to reply, got err: %v", err)
	}
	if inserted == nil || inserted.UserId.Hex() != ownerID.Hex() {
		t.Fatalf("reply author must be metadata callerID (owner), got %+v", inserted)
	}
}

func TestReplySubmissionComment_AdminPasses(t *testing.T) {
	ownerID := primitive.NewObjectID()
	adminID := primitive.NewObjectID()
	sub := &taskmodel.Submission{ID: primitive.NewObjectID(), UserId: ownerID}
	father := &taskmodel.Comment{ID: primitive.NewObjectID(), SubmissionID: sub.ID, UserId: primitive.NewObjectID()}
	svcCtx := newReplySvcCtx(sub, &usermodel.UserInfo{UserType: globalKey.Admin}, father)
	l := NewReplySubmissionCommentLogic(callerCtx(t, adminID.Hex()), svcCtx)

	_, err := l.ReplySubmissionComment(&pb.ReplySubmissionCommentReq{
		UserId:       adminID.Hex(),
		SubmissionID: sub.ID.Hex(),
		FatherID:     father.ID.Hex(),
		Content:      "已审阅",
	})
	if err != nil {
		t.Fatalf("admin should be able to reply, got err: %v", err)
	}
}

func TestReplySubmissionComment_OtherUserRejected(t *testing.T) {
	ownerID := primitive.NewObjectID()
	otherID := primitive.NewObjectID()
	sub := &taskmodel.Submission{ID: primitive.NewObjectID(), UserId: ownerID}
	father := &taskmodel.Comment{ID: primitive.NewObjectID(), SubmissionID: sub.ID, UserId: primitive.NewObjectID()}
	svcCtx := newReplySvcCtx(sub, &usermodel.UserInfo{UserType: globalKey.Freshman}, father)
	l := NewReplySubmissionCommentLogic(callerCtx(t, otherID.Hex()), svcCtx)

	_, err := l.ReplySubmissionComment(&pb.ReplySubmissionCommentReq{
		UserId:       otherID.Hex(),
		SubmissionID: sub.ID.Hex(),
		FatherID:     father.ID.Hex(),
		Content:      "x",
	})
	if err == nil {
		t.Fatal("unrelated user should be rejected")
	}
}

func TestReplySubmissionComment_FatherNotFound(t *testing.T) {
	ownerID := primitive.NewObjectID()
	sub := &taskmodel.Submission{ID: primitive.NewObjectID(), UserId: ownerID}
	svcCtx := &svc.ServiceContext{
		SubmissionModel: &fakeSubmissionModel{
			findOneFn: func(ctx context.Context, id string) (*taskmodel.Submission, error) { return sub, nil },
		},
		CommentModel: &fakeCommentModel{
			findOneFn: func(ctx context.Context, id string) (*taskmodel.Comment, error) { return nil, mon.ErrNotFound },
		},
	}
	l := NewReplySubmissionCommentLogic(callerCtx(t, ownerID.Hex()), svcCtx)

	_, err := l.ReplySubmissionComment(&pb.ReplySubmissionCommentReq{
		SubmissionID: sub.ID.Hex(),
		FatherID:     primitive.NewObjectID().Hex(),
		Content:      "x",
	})
	if err == nil || err.Error() != "父评论不存在" {
		t.Fatalf("missing father should normalize, got: %v", err)
	}
}

func TestReplySubmissionComment_FatherCrossSubmissionRejected(t *testing.T) {
	ownerID := primitive.NewObjectID()
	sub := &taskmodel.Submission{ID: primitive.NewObjectID(), UserId: ownerID}
	// father 属于另一个 submission → 跨提交伪造回复被拒
	father := &taskmodel.Comment{ID: primitive.NewObjectID(), SubmissionID: primitive.NewObjectID(), UserId: primitive.NewObjectID()}
	svcCtx := newReplySvcCtx(sub, nil, father)
	l := NewReplySubmissionCommentLogic(callerCtx(t, ownerID.Hex()), svcCtx)

	_, err := l.ReplySubmissionComment(&pb.ReplySubmissionCommentReq{
		SubmissionID: sub.ID.Hex(),
		FatherID:     father.ID.Hex(),
		Content:      "x",
	})
	if err == nil || err.Error() != "父评论不属于该提交" {
		t.Fatalf("cross-submission father should be rejected, got: %v", err)
	}
}

func TestReplySubmissionComment_NoMetadataRejected(t *testing.T) {
	svcCtx := &svc.ServiceContext{CommentModel: &fakeCommentModel{}}
	l := NewReplySubmissionCommentLogic(context.Background(), svcCtx)

	_, err := l.ReplySubmissionComment(&pb.ReplySubmissionCommentReq{
		SubmissionID: primitive.NewObjectID().Hex(),
		FatherID:     primitive.NewObjectID().Hex(),
		Content:      "x",
	})
	if err == nil {
		t.Fatal("missing caller identity should be rejected")
	}
}
