package comment

import (
	"context"
	"fmt"

	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/commentclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"google.golang.org/grpc/metadata"

	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DelSubmissionCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDelSubmissionCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DelSubmissionCommentLogic {
	return &DelSubmissionCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DelSubmissionCommentLogic) DelSubmissionComment(req *types.DelSubmissionCommentReq) (resp *types.DelSubmissionCommentResp, err error) {
	// 经 grpc metadata 注入调用者身份，供 comment-rpc 做归属校验
	ctx := metadata.AppendToOutgoingContext(l.ctx, ctxData.CallerIDKey, ctxData.GetUserIdFromCtx(l.ctx))
	// 双层鉴权（有意的设计）：
	// API 层经 IsMyComment 预拦——仅评论作者本人可删（管理员也不能在 API 层删他人评论，
	// 前端当前无删除入口，保守不扩大权限；将来管理端需要删评论需另设入口）；
	// RPC 层 Del 再校验"作者本人或 Admin"，admin 分支是防 RPC 直连的纵深防御。
	isMyCmtResp, err := l.svcCtx.CommentClient.IsMyComment(ctx, &commentclient.IsMyCommentReq{
		CommentID: req.CommentID,
	})
	if err != nil {
		return nil, err
	}
	if !isMyCmtResp.Flag {
		return nil, fmt.Errorf("no permission to delete the comment")
	}
	delCommentResp, err := l.svcCtx.CommentClient.DelSubmissionComment(ctx, &commentclient.DelSubmissionCommentReq{
		CommentID: req.CommentID,
	})
	return &types.DelSubmissionCommentResp{
		Flag: delCommentResp.Flag,
	}, nil
}
