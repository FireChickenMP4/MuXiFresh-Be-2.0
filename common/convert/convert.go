package convert

import "MuXiFresh-Be-2.0/common/globalKey"

func GroupCvtChinese(group string) string {
	switch group {
	case "Backend":
		group = "后端组"
	case "Product":
		group = "产品组"
	case "Design":
		group = "设计组"
	case "Android":
		group = "安卓组"
	case "Frontend":
		group = "前端组"
	case "Operation":
		group = "运营组"
	}
	return group
}

func TypeCvtChinese(Type string) string {
	switch Type {
	case globalKey.Register:
		Type = "注册账户"
	case globalKey.SetPassword:
		Type = "找回密码"
	case globalKey.SetEmail:
		Type = "修改邮箱"
	}
	return Type
}
