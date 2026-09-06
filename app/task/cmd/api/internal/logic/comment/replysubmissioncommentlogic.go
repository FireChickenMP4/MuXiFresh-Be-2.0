package comment

import (
	"context"

	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/commentclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"google.golang.org/grpc/metadata"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReplySubmissionCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReplySubmissionCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReplySubmissionCommentLogic {
	return &ReplySubmissionCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReplySubmissionCommentLogic) ReplySubmissionComment(req *types.ReplySubmissionCommentReq) (resp *types.ReplySubmissionCommentResp, err error) {
	// 经 grpc metadata 注入调用者身份，供 comment-rpc 做归属校验
	ctx := metadata.AppendToOutgoingContext(l.ctx, ctxData.CallerIDKey, ctxData.GetUserIdFromCtx(l.ctx))
	replyCommentResp, err := l.svcCtx.CommentClient.ReplySubmissionComment(ctx, &commentclient.ReplySubmissionCommentReq{
		SubmissionID: req.SubmissionID,
		FatherID:     req.FatherID,
		Content:      req.Content,
	})
	if err != nil {
		return nil, err
	}
	return &types.ReplySubmissionCommentResp{
		Flag: replyCommentResp.Flag,
	}, nil
}
