package logic

import (
	"context"
	"errors"
	"time"

	"MuXiFresh-Be-2.0/app/task/model"
	"MuXiFresh-Be-2.0/common/globalKey"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetSubmissionCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetSubmissionCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetSubmissionCommentLogic {
	return &SetSubmissionCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetSubmissionCommentLogic) SetSubmissionComment(in *pb.SetSubmissionCommentReq) (*pb.SetSubmissionCommentResp, error) {

	// 仅 Admin/SuperAdmin 可发根评论（前端学生端无发评论入口，收紧隐藏能力）；
	// 身份经 grpc metadata 注入，防绕过前端直调
	callerID, err := callerIDFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	// 先做输入 hex 校验（归一错误），再做归属/角色校验；
	// 评论作者一律取 metadata callerID（唯一身份来源），不信任请求参数 in.UserId
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
	if callerUserType != globalKey.Admin && callerUserType != globalKey.SuperAdmin {
		return nil, errors.New("无权评论该提交")
	}
	comment := &model.Comment{
		UserId:       userId,
		SubmissionID: submissionId,
		Content:      in.Content,
		UpdateAt:     time.Now(),
		CreateAt:     time.Now(),
	}

	if err = l.svcCtx.CommentModel.Insert(l.ctx, comment); err != nil {
		return nil, err
	}

	// 管理员评论才更新审阅状态（callerUserType 来自归属校验结果，避免重复查询调用者）
	if callerUserType == globalKey.Admin || callerUserType == globalKey.SuperAdmin {
		submission := model.Submission{
			ID:     submissionId,
			Status: globalKey.Reviewed,
		}

		if _, err = l.svcCtx.SubmissionModel.Update(l.ctx, &submission); err != nil {
			return nil, err
		}

	}

	return &pb.SetSubmissionCommentResp{
		Flag: true,
	}, nil
}
