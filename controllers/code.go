package controllers

type ResCode int

const (
	CodeSuccess = 1000 + iota
	CodeInvalidParam
	CodeUserExist
	CodeUserNotExist
	CodeInvalidPassword
	CodeServerBusy
	CodeNeedLogin
	CodeInvalidToken
	CodeRateLimit
	CodeOTPExpired
	CodeOTPInvalid
	CodeModifyNil
)

var codeMsgMap = map[int]string{
	CodeSuccess:         "success",
	CodeInvalidParam:    "请求参数错误",
	CodeUserExist:       "用户名已存在",
	CodeUserNotExist:    "用户不存在",
	CodeInvalidPassword: "密码错误",
	CodeServerBusy:      "服务繁忙",
	CodeNeedLogin:       "需要登录",
	CodeInvalidToken:    "无效的token",
	CodeRateLimit:       "限制流量",
	CodeOTPExpired:      "验证码过期",
	CodeOTPInvalid:      "验证码无效",
	CodeModifyNil:       "不允许修改",
}

func (code ResCode) Msg() string {
	msg, ok := codeMsgMap[int(code)]
	if !ok {
		msg = codeMsgMap[CodeServerBusy]
	}
	return msg
}
