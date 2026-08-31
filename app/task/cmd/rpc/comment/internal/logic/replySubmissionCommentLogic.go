package logic

import (
	"MuXiFresh-Be-2.0/app/task/model"
	"MuXiFresh-Be-2.0/common/globalKey"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/mon"
)

type ReplySubmissionCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReplySubmissionCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReplySubmissionCommentLogic {
	return &ReplySubmissionCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ReplySubmissionCommentLogic) ReplySubmissionComment(in *pb.ReplySubmissionCommentReq) (*pb.ReplySubmissionCommentResp, error) {
	// 归属校验：仅提交者本人或 Admin/SuperAdmin 可回复该提交的评论（身份经 grpc metadata 注入）
	callerID, err := callerIDFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	// 先做输入 hex 校验（归一错误），再做归属/角色校验；
	// 回复作者一律取 metadata callerID（唯一身份来源），不信任请求参数 in.UserId
	userId, err := primitive.ObjectIDFromHex(callerID)
	if err != nil {
		return nil, err
	}
	submissionId, err := primitive.ObjectIDFromHex(in.SubmissionID)
	if err != nil {
		return nil, err
	}
	callerUserType, err := checkSubmissionAccess(l.ctx, l.svcCtx, callerID, in.SubmissionID)
	if err != nil {
		return nil, err
	}

	// 校验 fatherId（必须存在，回复就是基于父评论）
	fatherId, err := primitive.ObjectIDFromHex(in.FatherID)
	if err != nil {
		return nil, err
	}

	// fatherId 归属校验：父评论必须存在且属于同一 submission，防止跨提交伪造回复
	father, err := l.svcCtx.CommentModel.FindOne(l.ctx, in.FatherID)
	if err != nil {
		if errors.Is(err, mon.ErrNotFound) {
			return nil, errors.New("父评论不存在")
		}
		return nil, err
	}
	if father.SubmissionID != submissionId {
		return nil, errors.New("父评论不属于该提交")
	}

	// 构造新的回复评论
	reply := &model.Comment{
		UserId:       userId,
		SubmissionID: submissionId,
		Content:      in.Content,
		FatherId:     fatherId,
		UpdateAt:     time.Now(),
		CreateAt:     time.Now(),
	}

	// 插入数据库
	if err = l.svcCtx.CommentModel.Insert(l.ctx, reply); err != nil {
		return nil, err
	}

	// 仅管理员回复才更新 submission 审阅状态（与 SetSubmissionComment 对齐），
	// 防止普通新生通过回复任意翻转审阅状态（callerUserType 来自归属校验结果，避免重复查询）
	if callerUserType == globalKey.Admin || callerUserType == globalKey.SuperAdmin {
		submission := model.Submission{
			ID:     submissionId,
			Status: globalKey.Reviewed,
		}
		if _, err = l.svcCtx.SubmissionModel.Update(l.ctx, &submission); err != nil {
			return nil, err
		}
	}

	return &pb.ReplySubmissionCommentResp{
		Flag: true,
	}, nil
}
