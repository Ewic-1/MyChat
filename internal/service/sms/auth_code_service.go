package sms

import (
	"mychat_server/internal/service/myredis"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/utils/random"
	"mychat_server/pkg/utils/zlog"
	"strconv"
	"time"
)

func VerifyCode(telephone string) (msg string, ret int) {
	key := "auth_code_" + telephone
	code, err := myredis.GetKey(key)
	if err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}

	if code != "" {
		zlog.Info("验证码已存在")
		return "验证码已存在", -2
	}

	code = strconv.Itoa(random.GetRandomInt(6))
	zlog.Info(code)

	if err = myredis.SetKeyEx(key, code, time.Duration(constants.REDIS_TIMEOUT)*time.Minute); err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}

	return "已发送验证码", 0
}
