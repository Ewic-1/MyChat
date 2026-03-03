package sms

import (
	"mychat_server/internal/service/redis"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/utils/random"
	"mychat_server/pkg/utils/zlog"
	"strconv"
	"time"
)

func VerifyCode(telephone string) (msg string, ret int) {
	key := "auth_code_" + telephone
	code, err := redis.GetKey(key)
	if err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}

	if code != "" {
		zlog.Info("verification code exists and is not expired")
		return "verification code exists, please retry later", -2
	}

	code = strconv.Itoa(random.GetRandomInt(6))
	zlog.Info(code)

	if err = redis.SetKeyEx(key, code, time.Duration(constants.REDIS_TIMEOUT)*time.Minute); err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1
	}

	return "verification code sent", 0
}
