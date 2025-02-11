package mysql

import "errors"

var (
	ErrorUserExist    = errors.New("用户已存在")
	ErrorUserNotExist = errors.New("用户不存在")
	ErrorUserPassword = errors.New("密码错误")
	ErrorInvalidInfo  = errors.New("无效的信息")
)
