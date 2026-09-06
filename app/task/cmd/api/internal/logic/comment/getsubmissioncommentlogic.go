package comment

import (
	"context"

	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/commentclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"github.com/jinzhu/copier"
	"google.golang.org/grpc/metadata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSubmissionCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSubmissionCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSubmissionCommentLogic {
	return &GetSubmissionCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSubmissionCommentLogic) GetSubmissionComment(req *types.GetSubmissionCommentReq) (resp *types.GetSubmissionCommentResp, err error) {
	// 经 grpc metadata 注入调用者身份，供 comment-rpc 做归属校验
	ctx := metadata.AppendToOutgoingContext(l.ctx, ctxData.CallerIDKey, ctxData.GetUserIdFromCtx(l.ctx))
	getCommentResp, err := l.svcCtx.CommentClient.GetSubmissionComment(ctx, &commentclient.GetSubmissionCommentReq{
		SubmissionID: req.SubmissionID,
	})
	if err != nil {
		return nil, err
	}
	var comments []types.Comment
	err = copier.Copy(&comments, &getCommentResp.Comments)
	if err != nil {
		return nil, err
	}
	return &types.GetSubmissionCommentResp{
		Comments: buildCommentTree(comments),
	}, nil
}

// buildCommentTree 将评论按父链递归挂载（根评论 FatherID 为 24 个 0）。
// 父评论缺失的孤儿不挂载，避免 tree 越界 panic。
func buildCommentTree(comments []types.Comment) []types.Comment {
	children := make(map[string][]types.Comment)
	for _, comment := range comments {
		children[comment.FatherID] = append(children[comment.FatherID], comment)
	}
	var build func(parentID string) []types.Comment
	build = func(parentID string) []types.Comment {
		var out []types.Comment
		for _, comment := range children[parentID] {
			comment.Replies = build(comment.CommentID)
			out = append(out, comment)
		}
		return out
	}
	return build("000000000000000000000000")
}
