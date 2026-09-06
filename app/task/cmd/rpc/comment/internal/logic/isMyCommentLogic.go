package logic

import (
	"context"
	"errors"

	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/mon"
)

type IsMyCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIsMyCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsMyCommentLogic {
	return &IsMyCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IsMyCommentLogic) IsMyComment(in *pb.IsMyCommentReq) (*pb.IsMyCommentResp, error) {
	// 身份经 grpc metadata 注入，不再信任请求参数 in.UserId（RPC 直连不可信）
	callerID, err := callerIDFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	comment, err := l.svcCtx.CommentModel.FindOne(l.ctx, in.CommentID)
	if err != nil {
		if errors.Is(err, mon.ErrNotFound) {
			return nil, errors.New("评论不存在")
		}
		return nil, err
	}
	return &pb.IsMyCommentResp{
		Flag: comment.UserId.Hex() == callerID,
	}, nil
}
