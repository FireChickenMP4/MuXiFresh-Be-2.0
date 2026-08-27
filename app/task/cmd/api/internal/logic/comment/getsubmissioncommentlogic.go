package comment

import (
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/commentclient"
	"context"
	"github.com/jinzhu/copier"

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
	getCommentResp, err := l.svcCtx.CommentClient.GetSubmissionComment(l.ctx, &commentclient.GetSubmissionCommentReq{
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

// buildCommentTree 将评论按 FatherID 挂到根评论下（根评论 FatherID 为 24 个 0）。
// 父评论缺失的回复不挂载，避免 tree 越界 panic。
func buildCommentTree(comments []types.Comment) []types.Comment {
	commentMap := make(map[string]int)
	var index = 0
	var tree []types.Comment
	for _, comment := range comments {
		if comment.FatherID == "000000000000000000000000" {
			tree = append(tree, comment)
			commentMap[comment.CommentID] = index
			index++
		}
	}
	for _, comment := range comments {
		if _, ok := commentMap[comment.CommentID]; !ok {
			if idx, ok := commentMap[comment.FatherID]; ok {
				tree[idx].Replies = append(tree[idx].Replies, comment)
			}
		}
	}
	return tree
}
