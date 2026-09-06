package logic

import (
	"context"
	"errors"

	"MuXiFresh-Be-2.0/app/task/cmd/rpc/comment/internal/svc"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/metadata"
)

// callerIDKey 是 API 层经 grpc metadata 注入调用者身份的键（定义见 common/ctxData.CallerIDKey）。
// 评论入口（Get/Set/Reply/Del/IsMy）统一从 metadata 取身份：
// 调用者身份必须由 API 层从 JWT 上下文注入，RPC 直连不可信——
// 本批防护边界是"API 层可信"，根治依赖 issue-tracker ②-1（N-H1 RPC 鉴权中间件）。
const callerIDKey = ctxData.CallerIDKey

// callerIDFromCtx 从 grpc metadata 提取调用者身份；缺失/为空/非合法 ObjectID 一律拒绝（fail closed）。
// 注意：RPC 直连的客户端可自行设置该 metadata 值冒充任意用户，本函数仅做格式兜底，
// 真正的身份来源保障依赖 API 层注入 + RPC 层统一鉴权中间件（N-H1 根治）。
func callerIDFromCtx(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("缺少用户身份")
	}
	vals := md.Get(callerIDKey)
	if len(vals) == 0 || vals[0] == "" {
		return "", errors.New("缺少用户身份")
	}
	if _, err := primitive.ObjectIDFromHex(vals[0]); err != nil {
		return "", errors.New("非法的用户身份")
	}
	return vals[0], nil
}

// callerUserTypeOf 查询调用者的用户类型；普通用户返回空串，管理员返回 admin/super_admin。
func callerUserTypeOf(ctx context.Context, svcCtx *svc.ServiceContext, callerID string) (string, error) {
	caller, err := svcCtx.UserInfoModel.FindOne(ctx, callerID)
	if err != nil {
		if errors.Is(err, mon.ErrNotFound) {
			return "", errors.New("无权访问该提交")
		}
		return "", err
	}
	if caller == nil {
		return "", errors.New("无权访问该提交")
	}
	if caller.UserType == globalKey.Admin || caller.UserType == globalKey.SuperAdmin {
		return caller.UserType, nil
	}
	return "", nil
}

// checkSubmissionAccess 校验调用者是否有权访问指定 submission：
// 提交者本人或 Admin/SuperAdmin 可读评论/评论/回复，其余一律拒绝。
// 返回调用者的 UserType（普通用户返回空串，管理员返回 admin/super_admin），
// 供上层做管理员判断（如回复是否翻转审阅状态），避免重复查询调用者信息。
// 提交者本人若为管理员，仍保留管理员语义（发根评论/回复可置已审阅）。
// submission/caller 不存在均归一为"无权访问该提交"，不泄露提交存在性。
func checkSubmissionAccess(ctx context.Context, svcCtx *svc.ServiceContext, callerID, submissionID string) (callerUserType string, err error) {
	submission, err := svcCtx.SubmissionModel.FindOne(ctx, submissionID)
	if err != nil {
		if errors.Is(err, mon.ErrNotFound) {
			return "", errors.New("无权访问该提交")
		}
		return "", err
	}
	// 提交者本人（含本人为管理员的情形，保留管理员语义）
	if submission.UserId.Hex() == callerID {
		return callerUserTypeOf(ctx, svcCtx, callerID)
	}
	// 非本人：仅 Admin/SuperAdmin 可访问，普通用户一律拒绝
	ut, err := callerUserTypeOf(ctx, svcCtx, callerID)
	if err != nil {
		return "", err
	}
	if ut == "" {
		return "", errors.New("无权访问该提交")
	}
	return ut, nil
}
