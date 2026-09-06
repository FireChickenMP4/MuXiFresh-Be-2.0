package ctxData

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	CtxKeyJwtEmail  = "jwtEmail"
	CtxKeyJwtUserID = "jwtUserId"

	// CallerIDKey 是 API 层经 grpc metadata 注入调用者身份的键，
	// 供 rpc 服务做归属/角色校验（如 comment-rpc 的 checkSubmissionAccess）。
	// 跨服务协议键，只在此定义一处，防止 API/RPC 两侧拼写漂移。
	CallerIDKey = "user_id"
)

func GetEmailFromCtx(ctx context.Context) string {
	email, ok := ctx.Value(CtxKeyJwtEmail).(string)
	if !ok {
		logx.WithContext(ctx).Errorf("GetEmailFromCtx failed")
	}
	return email
}

func GetUserIdFromCtx(ctx context.Context) string {
	userID, ok := ctx.Value(CtxKeyJwtUserID).(string)
	if !ok {
		logx.WithContext(ctx).Errorf("GetEmailFromCtx failed")
	}
	return userID
}
