package comment

import (
	"context"

	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/commentclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"google.golang.org/grpc/metadata"

	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetSubmissionCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetSubmissionCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetSubmissionCommentLogic {
	return &SetSubmissionCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetSubmissionCommentLogic) SetSubmissionComment(req *types.SetSubmissionCommentReq) (resp *types.SetSubmissionCommentResp, err error) {
	// 经 grpc metadata 注入调用者身份，供 comment-rpc 做归属校验
	ctx := metadata.AppendToOutgoingContext(l.ctx, ctxData.CallerIDKey, ctxData.GetUserIdFromCtx(l.ctx))
	setCommentResp, err := l.svcCtx.CommentClient.SetSubmissionComment(ctx, &commentclient.SetSubmissionCommentReq{
		SubmissionID: req.SubmissionID,
		Content:      req.Content,
	})
	if err != nil {
		return nil, err
	}
	return &types.SetSubmissionCommentResp{
		Flag: setCommentResp.Flag,
	}, nil
}
